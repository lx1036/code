package cadvisor

import (
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/events"
	cadvisorapi "k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/info/v1"
	cadvisorapiv2 "k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/info/v2"
)

// Interface is an abstract interface for testability.  It abstracts the interface to cAdvisor.
type Interface interface {
	Start() error
	DockerContainer(name string, req *cadvisorapi.ContainerInfoRequest) (cadvisorapi.ContainerInfo, error)
	ContainerInfo(name string, req *cadvisorapi.ContainerInfoRequest) (*cadvisorapi.ContainerInfo, error)
	ContainerInfoV2(name string, options cadvisorapiv2.RequestOptions) (map[string]cadvisorapiv2.ContainerInfo, error)
	GetRequestedContainersInfo(containerName string, options cadvisorapiv2.RequestOptions) (map[string]*cadvisorapi.ContainerInfo, error)
	SubcontainerInfo(name string, req *cadvisorapi.ContainerInfoRequest) (map[string]*cadvisorapi.ContainerInfo, error)
	MachineInfo() (*cadvisorapi.MachineInfo, error)

	VersionInfo() (*cadvisorapi.VersionInfo, error)

	// Returns usage information about the filesystem holding container images.
	ImagesFsInfo() (cadvisorapiv2.FsInfo, error)

	// Returns usage information about the root filesystem.
	RootFsInfo() (cadvisorapiv2.FsInfo, error)

	// Get events streamed through passedChannel that fit the request.
	WatchEvents(request *events.Request) (*events.EventChannel, error)

	// Get filesystem information for the filesystem that contains the given file.
	GetDirFsInfo(path string) (cadvisorapiv2.FsInfo, error)
}

// ImageFsInfoProvider informs cAdvisor how to find imagefs for container images.
type ImageFsInfoProvider interface {
	// ImageFsInfoLabel returns the label cAdvisor should use to find the filesystem holding container images.
	ImageFsInfoLabel() (string, error)
}
