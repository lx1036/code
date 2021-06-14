package clusterstate

import (
	"fmt"
	"k8s.io/apimachinery/pkg/labels"
	"time"

	v1 "k8s-lx1036/k8s/monitor/vpa/recommender/pkg/apis/autoscaling.k9s.io/v1"
	corev1 "k8s.io/api/core/v1"
)

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

	newlabelSetKey := cluster.getLabelSetKey(newLabels)
	if podExists && pod.labelSetKey != newlabelSetKey {
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
