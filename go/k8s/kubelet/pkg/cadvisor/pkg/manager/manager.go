package manager

import (
	"net/http"
	"time"

	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/cache/memory"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/container"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/events"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/info/v1"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/info/v2"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/utils/sysfs"
)

// The Manager interface defines operations for starting a manager and getting
// container and machine information.
type Manager interface {
	// Start the manager. Calling other manager methods before this returns
	// may produce undefined behavior.
	Start() error

	// Stops the manager.
	Stop() error

	//  information about a container.
	GetContainerInfo(containerName string, query *v1.ContainerInfoRequest) (*v1.ContainerInfo, error)

	// Get V2 information about a container.
	// Recursive (subcontainer) requests are best-effort, and may return a partial result alongside an
	// error in the partial failure case.
	GetContainerInfoV2(containerName string, options v2.RequestOptions) (map[string]v2.ContainerInfo, error)

	// Get information about all subcontainers of the specified container (includes self).
	SubcontainersInfo(containerName string, query *v1.ContainerInfoRequest) ([]*v1.ContainerInfo, error)

	// Gets all the Docker containers. Return is a map from full container name to ContainerInfo.
	AllDockerContainers(query *v1.ContainerInfoRequest) (map[string]v1.ContainerInfo, error)

	// Gets information about a specific Docker container. The specified name is within the Docker namespace.
	DockerContainer(dockerName string, query *v1.ContainerInfoRequest) (v1.ContainerInfo, error)

	// Gets spec for all containers based on request options.
	GetContainerSpec(containerName string, options v2.RequestOptions) (map[string]v2.ContainerSpec, error)

	// Gets summary stats for all containers based on request options.
	//GetDerivedStats(containerName string, options v2.RequestOptions) (map[string]v2.DerivedStats, error)

	// Get info for all requested containers based on the request options.
	GetRequestedContainersInfo(containerName string, options v2.RequestOptions) (map[string]*v1.ContainerInfo, error)

	// Returns true if the named container exists.
	Exists(containerName string) bool

	// Get information about the machine.
	GetMachineInfo() (*v1.MachineInfo, error)

	// Get version information about different components we depend on.
	GetVersionInfo() (*v1.VersionInfo, error)

	// GetFsInfoByFsUUID returns the information of the device having the
	// specified filesystem uuid. If no such device with the UUID exists, this
	// function will return the fs.ErrNoSuchDevice error.
	GetFsInfoByFsUUID(uuid string) (v2.FsInfo, error)

	// Get filesystem information for the filesystem that contains the given directory
	GetDirFsInfo(dir string) (v2.FsInfo, error)

	// Get filesystem information for a given label.
	// Returns information for all global filesystems if label is empty.
	GetFsInfo(label string) ([]v2.FsInfo, error)

	// Get ps output for a container.
	GetProcessList(containerName string, options v2.RequestOptions) ([]v2.ProcessInfo, error)

	// Get events streamed through passedChannel that fit the request.
	WatchForEvents(request *events.Request) (*events.EventChannel, error)

	// Get past events that have been detected and that fit the request.
	GetPastEvents(request *events.Request) ([]*v1.Event, error)

	CloseEventChannel(watchID int)

	// Get status information about docker.
	DockerInfo() (v1.DockerStatus, error)

	// Get details about interesting docker images.
	DockerImages() ([]v1.DockerImage, error)

	// Returns debugging information. Map of lines per category.
	DebugInfo() map[string][]string
}

// Housekeeping configuration for the manager
type HouskeepingConfig = struct {
	Interval     *time.Duration
	AllowDynamic *bool
}

// New takes a memory storage and returns a new manager.
func New(memoryCache *memory.InMemoryCache, sysfs sysfs.SysFs, houskeepingConfig HouskeepingConfig,
	includedMetricsSet container.MetricSet, collectorHTTPClient *http.Client,
	rawContainerCgroupPathPrefixWhiteList []string, perfEventsFile string) (Manager, error) {

}
