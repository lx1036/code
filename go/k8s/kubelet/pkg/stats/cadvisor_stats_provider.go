package stats

import (
	"fmt"

	statsapi "k8s-lx1036/k8s/kubelet/pkg/apis/stats/v1alpha1"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor"
	cadvisorapiv2 "k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/info/v2"
	kubecontainer "k8s-lx1036/k8s/kubelet/pkg/container"
	"k8s-lx1036/k8s/kubelet/pkg/server/stats"
	"k8s-lx1036/k8s/kubelet/pkg/status"

	"k8s.io/klog/v2"
)

// cadvisorStatsProvider implements the containerStatsProvider interface by
// getting the container stats from cAdvisor. This is needed by
// integrations which do not provide stats from CRI. See
// `pkg/kubelet/cadvisor/util.go#UsingLegacyCadvisorStats` for the logic for
// determining which integrations do not provide stats from CRI.
type cadvisorStatsProvider struct {
	// cadvisor is used to get the stats of the cgroup for the containers that
	// are managed by pods.
	cadvisor cadvisor.Interface
	// resourceAnalyzer is used to get the volume stats of the pods.
	resourceAnalyzer stats.ResourceAnalyzer
	// imageService is used to get the stats of the image filesystem.
	imageService kubecontainer.ImageService
	// statusProvider is used to get pod metadata
	statusProvider status.PodStatusProvider
}

func (p *cadvisorStatsProvider) ListPodStatsAndUpdateCPUNanoCoreUsage() ([]statsapi.PodStats, error) {
	panic("implement me")
}

func (p *cadvisorStatsProvider) ListPodCPUAndMemoryStats() ([]statsapi.PodStats, error) {
	panic("implement me")
}

func (p *cadvisorStatsProvider) ImageFsStats() (*statsapi.FsStats, error) {
	panic("implement me")
}

func (p *cadvisorStatsProvider) ImageFsDevice() (string, error) {
	panic("implement me")
}

// newCadvisorStatsProvider returns a containerStatsProvider that provides
// container stats from cAdvisor.
func newCadvisorStatsProvider(
	cadvisor cadvisor.Interface,
	resourceAnalyzer stats.ResourceAnalyzer,
	imageService kubecontainer.ImageService,
	statusProvider status.PodStatusProvider,
) containerStatsProvider {
	return &cadvisorStatsProvider{
		cadvisor:         cadvisor,
		resourceAnalyzer: resourceAnalyzer,
		imageService:     imageService,
		statusProvider:   statusProvider,
	}
}

// ListPodStats returns the stats of all the pod-managed containers.
func (p *cadvisorStatsProvider) ListPodStats() ([]statsapi.PodStats, error) {

	return nil, nil
}

func getCadvisorContainerInfo(ca cadvisor.Interface) (map[string]cadvisorapiv2.ContainerInfo, error) {
	infos, err := ca.ContainerInfoV2("/", cadvisorapiv2.RequestOptions{
		IdType:    cadvisorapiv2.TypeName,
		Count:     2, // 2 samples are needed to compute "instantaneous" CPU
		Recursive: true,
	})
	if err != nil {
		if _, ok := infos["/"]; ok {
			// If the failure is partial, log it and return a best-effort
			// response.
			klog.Errorf("Partial failure issuing cadvisor.ContainerInfoV2: %v", err)
		} else {
			return nil, fmt.Errorf("failed to get root cgroup stats: %v", err)
		}
	}

	return infos, nil
}
