package cgroup

import (
	"time"

	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
	"k8s.io/kubernetes/pkg/kubelet/cm/cpuset"
	"k8s.io/kubernetes/pkg/kubelet/cri/remote"
)

type runtimeService interface {
	UpdateContainerResources(id string, resources *runtimeapi.LinuxContainerResources) error
}

type Manager struct {
	containerRuntime runtimeService
}

// @see https://github.com/kubernetes/kubernetes/blob/release-1.19/pkg/kubelet/cm/cpumanager/cpu_manager.go#L456-L466
func (manager *Manager) UpdateContainerResource(containerID string, cpus cpuset.CPUSet) error {
	err := manager.containerRuntime.UpdateContainerResources(containerID,
		&runtimeapi.LinuxContainerResources{
			CpusetCpus: cpus.String(),
		})
	if err != nil {
		return err
	}

	return nil
}

func NewManager(remoteRuntimeEndpoint string, connectionTimeout time.Duration) (*Manager, error) {
	remoteRuntimeService, err := remote.NewRemoteRuntimeService(remoteRuntimeEndpoint, connectionTimeout)
	if err != nil {
		return nil, err
	}

	return &Manager{
		containerRuntime: remoteRuntimeService,
	}, nil
}
