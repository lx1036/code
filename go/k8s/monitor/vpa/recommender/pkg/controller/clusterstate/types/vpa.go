package types

import (
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/apis/autoscaling"
	"time"

	apisv1 "k8s-lx1036/k8s/monitor/vpa/recommender/pkg/apis/autoscaling.k9s.io/v1"
	v1 "k8s-lx1036/k8s/monitor/vpa/recommender/pkg/client/clientset/versioned/typed/autoscaling.k9s.io/v1"
)

// VpaID contains information needed to identify a VPA API object within a cluster.
type VpaID struct {
	Namespace string
	VpaName   string
}

type vpaConditionsMap map[apisv1.VerticalPodAutoscalerConditionType]apisv1.VerticalPodAutoscalerCondition

func (conditionsMap *vpaConditionsMap) Set(
	conditionType apisv1.VerticalPodAutoscalerConditionType,
	status bool, reason string, message string) *vpaConditionsMap {

	return nil
}

// Map from VPA annotation key to value.
type vpaAnnotationsMap map[string]string

// Vpa (Vertical Pod Autoscaler) object is responsible for vertical scaling of
// Pods matching a given label selector.
type Vpa struct {
	ID VpaID
	// Labels selector that determines which Pods are controlled by this VPA
	// object. Can be nil, in which case no Pod is matched.
	PodSelector labels.Selector
	// Map of the object annotations (key-value pairs).
	Annotations vpaAnnotationsMap
	// Map of the status conditions (keys are condition types).
	Conditions vpaConditionsMap
	// Most recently computed recommendation. Can be nil.
	Recommendation *apisv1.RecommendedPodResources
	// All container aggregations that contribute to this VPA.
	// TODO: Garbage collect old AggregateContainerStates.
	aggregateContainerStates aggregateContainerStatesMap
	// Pod Resource Policy provided in the VPA API object. Can be nil.
	ResourcePolicy *apisv1.PodResourcePolicy
	// Initial checkpoints of AggregateContainerStates for containers.
	// The key is container name.
	ContainersInitialAggregateState ContainerNameToAggregateStateMap
	// UpdateMode describes how recommendations will be applied to pods
	UpdateMode *apisv1.UpdateMode
	// Created denotes timestamp of the original VPA object creation
	Created time.Time
	// CheckpointWritten indicates when last checkpoint for the VPA object was stored.
	CheckpointWritten time.Time
	// IsV1Beta1API is set to true if VPA object has labelSelector defined as in v1beta1 api.
	IsV1Beta1API bool
	// TargetRef points to the controller managing the set of pods.
	TargetRef *autoscaling.CrossVersionObjectReference
	// PodCount contains number of live Pods matching a given VPA object.
	PodCount int
}

// UpdateRecommendation updates the recommended resources in the VPA and its
// aggregations with the given recommendation.
func (vpa *Vpa) UpdateRecommendation(recommendation *apisv1.RecommendedPodResources) {
	for _, containerRecommendation := range recommendation.ContainerRecommendations {

		klog.Info(containerRecommendation)
	}

	vpa.Recommendation = recommendation
}

// AggregateStateByContainerName returns a map from container name to the aggregated state
// of all containers with that name, belonging to pods matched by the VPA.
func (vpa *Vpa) AggregateStateByContainerName() ContainerNameToAggregateStateMap {

	return nil
}

// UpdateConditions updates the conditions of VPA objects based on it's state.
// PodsMatched is passed to indicate if there are currently active pods in the
// cluster matching this VPA.
func (vpa *Vpa) UpdateConditions(podsMatched bool) {
	reason := ""
	msg := ""
	if podsMatched {
		delete(vpa.Conditions, apisv1.NoPodsMatched)
	} else {
		reason = "NoPodsMatched"
		msg = "No pods match this VPA object"
		vpa.Conditions.Set(apisv1.NoPodsMatched, true, reason, msg)
	}

	if vpa.HasRecommendation() {
		vpa.Conditions.Set(apisv1.RecommendationProvided, true, "", "")
	} else {
		vpa.Conditions.Set(apisv1.RecommendationProvided, false, reason, msg)
	}
}

// HasRecommendation returns if the VPA object contains any recommendation
func (vpa *Vpa) HasRecommendation() bool {
	return (vpa.Recommendation != nil) && len(vpa.Recommendation.ContainerRecommendations) > 0
}

// SetUpdateMode updates the update mode of the VPA and aggregators under this VPA.
func (vpa *Vpa) SetUpdateMode(updatePolicy *apisv1.PodUpdatePolicy) {

}

// SetResourcePolicy updates the resource policy of the VPA and the scaling
// policies of aggregators under this VPA.
func (vpa *Vpa) SetResourcePolicy(resourcePolicy *apisv1.PodResourcePolicy) {

}

// NewVpa returns a new Vpa with a given ID and pod selector. Doesn't set the
// links to the matched aggregations.
func NewVpa(id VpaID, selector labels.Selector, created time.Time) *Vpa {
	vpa := &Vpa{
		ID:                              id,
		PodSelector:                     selector,
		aggregateContainerStates:        make(aggregateContainerStatesMap),
		ContainersInitialAggregateState: make(ContainerNameToAggregateStateMap),
		Created:                         created,
		Annotations:                     make(vpaAnnotationsMap),
		Conditions:                      make(vpaConditionsMap),
		IsV1Beta1API:                    false,
		PodCount:                        0,
	}
	return vpa
}

// UpdateVpaStatusIfNeeded updates the status field of the VPA API object.
func UpdateVpaStatusIfNeeded(vpaClient v1.VerticalPodAutoscalerInterface, vpaName string, newStatus,
	oldStatus *apisv1.VerticalPodAutoscalerStatus) (result *apisv1.VerticalPodAutoscaler, err error) {

	return nil, err
}

// GetContainerNameToAggregateStateMap returns ContainerNameToAggregateStateMap for pods.
func GetContainerNameToAggregateStateMap(vpa *Vpa) ContainerNameToAggregateStateMap {
	containerNameToAggregateStateMap := vpa.AggregateStateByContainerName()
	filteredContainerNameToAggregateStateMap := make(ContainerNameToAggregateStateMap)

	for containerName, aggregatedContainerState := range containerNameToAggregateStateMap {
		containerResourcePolicy := GetContainerResourcePolicy(containerName, vpa.ResourcePolicy)
		autoscalingDisabled := containerResourcePolicy != nil && containerResourcePolicy.Mode != nil &&
			*containerResourcePolicy.Mode == apisv1.ContainerScalingModeOff
		if !autoscalingDisabled && aggregatedContainerState.TotalSamplesCount > 0 {
			aggregatedContainerState.UpdateFromPolicy(containerResourcePolicy)
			filteredContainerNameToAggregateStateMap[containerName] = aggregatedContainerState
		}
	}

	return filteredContainerNameToAggregateStateMap
}

// GetContainerResourcePolicy returns the ContainerResourcePolicy for a given policy
// and container name. It returns nil if there is no policy specified for the container.
func GetContainerResourcePolicy(containerName string, policy *apisv1.PodResourcePolicy) *apisv1.ContainerResourcePolicy {
	var defaultPolicy *apisv1.ContainerResourcePolicy
	if policy != nil {
		for i, containerPolicy := range policy.ContainerPolicies {
			if containerPolicy.ContainerName == containerName {
				return &policy.ContainerPolicies[i]
			}
			if containerPolicy.ContainerName == apisv1.DefaultContainerResourcePolicy {
				defaultPolicy = &policy.ContainerPolicies[i]
			}
		}
	}
	return defaultPolicy
}
