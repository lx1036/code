package dockershim

import (
	"context"
	"fmt"
	"k8s.io/klog/v2"
	"path/filepath"
	"time"

	"k8s-lx1036/k8s/kubelet/pkg/dockershim/libdocker"

	dockertypes "github.com/docker/docker/api/types"
	dockercontainer "github.com/docker/docker/api/types/container"
	dockerstrslice "github.com/docker/docker/api/types/strslice"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
)

type containerCleanupInfo struct{}

func getContainerTimestamps(r *dockertypes.ContainerJSON) (time.Time, time.Time, time.Time, error) {
	var createdAt, startedAt, finishedAt time.Time
	var err error

	createdAt, err = libdocker.ParseDockerTimestamp(r.Created)
	if err != nil {
		return createdAt, startedAt, finishedAt, err
	}
	startedAt, err = libdocker.ParseDockerTimestamp(r.State.StartedAt)
	if err != nil {
		return createdAt, startedAt, finishedAt, err
	}
	finishedAt, err = libdocker.ParseDockerTimestamp(r.State.FinishedAt)
	if err != nil {
		return createdAt, startedAt, finishedAt, err
	}
	return createdAt, startedAt, finishedAt, nil
}

// CreateContainer INFO: 这里 CreateContainer 是创建普通container，而不是 sandbox container，所以不需要 NetworkingConfig 网络配置，
// 普通container会 nsenter 到 sandbox net namespace 中
// CreateContainer creates a new container in the given PodSandbox
// Docker cannot store the log to an arbitrary location (yet), so we create an
// symlink at LogPath, linking to the actual path of the log.
func (ds *dockerService) CreateContainer(ctx context.Context, request *runtimeapi.CreateContainerRequest) (*runtimeapi.CreateContainerResponse, error) {
	podSandboxID := request.PodSandboxId
	config := request.GetConfig()
	sandboxConfig := request.GetSandboxConfig()
	if config == nil {
		return nil, fmt.Errorf("container config is nil")
	}
	if sandboxConfig == nil {
		return nil, fmt.Errorf("sandbox config is nil for container %q", config.Metadata.Name)
	}

	// INFO: (1)准备CreateContainerConfig，主要是 Config 和 HostConfig 两个属性
	// Config
	labels := makeLabels(config.GetLabels(), config.GetAnnotations())
	// Apply a the container type label.
	labels[containerTypeLabelKey] = containerTypeLabelContainer
	// Write the container log path in the labels.
	labels[containerLogPathLabelKey] = filepath.Join(sandboxConfig.LogDirectory, config.LogPath)
	// Write the sandbox ID in the labels.
	labels[sandboxIDLabelKey] = podSandboxID
	image := ""
	if iSpec := config.GetImage(); iSpec != nil {
		image = iSpec.Image
	}
	containerName := makeContainerName(sandboxConfig, config)
	createConfig := dockertypes.ContainerCreateConfig{
		Name: containerName,
		Config: &dockercontainer.Config{
			// TODO: set User.
			Entrypoint: dockerstrslice.StrSlice(config.Command),
			Cmd:        dockerstrslice.StrSlice(config.Args),
			Env:        generateEnvList(config.GetEnvs()),
			Image:      image,
			WorkingDir: config.WorkingDir,
			Labels:     labels,
			// Interactive containers:
			OpenStdin: config.Stdin,
			StdinOnce: config.StdinOnce,
			Tty:       config.Tty,
			// Disable Docker's health check until we officially support it
			// (https://github.com/kubernetes/kubernetes/issues/25829).
			Healthcheck: &dockercontainer.HealthConfig{
				Test: []string{"NONE"},
			},
		},
		HostConfig: &dockercontainer.HostConfig{
			Binds: generateMountBindings(config.GetMounts()),
			RestartPolicy: dockercontainer.RestartPolicy{
				Name: "no",
			},
		},
	}
	// INFO: HostConfig
	hc := createConfig.HostConfig
	err := ds.updateCreateConfig(&createConfig, config, sandboxConfig, podSandboxID, securityOptSeparator)
	if err != nil {
		return nil, fmt.Errorf("failed to update container create config: %v", err)
	}
	// Set devices for container.
	devices := make([]dockercontainer.DeviceMapping, len(config.Devices))
	for i, device := range config.Devices {
		devices[i] = dockercontainer.DeviceMapping{
			PathOnHost:        device.HostPath,
			PathInContainer:   device.ContainerPath,
			CgroupPermissions: device.Permissions,
		}
	}
	hc.Resources.Devices = devices

	createResp, createErr := ds.client.CreateContainer(createConfig)
	if createErr != nil {
		// INFO: directly return
		return nil, createErr
	}
	if createResp != nil {
		containerID := createResp.ID

		return &runtimeapi.CreateContainerResponse{ContainerId: containerID}, nil
	}

	return nil, createErr
}

func (ds *dockerService) updateCreateConfig(
	createConfig *dockertypes.ContainerCreateConfig,
	config *runtimeapi.ContainerConfig,
	sandboxConfig *runtimeapi.PodSandboxConfig,
	podSandboxID string, securityOptSep rune) error {
	// Apply Linux-specific options if applicable.
	if lc := config.GetLinux(); lc != nil {
		rOpts := lc.GetResources()
		if rOpts != nil {
			createConfig.HostConfig.Resources = dockercontainer.Resources{
				// Memory and MemorySwap are set to the same value, this prevents containers from using any swap.
				Memory:     rOpts.MemoryLimitInBytes,
				MemorySwap: rOpts.MemoryLimitInBytes,
				CPUShares:  rOpts.CpuShares,
				CPUQuota:   rOpts.CpuQuota,
				CPUPeriod:  rOpts.CpuPeriod,
			}
			createConfig.HostConfig.OomScoreAdj = int(rOpts.OomScoreAdj)
		}
		// Note: ShmSize is handled in kube_docker_client.go

		// Apply security context.
	}

	// Apply cgroupsParent derived from the sandbox config.
	if lc := sandboxConfig.GetLinux(); lc != nil {
		// Apply Cgroup options.
		cgroupParent, err := ds.GenerateExpectedCgroupParent(lc.CgroupParent)
		if err != nil {
			return fmt.Errorf("failed to generate cgroup parent in expected syntax for container %q: %v", config.Metadata.Name, err)
		}
		createConfig.HostConfig.CgroupParent = cgroupParent
	}

	return nil
}

