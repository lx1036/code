package stats

import (
	"time"
)

// ResourceAnalyzer provides statistics on node resource consumption
type ResourceAnalyzer interface {
	Start()

	fsResourceAnalyzerInterface
	SummaryProvider
}

// resourceAnalyzer implements ResourceAnalyzer
type resourceAnalyzer struct {
	*fsResourceAnalyzer
	SummaryProvider
}

// Start starts background functions necessary for the ResourceAnalyzer to function
func (ra *resourceAnalyzer) Start() {
	ra.fsResourceAnalyzer.Start()
}

// NewResourceAnalyzer returns a new ResourceAnalyzer
func NewResourceAnalyzer(statsProvider Provider, calVolumeFrequency time.Duration) ResourceAnalyzer {
	fsAnalyzer := newFsResourceAnalyzer(statsProvider, calVolumeFrequency)
	summaryProvider := NewSummaryProvider(statsProvider)
	return &resourceAnalyzer{fsAnalyzer, summaryProvider}
}
