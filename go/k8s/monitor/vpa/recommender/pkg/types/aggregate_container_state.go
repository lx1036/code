package types

import (
	corev1 "k8s.io/api/core/v1"
	"time"

	v1 "k8s-lx1036/k8s/monitor/vpa/recommender/pkg/apis/autoscaling.k9s.io/v1"
)

// ContainerNameToAggregateStateMap maps a container name to AggregateContainerState
// that aggregates state of containers with that name.
type ContainerNameToAggregateStateMap map[string]*AggregateContainerState

// AggregateContainerState holds input signals aggregated from a set of containers.
// It can be used as an input to compute the recommendation.
// The CPU and memory distributions use decaying histograms by default
// (see NewAggregateContainerState()).
// Implements ContainerStateAggregator interface.
type AggregateContainerState struct {
	// AggregateCPUUsage is a distribution of all CPU samples.
	AggregateCPUUsage util.Histogram
	// AggregateMemoryPeaks is a distribution of memory peaks from all containers:
	// each container should add one peak per memory aggregation interval (e.g. once every 24h).
	AggregateMemoryPeaks util.Histogram
	// Note: first/last sample timestamps as well as the sample count are based only on CPU samples.
	FirstSampleStart  time.Time
	LastSampleStart   time.Time
	TotalSamplesCount int
	CreationTime      time.Time

	// Following fields are needed to correctly report quality metrics
	// for VPA. When we record a new sample in an AggregateContainerState
	// we want to know if it needs recommendation, if the recommendation
	// is present and if the automatic updates are on (are we able to
	// apply the recommendation to the pods).
	LastRecommendation  corev1.ResourceList
	IsUnderVPA          bool
	UpdateMode          *v1.UpdateMode
	ScalingMode         *v1.ContainerScalingMode
	ControlledResources *[]ResourceName
}
