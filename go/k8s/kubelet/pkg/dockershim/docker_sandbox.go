package dockershim

import (
	"context"
	"fmt"

	"k8s-lx1036/k8s/kubelet/pkg/checkpointmanager/errors"
	kubecontainer "k8s-lx1036/k8s/kubelet/pkg/container"

	dockertypes "github.com/docker/docker/api/types"
	dockerfilters "github.com/docker/docker/api/types/filters"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
	"k8s.io/klog/v2"
)

const (
	// String used to detect docker host mode for various namespaces (e.g.
	// networking). Must match the value returned by docker inspect -f
	// '{{.HostConfig.NetworkMode}}'.
	namespaceModeHost = "host"

	// Name of the underlying container runtime
	runtimeName = "docker"
)

// getIPsFromPlugin interrogates the network plugin for sandbox IPs.
func (ds *dockerService) getIPsFromPlugin(sandbox *dockertypes.ContainerJSON) ([]string, error) {
	metadata, err := parseSandboxName(sandbox.Name)
	if err != nil {
		return nil, err
	}

	msg := fmt.Sprintf("Couldn't find network status for %s/%s through plugin", metadata.Namespace, metadata.Name)
	cID := kubecontainer.BuildContainerID(runtimeName, sandbox.ID)
	// INFO: 调用 network 模块获取 pod network status
	networkStatus, err := ds.network.GetPodNetworkStatus(metadata.Namespace, metadata.Name, cID)
	if err != nil {
		return nil, err
	}
	if networkStatus == nil {
		return nil, fmt.Errorf("%v: invalid network status for", msg)
	}

	ips := make([]string, 0)
	for _, ip := range networkStatus.IPs {
		ips = append(ips, ip.String())
	}
	// if we don't have any ip in our list then cni is using classic primary IP only
	if len(ips) == 0 {
		ips = append(ips, networkStatus.IP.String())
	}

	return ips, nil
}

// Returns whether the sandbox network is ready, and whether the sandbox is known
func (ds *dockerService) getNetworkReady(podSandboxID string) (bool, bool) {
	ds.networkReadyLock.Lock()
	defer ds.networkReadyLock.Unlock()
	ready, ok := ds.networkReady[podSandboxID]
	return ready, ok
}

// getIPs returns the ip given the output of `docker inspect` on a pod sandbox,
// first interrogating any registered plugins, then simply trusting the ip
// in the sandbox itself. We look for an ipv4 address before ipv6.
func (ds *dockerService) getIPs(podSandboxID string, sandbox *dockertypes.ContainerJSON) []string {
	if sandbox.NetworkSettings == nil {
		return nil
	}
	if networkNamespaceMode(sandbox) == runtimeapi.NamespaceMode_NODE {
		// For sandboxes using host network, the shim is not responsible for
		// reporting the IP.
		return nil
	}

	// Don't bother getting IP if the pod is known and networking isn't ready
	ready, ok := ds.getNetworkReady(podSandboxID)
	if ok && !ready {
		return nil
	}

	// INFO: 运行 `nsenter --net=/proc/${pid}/ns/net -F -- ip -o -4 addr show dev eth0 scope global` 获得 pod ip
	ips, err := ds.getIPsFromPlugin(sandbox)
	if err != nil {
		klog.Info(fmt.Sprintf("get ips from plugin err: %v", err))
		return nil
	}

	return ips
}

// Returns the inspect container response, the sandbox metadata, and network namespace mode
func (ds *dockerService) getPodSandboxDetails(podSandboxID string) (*dockertypes.ContainerJSON, *runtimeapi.PodSandboxMetadata, error) {
	resp, err := ds.client.InspectContainer(podSandboxID)
	if err != nil {
		return nil, nil, err
	}

	metadata, err := parseSandboxName(resp.Name)
	if err != nil {
		return nil, nil, err
	}

	return resp, metadata, nil
}

// PodSandboxStatus pod ip 在 sandbox status 中
func (ds *dockerService) PodSandboxStatus(ctx context.Context, request *runtimeapi.PodSandboxStatusRequest) (*runtimeapi.PodSandboxStatusResponse, error) {
	podSandboxID := request.PodSandboxId

	r, metadata, err := ds.getPodSandboxDetails(podSandboxID)
	if err != nil {
		return nil, err
	}

	// Parse the timestamps.
	createdAt, _, _, err := getContainerTimestamps(r)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamp for container %q: %v", podSandboxID, err)
	}
	ct := createdAt.UnixNano()

	// Translate container to sandbox state.
	state := runtimeapi.PodSandboxState_SANDBOX_NOTREADY
	if r.State.Running {
		state = runtimeapi.PodSandboxState_SANDBOX_READY
	}

	// INFO: 获取 pod ip
	var ips []string
	ips = ds.getIPs(podSandboxID, r)
	// ip is primary ips
	// ips is all other ips
	ip := ""
	if len(ips) != 0 {
		ip = ips[0]
		ips = ips[1:]
	}

	labels, annotations := extractLabels(r.Config.Labels)
	status := &runtimeapi.PodSandboxStatus{
		Id:          r.ID,
		State:       state,
		CreatedAt:   ct,
		Metadata:    metadata,
		Labels:      labels,
		Annotations: annotations,
		Network: &runtimeapi.PodSandboxNetworkStatus{
			Ip: ip,
		},
		Linux: &runtimeapi.LinuxPodSandboxStatus{
			Namespaces: &runtimeapi.Namespace{
				Options: &runtimeapi.NamespaceOption{
					Network: networkNamespaceMode(r),
					Pid:     pidNamespaceMode(r),
					Ipc:     ipcNamespaceMode(r),
				},
			},
		},
	}
	// add additional IPs
	additionalPodIPs := make([]*runtimeapi.PodIP, 0, len(ips))
	for _, ip := range ips {
		additionalPodIPs = append(additionalPodIPs, &runtimeapi.PodIP{
			Ip: ip,
		})
	}
	status.Network.AdditionalIps = additionalPodIPs

	return &runtimeapi.PodSandboxStatusResponse{Status: status}, nil
}

