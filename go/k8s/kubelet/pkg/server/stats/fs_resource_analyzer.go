package stats

import (
	"sync"
	"sync/atomic"
	"time"

	"k8s.io/apimachinery/pkg/types"
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

// newFsResourceAnalyzer returns a new fsResourceAnalyzer implementation
func newFsResourceAnalyzer(statsProvider Provider, calcVolumePeriod time.Duration) *fsResourceAnalyzer {
	r := &fsResourceAnalyzer{
		statsProvider: statsProvider,
		calcPeriod:    calcVolumePeriod,
	}
	r.cachedVolumeStats.Store(make(statCache))
	return r
}
