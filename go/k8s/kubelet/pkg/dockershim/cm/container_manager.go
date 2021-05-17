package cm

import (
	"k8s-lx1036/k8s/kubelet/pkg/dockershim/libdocker"
	"k8s-lx1036/k8s/kubelet/runc/libcontainer/cgroups"
)

// ContainerManager is an interface that abstracts the basic operations of a
// container manager.
type ContainerManager interface {
	Start() error
}

type containerManager struct {
	// Docker client.
	client libdocker.Interface
	// Name of the cgroups.
	cgroupsName string
	// Manager for the cgroups.
	cgroupsManager cgroups.Manager
}

func (c *containerManager) Start() error {
	return nil
}

// NewContainerManager creates a new instance of ContainerManager
func NewContainerManager(cgroupsName string, client libdocker.Interface) ContainerManager {
	return &containerManager{
		cgroupsName: cgroupsName,
		client:      client,
	}
}
