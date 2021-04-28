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

func (manager *Manager) UpdateContainerResource(containerID string, cpus cpuset.CPUSet) error {
	//cpus := cpuset.NewCPUSet(1, 13)
	//containerID := "0e8b25a584ce27c6c88a59d9411cafc6ac82bd90ee67ccaead109ffbccd46cf4"
	err := manager.containerRuntime.UpdateContainerResources(containerID,
		&runtimeapi.LinuxContainerResources{
			CpusetCpus: cpus.String(),
		})
	if err != nil {
		return err
	}

	return nil
}

func NewManager(remoteRuntimeEndpoint string, connectionTimeout time.Duration) *Manager {
	//remoteRuntimeEndpoint := "unix:///var/run/dockershim.sock"
	//remoteRuntimeService, err := remote.NewRemoteRuntimeService(remoteRuntimeEndpoint, time.Minute*2)
	remoteRuntimeService, err := remote.NewRemoteRuntimeService(remoteRuntimeEndpoint, connectionTimeout)
	if err != nil {
		panic(err)
	}

	return &Manager{
		containerRuntime: remoteRuntimeService,
	}
}
