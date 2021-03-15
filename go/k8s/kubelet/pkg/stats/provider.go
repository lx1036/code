package stats

import (
	statsapi "k8s-lx1036/k8s/kubelet/pkg/apis/v1alpha1"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor"
	"k8s-lx1036/k8s/kubelet/pkg/server/stats"
	kubecontainer "k8s.io/kubernetes/pkg/kubelet/container"
	kubepod "k8s.io/kubernetes/pkg/kubelet/pod"
)

// Provider provides the stats of the node and the pod-managed containers.
type Provider struct {
	cadvisor     cadvisor.Interface
	podManager   kubepod.Manager
	runtimeCache kubecontainer.RuntimeCache
	containerStatsProvider
	rlimitStatsProvider
}

// containerStatsProvider is an interface that provides the stats of the
// containers managed by pods.
type containerStatsProvider interface {
	ListPodStats() ([]statsapi.PodStats, error)
	ListPodStatsAndUpdateCPUNanoCoreUsage() ([]statsapi.PodStats, error)
	ListPodCPUAndMemoryStats() ([]statsapi.PodStats, error)
	ImageFsStats() (*statsapi.FsStats, error)
	ImageFsDevice() (string, error)
}

type rlimitStatsProvider interface {
	RlimitStats() (*statsapi.RlimitStats, error)
}

// NewCRIStatsProvider returns a Provider that provides the node stats
// from cAdvisor and the container stats from CRI.
func NewCRIStatsProvider(
	cadvisor cadvisor.Interface,
	resourceAnalyzer stats.ResourceAnalyzer,
	podManager kubepod.Manager,
	runtimeCache kubecontainer.RuntimeCache,
	runtimeService internalapi.RuntimeService,
	imageService internalapi.ImageManagerService,
	logMetricsService LogMetricsService,
	osInterface kubecontainer.OSInterface,
) *Provider {
	return newStatsProvider(cadvisor, podManager, runtimeCache, newCRIStatsProvider(cadvisor, resourceAnalyzer,
		runtimeService, imageService, logMetricsService, osInterface))
}

// newStatsProvider returns a new Provider that provides node stats from
// cAdvisor and the container stats using the containerStatsProvider.
func newStatsProvider(
	cadvisor cadvisor.Interface,
	podManager kubepod.Manager,
	runtimeCache kubecontainer.RuntimeCache,
	containerStatsProvider containerStatsProvider,
) *Provider {
	return &Provider{
		cadvisor:               cadvisor,
		podManager:             podManager,
		runtimeCache:           runtimeCache,
		containerStatsProvider: containerStatsProvider,
	}
}
