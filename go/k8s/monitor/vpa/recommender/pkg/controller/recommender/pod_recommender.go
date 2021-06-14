package recommender

import (
	"k8s-lx1036/k8s/monitor/vpa/recommender/pkg/types"
)

type PodResourceRecommender struct {
	targetEstimator     ResourceEstimator
	lowerBoundEstimator ResourceEstimator
	upperBoundEstimator ResourceEstimator
}

// RecommendedPodResources is a Map from container name to recommended resources.
type RecommendedPodResources map[string]RecommendedContainerResources

// RecommendedContainerResources is the recommendation of resources for a
// container.
type RecommendedContainerResources struct {
	// Recommended optimal amount of resources.
	Target types.Resources
	// Recommended minimum amount of resources.
	LowerBound types.Resources
	// Recommended maximum amount of resources.
	UpperBound types.Resources
}

func (r *PodResourceRecommender) GetRecommendedPodResources(containerNameToAggregateStateMap types.ContainerNameToAggregateStateMap) RecommendedPodResources {

}