// GenerateExpectedCgroupParent returns cgroup parent in syntax expected by cgroup driver
func (ds *dockerService) GenerateExpectedCgroupParent(cgroupParent string) (string, error) {
	return cgroupParent, nil
}

// getContainerLogPath returns the container log path specified by kubelet and the real
// path where docker stores the container log.
func (ds *dockerService) getContainerLogPath(containerID string) (string, string, error) {
	info, err := ds.client.InspectContainer(containerID)
	if err != nil {
		return "", "", fmt.Errorf("failed to inspect container %q: %v", containerID, err)
	}

	// info.Config.Labels[containerLogPathLabelKey]="/var/log/pods/default_cgroup1-75cb7bc8c5-vbzww_cf1f7aa0-acb5-48cf-a2e6-50d34d553d96/cgroup1-0/0.log"
	// info.LogPath="/var/lib/docker/containers/7b7ddc66c3ad06ec236e730d66534cc3c4e6551d5b0dd87d607da0c58491bc8d/7b7ddc66c3ad06ec236e730d66534cc3c4e6551d5b0dd87d607da0c58491bc8d-json.log"
	return info.Config.Labels[containerLogPathLabelKey], info.LogPath, nil
}

// INFO: 这个函数很重要，会把 docker logs path 做个软链接到 /var/log/pods/xxx, 比如:
// /var/log/pods/default_cgroup1-75cb7bc8c5-vbzww_cf1f7aa0-acb5-48cf-a2e6-50d34d553d96/cgroup1-0/0.log ->
// /var/lib/docker/containers/7b7ddc66c3ad06ec236e730d66534cc3c4e6551d5b0dd87d607da0c58491bc8d/7b7ddc66c3ad06ec236e730d66534cc3c4e6551d5b0dd87d607da0c58491bc8d-json.log
func (ds *dockerService) createContainerLogSymlink(containerID string) error {
	path, realPath, err := ds.getContainerLogPath(containerID)
	if err != nil {
		return fmt.Errorf("failed to get container %q log path: %v", containerID, err)
	}

	if path == "" {
		klog.V(5).Infof("Container %s log path isn't specified, will not create the symlink", containerID)
		return nil
	}

	if realPath != "" {
		// Only create the symlink when container log path is specified and log file exists.
		// Delete possibly existing file first
		if err = ds.os.Remove(path); err == nil {
			klog.Warningf("Deleted previously existing symlink file: %q", path)
		}
		if err = ds.os.Symlink(realPath, path); err != nil {
			return fmt.Errorf("failed to create symbolic link %q to the container log file %q for container %q: %v",
				path, realPath, containerID, err)
		}
	} else {
		supported, err := ds.IsCRISupportedLogDriver()
		if err != nil {
			klog.Warningf("Failed to check supported logging driver by CRI: %v", err)
			return nil
		}

		if supported {
			klog.Warningf("Cannot create symbolic link because container log file doesn't exist!")
		} else {
			klog.V(5).Infof("Unsupported logging driver by CRI")
		}
	}

	return nil
}

func (ds *dockerService) StartContainer(ctx context.Context, request *runtimeapi.StartContainerRequest) (*runtimeapi.StartContainerResponse, error) {
	err := ds.client.StartContainer(request.ContainerId)

	// Create container log symlink for all containers (including failed ones).
	if linkError := ds.createContainerLogSymlink(request.ContainerId); linkError != nil {
		return nil, linkError
	}

	if err != nil {
		return nil, fmt.Errorf("failed to start container %q: %v", request.ContainerId, err)
	}

	return &runtimeapi.StartContainerResponse{}, nil
}

func (ds *dockerService) StopContainer(ctx context.Context, request *runtimeapi.StopContainerRequest) (*runtimeapi.StopContainerResponse, error) {
	panic("implement me")
}

func (ds *dockerService) RemoveContainer(ctx context.Context, request *runtimeapi.RemoveContainerRequest) (*runtimeapi.RemoveContainerResponse, error) {
	panic("implement me")
}

func (ds *dockerService) ListContainers(ctx context.Context, request *runtimeapi.ListContainersRequest) (*runtimeapi.ListContainersResponse, error) {
	panic("implement me")
}

func (ds *dockerService) ContainerStatus(ctx context.Context, request *runtimeapi.ContainerStatusRequest) (*runtimeapi.ContainerStatusResponse, error) {
	panic("implement me")
}

func (ds *dockerService) UpdateContainerResources(ctx context.Context, request *runtimeapi.UpdateContainerResourcesRequest) (*runtimeapi.UpdateContainerResourcesResponse, error) {
	panic("implement me")
}

func (ds *dockerService) ContainerStats(ctx context.Context, request *runtimeapi.ContainerStatsRequest) (*runtimeapi.ContainerStatsResponse, error) {
	panic("implement me")
}

func (ds *dockerService) ListContainerStats(ctx context.Context, request *runtimeapi.ListContainerStatsRequest) (*runtimeapi.ListContainerStatsResponse, error) {
	panic("implement me")
}
