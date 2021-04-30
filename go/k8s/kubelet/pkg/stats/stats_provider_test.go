package stats

import (
	statsapi "k8s-lx1036/k8s/kubelet/pkg/apis/stats/v1alpha1"
	serverstats "k8s-lx1036/k8s/kubelet/pkg/server/stats"

	"k8s.io/apimachinery/pkg/types"
)

type fakeResourceAnalyzer struct {
	podVolumeStats serverstats.PodVolumeStats
}

func (o *fakeResourceAnalyzer) Start()                                           {}
func (o *fakeResourceAnalyzer) Get(bool) (*statsapi.Summary, error)              { return nil, nil }
func (o *fakeResourceAnalyzer) GetCPUAndMemoryStats() (*statsapi.Summary, error) { return nil, nil }
func (o *fakeResourceAnalyzer) GetPodVolumeStats(uid types.UID) (serverstats.PodVolumeStats, bool) {
	return o.podVolumeStats, true
}
