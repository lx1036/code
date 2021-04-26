package container

import (
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/info/v1"
)

// ListType describes whether listing should be just for a
// specific container or performed recursively.
type ListType int

const (
	ListSelf ListType = iota
	ListRecursive
)

type ContainerType int

const (
	ContainerTypeRaw ContainerType = iota
	ContainerTypeDocker
	ContainerTypeCrio
	ContainerTypeContainerd
	ContainerTypeMesos
)

// Interface for container operation handlers.
type ContainerHandler interface {
	// Returns the ContainerReference
	ContainerReference() (v1.ContainerReference, error)

	// Returns container's isolation spec.
	GetSpec() (v1.ContainerSpec, error)

	// Returns the current stats values of the container.
	GetStats() (*v1.ContainerStats, error)

	// Returns the subcontainers of this container.
	ListContainers(listType ListType) ([]v1.ContainerReference, error)

	// Returns the processes inside this container.
	ListProcesses(listType ListType) ([]int, error)

	// Returns absolute cgroup path for the requested resource.
	GetCgroupPath(resource string) (string, error)

	// Returns container labels, if available.
	GetContainerLabels() map[string]string

	// Returns the container's ip address, if available
	GetContainerIPAddress() string

	// Returns whether the container still exists.
	Exists() bool

	// Cleanup frees up any resources being held like fds or go routines, etc.
	Cleanup()

	// Start starts any necessary background goroutines - must be cleaned up in Cleanup().
	// It is expected that most implementations will be a no-op.
	Start()

	// Type of handler
	Type() ContainerType
}
