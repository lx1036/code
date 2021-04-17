package stats

import (
	statsapi "k8s-lx1036/k8s/kubelet/pkg/apis/stats/v1alpha1"
	"k8s-lx1036/k8s/kubelet/pkg/cm"

	cadvisorapi "github.com/google/cadvisor/info/v1"
	cadvisorv2 "github.com/google/cadvisor/info/v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubernetes/pkg/volume"
)

// Provider hosts methods required by stats handlers.
type Provider interface {
	// The following stats are provided by either CRI or cAdvisor.
	//
	// ListPodStats returns the stats of all the containers managed by pods.
	ListPodStats() ([]statsapi.PodStats, error)
	// ListPodStatsAndUpdateCPUNanoCoreUsage updates the cpu nano core usage for
	// the containers and returns the stats for all the pod-managed containers.
	ListPodCPUAndMemoryStats() ([]statsapi.PodStats, error)
	// ListPodStatsAndUpdateCPUNanoCoreUsage returns the stats of all the
	// containers managed by pods and force update the cpu usageNanoCores.
	// This is a workaround for CRI runtimes that do not integrate with
	// cadvisor. See https://github.com/kubernetes/kubernetes/issues/72788
	// for more details.
	ListPodStatsAndUpdateCPUNanoCoreUsage() ([]statsapi.PodStats, error)
	// ImageFsStats returns the stats of the image filesystem.
	ImageFsStats() (*statsapi.FsStats, error)

	// The following stats are provided by cAdvisor.
	//
	// GetCgroupStats returns the stats and the networking usage of the cgroup
	// with the specified cgroupName.
	GetCgroupStats(cgroupName string, updateStats bool) (*statsapi.ContainerStats, *statsapi.NetworkStats, error)
	// GetCgroupCPUAndMemoryStats returns the CPU and memory stats of the cgroup with the specified cgroupName.
	GetCgroupCPUAndMemoryStats(cgroupName string, updateStats bool) (*statsapi.ContainerStats, error)

	// RootFsStats returns the stats of the node root filesystem.
	RootFsStats() (*statsapi.FsStats, error)

	// The following stats are provided by cAdvisor for legacy usage.
	//
	// GetContainerInfo returns the information of the container with the
	// containerName managed by the pod with the uid.
	GetContainerInfo(podFullName string, uid types.UID, containerName string, req *cadvisorapi.ContainerInfoRequest) (*cadvisorapi.ContainerInfo, error)
	// GetRawContainerInfo returns the information of the container with the
	// containerName. If subcontainers is true, this function will return the
	// information of all the sub-containers as well.
	GetRawContainerInfo(containerName string, req *cadvisorapi.ContainerInfoRequest, subcontainers bool) (map[string]*cadvisorapi.ContainerInfo, error)
	// GetRequestedContainersInfo returns the information of the container with
	// the containerName, and with the specified cAdvisor options.
	GetRequestedContainersInfo(containerName string, options cadvisorv2.RequestOptions) (map[string]*cadvisorapi.ContainerInfo, error)

	// The following information is provided by Kubelet.
	//
	// GetPodByName returns the spec of the pod with the name in the specified
	// namespace.
	GetPodByName(namespace, name string) (*v1.Pod, bool)
	// GetNode returns the spec of the local node.
	GetNode() (*v1.Node, error)
	// GetNodeConfig returns the configuration of the local node.
	GetNodeConfig() cm.NodeConfig
	// ListVolumesForPod returns the stats of the volume used by the pod with
	// the podUID.
	ListVolumesForPod(podUID types.UID) (map[string]volume.Volume, bool)
	// GetPods returns the specs of all the pods running on this node.
	GetPods() []*v1.Pod

	// RlimitStats returns the rlimit stats of system.
	RlimitStats() (*statsapi.RlimitStats, error)

	// GetPodCgroupRoot returns the literal cgroupfs value for the cgroup containing all pods
	GetPodCgroupRoot() string

	// GetPodByCgroupfs provides the pod that maps to the specified cgroup literal, as well
	// as whether the pod was found.
	GetPodByCgroupfs(cgroupfs string) (*v1.Pod, bool)
}
