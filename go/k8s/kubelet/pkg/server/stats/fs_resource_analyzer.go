package stats

import (
	"sync"
	"time"
)

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
