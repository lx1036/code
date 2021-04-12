package cpumanager

import (
	"fmt"
	"math"
	"sync"
	"time"

	"k8s-lx1036/k8s/kubelet/pkg/cm/containermap"
	"k8s-lx1036/k8s/kubelet/pkg/cm/cpumanager/state"
	"k8s-lx1036/k8s/kubelet/pkg/cm/cpumanager/topology"
	"k8s-lx1036/k8s/kubelet/pkg/cm/cpuset"
	"k8s-lx1036/k8s/kubelet/pkg/cm/topologymanager"

	cadvisorapi "github.com/google/cadvisor/info/v1"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/kubelet/config"
	kubecontainer "k8s.io/kubernetes/pkg/kubelet/container"
	"k8s.io/kubernetes/pkg/kubelet/status"
)

type ActivePodsFunc func() []*v1.Pod

type runtimeService interface {
	UpdateContainerResources(id string, resources *runtimeapi.LinuxContainerResources) error
}

const cpuManagerStateFileName = "cpu_manager_state"

type Manager interface {
	// Start is called during Kubelet initialization.
	Start(activePods ActivePodsFunc, sourcesReady config.SourcesReady,
		podStatusProvider status.PodStatusProvider, containerRuntime runtimeService,
		initialContainers containermap.ContainerMap) error

	// Called to trigger the allocation of CPUs to a container. This must be
	// called at some point prior to the AddContainer() call for a container,
	// e.g. at pod admission time.
	Allocate(pod *v1.Pod, container *v1.Container) error

	// AddContainer is called between container create and container start
	// so that initial CPU affinity settings can be written through to the
	// container runtime before the first process begins to execute.
	AddContainer(p *v1.Pod, c *v1.Container, containerID string) error

	// RemoveContainer is called after Kubelet decides to kill or delete a
	// container. After this call, the CPU manager stops trying to reconcile
	// that container and any CPUs dedicated to the container are freed.
	RemoveContainer(containerID string) error

	// State returns a read-only interface to the internal CPU manager state.
	State() state.Reader

	// GetTopologyHints implements the topologymanager.HintProvider Interface
	// and is consulted to achieve NUMA aware resource alignment among this
	// and other resource controllers.
	GetTopologyHints(*v1.Pod, *v1.Container) map[string][]topologymanager.TopologyHint
}

type manager struct {
	sync.Mutex
	policy Policy

	// reconcilePeriod is the duration between calls to reconcileState.
	reconcilePeriod time.Duration

	// state allows pluggable CPU assignment policies while sharing a common
	// representation of state for the system to inspect and reconcile.
	state state.State

	// containerRuntime is the container runtime service interface needed
	// to make UpdateContainerResources() calls against the containers.
	containerRuntime runtimeService

	// activePods is a method for listing active pods on the node
	// so all the containers can be updated in the reconciliation loop.
	activePods ActivePodsFunc

	// podStatusProvider provides a method for obtaining pod statuses
	// and the containerID of their containers
	podStatusProvider status.PodStatusProvider

	// containerMap provides a mapping from (pod, container) -> containerID
	// for all containers a pod
	containerMap containermap.ContainerMap

	topology *topology.CPUTopology

	nodeAllocatableReservation v1.ResourceList

	// sourcesReady provides the readiness of kubelet configuration sources such as apiserver update readiness.
	// We use it to determine when we can purge inactive pods from checkpointed state.
	sourcesReady config.SourcesReady

	// stateFileDirectory holds the directory where the state file for checkpoints is held.
	stateFileDirectory string
}

func (m *manager) Start(activePods ActivePodsFunc, sourcesReady config.SourcesReady,
	podStatusProvider status.PodStatusProvider, containerRuntime runtimeService,
	initialContainers containermap.ContainerMap) error {
	klog.Infof("[cpumanager] starting with %s policy", m.policy.Name())
	klog.Infof("[cpumanager] reconciling every %v", m.reconcilePeriod)

	m.sourcesReady = sourcesReady
	m.activePods = activePods
	m.podStatusProvider = podStatusProvider
	m.containerRuntime = containerRuntime
	m.containerMap = initialContainers

	stateImpl, err := state.NewCheckpointState(m.stateFileDirectory, cpuManagerStateFileName, m.policy.Name(), m.containerMap)
	if err != nil {
		klog.Errorf("[cpumanager] could not initialize checkpoint manager: %v, please drain node and remove policy state file", err)
		return err
	}
	m.state = stateImpl

	err = m.policy.Start(m.state)
	if err != nil {
		klog.Errorf("[cpumanager] policy start error: %v", err)
		return err
	}

	if m.policy.Name() == string(PolicyNone) {
		return nil
	}

	// 定时任务，sync所有guaranteed pod的cpuset
	go wait.Until(func() { m.reconcileState() }, m.reconcilePeriod, wait.NeverStop)
	return nil
}

