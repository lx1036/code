package stats

import (
	"sync"
	"sync/atomic"
	"time"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

type statCache map[types.UID]*volumeStatCalculator

// fsResourceAnalyzerInterface is for embedding fs functions into ResourceAnalyzer
type fsResourceAnalyzerInterface interface {
	GetPodVolumeStats(uid types.UID) (PodVolumeStats, bool)
}

// fsResourceAnalyzer provides stats about fs resource usage
type fsResourceAnalyzer struct {
	statsProvider     Provider
	calcPeriod        time.Duration
	cachedVolumeStats atomic.Value
	startOnce         sync.Once
}

// Start eager background caching of volume stats.
func (s *fsResourceAnalyzer) Start() {
	s.startOnce.Do(func() {
		if s.calcPeriod <= 0 {
			klog.Info("Volume stats collection disabled.")
			return
		}
		klog.Info("Starting FS ResourceAnalyzer")
		go wait.Forever(func() { s.updateCachedPodVolumeStats() }, s.calcPeriod)
	})
}

// GetPodVolumeStats returns the PodVolumeStats for a given pod.  Results are looked up from a cache that
// is eagerly populated in the background, and never calculated on the fly.
func (s *fsResourceAnalyzer) GetPodVolumeStats(uid types.UID) (PodVolumeStats, bool) {
	cache := s.cachedVolumeStats.Load().(statCache)
	statCalc, found := cache[uid]
	if !found {
		// TODO: Differentiate between stats being empty
		// See issue #20679
		return PodVolumeStats{}, false
	}
	return statCalc.GetLatest()
}

// newFsResourceAnalyzer returns a new fsResourceAnalyzer implementation
func newFsResourceAnalyzer(statsProvider Provider, calcVolumePeriod time.Duration) *fsResourceAnalyzer {
	r := &fsResourceAnalyzer{
		statsProvider: statsProvider,
		calcPeriod:    calcVolumePeriod,
	}
	r.cachedVolumeStats.Store(make(statCache))
	return r
}
