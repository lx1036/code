package recommender

import (
	"k8s-lx1036/k8s/monitor/vpa/recommender/pkg/types"
)

// ResourceEstimator is a function from AggregateContainerState to
// model.Resources, e.g. a prediction of resources needed by a group of
// containers.
type ResourceEstimator interface {
	GetResourceEstimation(s *types.AggregateContainerState) types.Resources
}