func (ds *dockerService) ListPodSandbox(ctx context.Context, request *runtimeapi.ListPodSandboxRequest) (*runtimeapi.ListPodSandboxResponse, error) {
	// INFO: filter
	filter := request.GetFilter()
	// By default, list all containers whether they are running or not.
	opts := dockertypes.ContainerListOptions{All: true}
	filterOutReadySandboxes := false
	opts.Filters = dockerfilters.NewArgs()
	f := newDockerFilter(&opts.Filters)
	// Add filter to select only sandbox containers.
	f.AddLabel(containerTypeLabelKey, containerTypeLabelSandbox)
	if filter != nil {
		if filter.Id != "" {
			f.Add("id", filter.Id)
		}
		if filter.State != nil {
			if filter.GetState().State == runtimeapi.PodSandboxState_SANDBOX_READY {
				// Only list running containers.
				opts.All = false
			} else {
				// runtimeapi.PodSandboxState_SANDBOX_NOTREADY can mean the
				// container is in any of the non-running state (e.g., created,
				// exited). We can't tell docker to filter out running
				// containers directly, so we'll need to filter them out
				// ourselves after getting the results.
				filterOutReadySandboxes = true
			}
		}
		if filter.LabelSelector != nil {
			for k, v := range filter.LabelSelector {
				f.AddLabel(k, v)
			}
		}
	}

	// Make sure we get the list of checkpoints first so that we don't include
	// new PodSandboxes that are being created right now.
	var err error
	checkpoints := []string{}
	if filter == nil {
		checkpoints, err = ds.checkpointManager.ListCheckpoints()
		if err != nil {
			klog.Errorf("Failed to list checkpoints: %v", err)
		}
	}

	containers, err := ds.client.ListContainers(opts)
	if err != nil {
		return nil, err
	}

	// Convert docker containers to runtime api sandboxes.
	result := []*runtimeapi.PodSandbox{}
	// using map as set
	sandboxIDs := make(map[string]bool)
	for i := range containers {
		c := containers[i]
		converted, err := containerToRuntimeAPISandbox(&c)
		if err != nil {
			klog.V(4).Infof("Unable to convert docker to runtime API sandbox %+v: %v", c, err)
			continue
		}
		if filterOutReadySandboxes && converted.State == runtimeapi.PodSandboxState_SANDBOX_READY {
			continue
		}
		sandboxIDs[converted.Id] = true
		result = append(result, converted)
	}

	// Include sandbox that could only be found with its checkpoint if no filter is applied
	// These PodSandbox will only include PodSandboxID, Name, Namespace.
	// These PodSandbox will be in PodSandboxState_SANDBOX_NOTREADY state.
	for _, id := range checkpoints {
		if _, ok := sandboxIDs[id]; ok {
			continue
		}
		checkpoint := NewPodSandboxCheckpoint("", "", &CheckpointData{})
		err := ds.checkpointManager.GetCheckpoint(id, checkpoint)
		if err != nil {
			klog.Errorf("Failed to retrieve checkpoint for sandbox %q: %v", id, err)
			if err == errors.ErrCorruptCheckpoint {
				err = ds.checkpointManager.RemoveCheckpoint(id)
				if err != nil {
					klog.Errorf("Failed to delete corrupt checkpoint for sandbox %q: %v", id, err)
				}
			}
			continue
		}
		result = append(result, checkpointToRuntimeAPISandbox(id, checkpoint))
	}

	return &runtimeapi.ListPodSandboxResponse{Items: result}, nil
}

// networkNamespaceMode returns the network runtimeapi.NamespaceMode for this container.
// Supports: POD, NODE
func networkNamespaceMode(container *dockertypes.ContainerJSON) runtimeapi.NamespaceMode {
	if container != nil && container.HostConfig != nil && string(container.HostConfig.NetworkMode) == namespaceModeHost {
		return runtimeapi.NamespaceMode_NODE
	}

	return runtimeapi.NamespaceMode_POD
}

// pidNamespaceMode returns the PID runtimeapi.NamespaceMode for this container.
// Supports: CONTAINER, NODE
func pidNamespaceMode(container *dockertypes.ContainerJSON) runtimeapi.NamespaceMode {
	if container != nil && container.HostConfig != nil && string(container.HostConfig.PidMode) == namespaceModeHost {
		return runtimeapi.NamespaceMode_NODE
	}

	return runtimeapi.NamespaceMode_CONTAINER
}

// ipcNamespaceMode returns the IPC runtimeapi.NamespaceMode for this container.
// Supports: POD, NODE
func ipcNamespaceMode(container *dockertypes.ContainerJSON) runtimeapi.NamespaceMode {
	if container != nil && container.HostConfig != nil && string(container.HostConfig.IpcMode) == namespaceModeHost {
		return runtimeapi.NamespaceMode_NODE
	}

	return runtimeapi.NamespaceMode_POD
}