type reconciledContainer struct {
	podName       string
	containerName string
	containerID   string
}

func (m *manager) reconcileState() (success []reconciledContainer, failure []reconciledContainer) {
	success = []reconciledContainer{}
	failure = []reconciledContainer{}

	m.removeStaleState()
	for _, pod := range m.activePods() {
		podStatus, ok := m.podStatusProvider.GetPodStatus(pod.UID)
		if !ok {
			klog.Warningf("[cpumanager] reconcileState: skipping pod; status not found (pod: %s)", pod.Name)
			failure = append(failure, reconciledContainer{pod.Name, "", ""})
			continue
		}

		allContainers := append(pod.Spec.InitContainers, pod.Spec.Containers...)
		for _, container := range allContainers {
			containerID, err := findContainerIDByName(&podStatus, container.Name)
			if err != nil {
				klog.Warningf(`[cpumanager] reconcileState: skipping container; ID not found in pod status
					(pod: %s, container: %s, error: %v)`, pod.Name, container.Name, err)
				failure = append(failure, reconciledContainer{pod.Name, container.Name, ""})
				continue
			}

			cstatus, err := findContainerStatusByName(&podStatus, container.Name)
			if err != nil {
				klog.Warningf("[cpumanager] reconcileState: skipping container; container status not found in pod status "+
					"(pod: %s, container: %s, error: %v)", pod.Name, container.Name, err)
				failure = append(failure, reconciledContainer{pod.Name, container.Name, ""})
				continue
			}

			if cstatus.State.Waiting != nil ||
				(cstatus.State.Waiting == nil && cstatus.State.Running == nil && cstatus.State.Terminated == nil) {
				klog.Warningf("[cpumanager] reconcileState: skipping container; container still in the waiting state "+
					"(pod: %s, container: %s, error: %v)", pod.Name, container.Name, err)
				failure = append(failure, reconciledContainer{pod.Name, container.Name, ""})
				continue
			}

			m.Lock()
			if cstatus.State.Terminated != nil {
				_, _, err := m.containerMap.GetContainerRef(containerID)
				if err == nil {
					klog.Warningf("[cpumanager] reconcileState: ignoring terminated container "+
						"(pod: %s, container id: %s)", pod.Name, containerID)
				}
				m.Unlock()
				continue
			}

			m.containerMap.Add(string(pod.UID), container.Name, containerID)
			m.Unlock()

			cpuSet := m.state.GetCPUSetOrDefault(string(pod.UID), container.Name)
			if cpuSet.IsEmpty() {
				// NOTE: This should not happen outside of tests.
				klog.Infof("[cpumanager] reconcileState: skipping container; assigned cpuset is empty "+
					"(pod: %s, container: %s)", pod.Name, container.Name)
				failure = append(failure, reconciledContainer{pod.Name, container.Name, containerID})
				continue
			}

			klog.V(4).Infof("[cpumanager] reconcileState: updating container "+
				"(pod: %s, container: %s, container id: %s, cpuset: %v)", pod.Name, container.Name, containerID, cpuSet)
			err = m.updateContainerCPUSet(containerID, cpuSet)
			if err != nil {
				klog.Errorf("[cpumanager] reconcileState: failed to update container (pod: %s, container: %s, "+
					"container id: %s, cpuset: %v, error: %v)", pod.Name, container.Name, containerID, cpuSet, err)
				failure = append(failure, reconciledContainer{pod.Name, container.Name, containerID})
				continue
			}

			success = append(success, reconciledContainer{pod.Name, container.Name, containerID})
		}
	}

	return success, failure
}

// 通过 cri UpdateContainerResources 接口更新到 cpuset cgroups 中，
// 即linux上文件：/sys/fs/cgroup/cpuset/kubepods/{pod_uid}/{container_id}/cpuset.cpus
func (m *manager) updateContainerCPUSet(containerID string, cpus cpuset.CPUSet) error {
	return m.containerRuntime.UpdateContainerResources(
		containerID,
		&runtimeapi.LinuxContainerResources{
			CpusetCpus: cpus.String(),
		})
}

