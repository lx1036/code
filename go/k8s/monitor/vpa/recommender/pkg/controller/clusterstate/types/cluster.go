package types

import (
	"fmt"
	"k8s.io/apimachinery/pkg/labels"
	"time"

	v1 "k8s-lx1036/k8s/monitor/vpa/recommender/pkg/apis/autoscaling.k9s.io/v1"
	corev1 "k8s.io/api/core/v1"
)

// AggregateStateKey determines the set of containers for which the usage samples
// are kept aggregated in the model.
type AggregateStateKey interface {
	Namespace() string
	ContainerName() string
	Labels() labels.Labels
}

// AggregateContainerStatesMap is a map from AggregateStateKey to AggregateContainerState.
type aggregateContainerStatesMap map[AggregateStateKey]*AggregateContainerState

// String representation of the labels.LabelSet. This is the value returned by
// labelSet.String(). As opposed to the LabelSet object, it can be used as a map key.
type labelSetKey string

// Map of label sets keyed by their string representation.
type labelSetMap map[labelSetKey]labels.Set

// ClusterState holds all runtime information about the cluster required for the
// VPA operations, i.e. configuration of resources (pods, containers,
// VPA objects), aggregated utilization of compute resources (CPU, memory) and
// events (container OOMs).
// All input to the VPA Recommender algorithm lives in this structure.
type ClusterState struct {
	// Pods in the cluster.
	Pods map[PodID]*PodState
	// VPA objects in the cluster.
	Vpas map[VpaID]*Vpa
	// VPA objects in the cluster that have no recommendation mapped to the first
	// time we've noticed the recommendation missing or last time we logged
	// a warning about it.
	EmptyVPAs map[VpaID]time.Time
	// Observed VPAs. Used to check if there are updates needed.
	ObservedVpas []*v1.VerticalPodAutoscaler

	// All container aggregations where the usage samples are stored.
	aggregateStateMap aggregateContainerStatesMap
	// Map with all label sets used by the aggregations. It serves as a cache
	// that allows to quickly access labels.Set corresponding to a labelSetKey.
	labelSetMap labelSetMap
}

func NewClusterState() *ClusterState {
	return &ClusterState{
		Pods:              make(map[PodID]*PodState),
		Vpas:              make(map[VpaID]*Vpa),
		EmptyVPAs:         make(map[VpaID]time.Time),
		aggregateStateMap: make(aggregateContainerStatesMap),
		labelSetMap:       make(labelSetMap),
	}
}

// AddSample adds a new usage sample to the proper container in the ClusterState
// object. Requires the container as well as the parent pod to be added to the
// ClusterState first. Otherwise an error is returned.
func (cluster *ClusterState) AddSample(sample *ContainerUsageSampleWithKey) error {
	pod, podExists := cluster.Pods[sample.Container.PodID]
	if !podExists {
		return fmt.Errorf("KeyError: %s", sample.Container.PodID)
	}
	containerState, containerExists := pod.Containers[sample.Container.ContainerName]
	if !containerExists {
		return fmt.Errorf("KeyError: %s", sample.Container.ContainerName)
	}

	if !containerState.AddSample(&sample.ContainerUsageSample) {
		return fmt.Errorf("sample discarded (invalid or out of order)")
	}
	return nil
}

// INFO: 把 vpa 对象缓存到 clusterstate 对象中；如果已经存在但pod selector变化，则更新
func (cluster *ClusterState) AddOrUpdateVpa(verticalPodAutoscaler *v1.VerticalPodAutoscaler, selector labels.Selector) error {
	vpaID := VpaID{Namespace: verticalPodAutoscaler.Namespace, VpaName: verticalPodAutoscaler.Name}
	annotationsMap := verticalPodAutoscaler.Annotations
	conditionsMap := make(vpaConditionsMap)
	for _, condition := range verticalPodAutoscaler.Status.Conditions {
		conditionsMap[condition.Type] = condition
	}
	var currentRecommendation *v1.RecommendedPodResources
	if conditionsMap[v1.RecommendationProvided].Status == corev1.ConditionTrue {
		currentRecommendation = verticalPodAutoscaler.Status.Recommendation
	}

	vpa, vpaExists := cluster.Vpas[vpaID]
	if vpaExists && (vpa.PodSelector.String() != selector.String()) { // 已经存在但pod selector变化

	}
	if !vpaExists { // 不存在则缓存到 clusterstate 对象中
		vpa = NewVpa(vpaID, selector, verticalPodAutoscaler.CreationTimestamp.Time)
		cluster.Vpas[vpaID] = vpa
	}

	// 更新缓存的 vpa 对象
	vpa.TargetRef = verticalPodAutoscaler.Spec.TargetRef
	vpa.Annotations = annotationsMap
	vpa.Conditions = conditionsMap
	vpa.Recommendation = currentRecommendation
	vpa.SetUpdateMode(verticalPodAutoscaler.Spec.UpdatePolicy)
	vpa.SetResourcePolicy(verticalPodAutoscaler.Spec.ResourcePolicy)
	return nil
}

// INFO: 更新或添加缓存对象 cluster.Pods
func (cluster *ClusterState) AddOrUpdatePod(podID PodID, newLabels labels.Set, phase corev1.PodPhase) {
	pod, podExists := cluster.Pods[podID]
	if !podExists {
		pod = newPod(podID)
		cluster.Pods[podID] = pod
	}

	newLabelSetKey := cluster.getLabelSetKey(newLabels)
	if podExists && pod.labelSetKey != newLabelSetKey {
		// This Pod is already counted in the old VPA, remove the link.
		cluster.removePodFromItsVpa(pod)
	}

	if !podExists || pod.labelSetKey != newLabelSetKey {
		pod.labelSetKey = newLabelSetKey

		// Set the links between the containers and aggregations based on the current pod labels.
		for containerName, container := range pod.Containers {
			containerID := ContainerID{PodID: podID, ContainerName: containerName}
			container.aggregator = cluster.findOrCreateAggregateContainerState(containerID)
		}

		cluster.addPodToItsVpa(pod)
	}

	pod.Phase = phase
}

// getLabelSetKey puts the given labelSet in the global labelSet map and returns a
// corresponding labelSetKey.
func (cluster *ClusterState) getLabelSetKey(labelSet labels.Set) labelSetKey {
	l := labelSetKey(labelSet.String())
	cluster.labelSetMap[l] = labelSet

	return l
}

// removePodFromItsVpa decreases the count of Pods associated with a VPA object.
func (cluster *ClusterState) removePodFromItsVpa(pod *PodState) {

}

// findOrCreateAggregateContainerState returns (possibly newly created) AggregateContainerState
// that should be used to aggregate usage samples from container with a given ID.
// The pod with the corresponding PodID must already be present in the ClusterState.
func (cluster *ClusterState) findOrCreateAggregateContainerState(containerID ContainerID) *AggregateContainerState {
	panic("implement me")
}

// addPodToItsVpa increases the count of Pods associated with a VPA object.
// Does a scan similar to findOrCreateAggregateContainerState so could be optimized if needed.
func (cluster *ClusterState) addPodToItsVpa(pod *PodState) {

}