// 同步 cpu assignment
func (m *manager) removeStaleState() {
	if !m.sourcesReady.AllReady() {
		return
	}

	// We grab the lock to ensure that no new containers will grab CPUs while
	// executing the code below. Without this lock, its possible that we end up
	// removing state that is newly added by an asynchronous call to
	// AddContainer() during the execution of this code.
	m.Lock()
	defer m.Unlock()

	activePods := m.activePods()
	activeContainers := make(map[string]map[string]struct{})
	for _, pod := range activePods {
		activeContainers[string(pod.UID)] = make(map[string]struct{})
		for _, container := range append(pod.Spec.InitContainers, pod.Spec.Containers...) {
			activeContainers[string(pod.UID)][container.Name] = struct{}{}
		}
	}

	assignments := m.state.GetCPUAssignments()
	for podUID := range assignments {
		for containerName := range assignments[podUID] {
			if _, ok := activeContainers[podUID][containerName]; !ok {
				klog.Errorf("[cpumanager] removeStaleState: removing (pod %s, container: %s)",
					podUID, containerName)
				err := m.policyRemoveContainerByRef(podUID, containerName)
				if err != nil {
					klog.Errorf(`[cpumanager] removeStaleState: failed to remove
							(pod %s, container %s), error: %v)`, podUID, containerName, err)
				}
			}
		}
	}
}

func (m *manager) policyRemoveContainerByRef(podUID string, containerName string) error {
	err := m.policy.RemoveContainer(m.state, podUID, containerName)
	if err == nil {
		m.containerMap.RemoveByContainerRef(podUID, containerName)
	}

	return err
}

func (m *manager) Allocate(pod *v1.Pod, container *v1.Container) error {
	panic("implement me")
}

func (m *manager) AddContainer(p *v1.Pod, c *v1.Container, containerID string) error {
	panic("implement me")
}

func (m *manager) RemoveContainer(containerID string) error {
	panic("implement me")
}

func (m *manager) State() state.Reader {
	panic("implement me")
}

func (m *manager) GetTopologyHints(pod *v1.Pod, container *v1.Container) map[string][]topologymanager.TopologyHint {
	panic("implement me")
}

type sourcesReadyStub struct{}

func (s *sourcesReadyStub) AddSource(source string) {}
func (s *sourcesReadyStub) AllReady() bool          { return true }

func NewManager(cpuPolicyName string, reconcilePeriod time.Duration,
	machineInfo *cadvisorapi.MachineInfo, specificCPUs cpuset.CPUSet,
	nodeAllocatableReservation v1.ResourceList, stateFileDirectory string,
	affinity topologymanager.Store) (Manager, error) {
	var topo *topology.CPUTopology
	var policy Policy

	switch policyName(cpuPolicyName) {
	case PolicyStatic:
		var err error
		topo, err = topology.Discover(machineInfo)
		if err != nil {
			return nil, err
		}
		klog.Infof("[cpumanager] detected CPU topology: %v", topo)
		reservedCPUs, ok := nodeAllocatableReservation[v1.ResourceCPU]
		if !ok {
			return nil, fmt.Errorf("[cpumanager] unable to determine reserved CPU resources for static policy")
		}
		if reservedCPUs.IsZero() {
			return nil, fmt.Errorf(`[cpumanager] the static policy requires systemreserved.cpu +
				kubereserved.cpu to be greater than zero`)
		}
		reservedCPUsFloat := float64(reservedCPUs.MilliValue()) / 1000
		numReservedCPUs := int(math.Ceil(reservedCPUsFloat))
		policy, err = NewStaticPolicy(topo, numReservedCPUs, specificCPUs, affinity)
		if err != nil {
			return nil, fmt.Errorf("new static policy error: %v", err)
		}
	default:
		return nil, fmt.Errorf("unknown policy: \"%s\"", cpuPolicyName)
	}

	manager := &manager{
		policy:                     policy,
		reconcilePeriod:            reconcilePeriod,
		topology:                   topo,
		nodeAllocatableReservation: nodeAllocatableReservation,
		stateFileDirectory:         stateFileDirectory,
	}
	manager.sourcesReady = &sourcesReadyStub{}

	return manager, nil
}

func findContainerIDByName(status *v1.PodStatus, name string) (string, error) {
	allStatuses := append(status.InitContainerStatuses, status.ContainerStatuses...)
	for _, containerStatus := range allStatuses {
		if containerStatus.Name == name && containerStatus.ContainerID != "" {
			containerID := &kubecontainer.ContainerID{}
			err := containerID.ParseString(containerStatus.ContainerID)
			if err != nil {
				return "", err
			}
			return containerID.ID, nil
		}
	}
	return "", fmt.Errorf("unable to find ID for container with name %v in pod status (it may not be running)", name)
}

func findContainerStatusByName(status *v1.PodStatus, name string) (*v1.ContainerStatus, error) {
	for _, containerStatus := range append(status.InitContainerStatuses, status.ContainerStatuses...) {
		if containerStatus.Name == name {
			return &containerStatus, nil
		}
	}
	return nil, fmt.Errorf("unable to find status for container with name %v in pod status (it may not be running)", name)
}
