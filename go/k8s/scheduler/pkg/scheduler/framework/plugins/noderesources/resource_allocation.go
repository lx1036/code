package noderesources

import (
	framework "k8s-lx1036/k8s/scheduler/pkg/scheduler/framework"
	schedutil "k8s-lx1036/k8s/scheduler/pkg/scheduler/util"

	v1 "k8s.io/api/core/v1"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/klog/v2"
	v1helper "k8s.io/kubernetes/pkg/apis/core/v1/helper"
	"k8s.io/kubernetes/pkg/features"
)

// resourceToValueMap contains resource name and score.
type resourceToValueMap map[v1.ResourceName]int64

// resourceToWeightMap contains resource name and weight.
type resourceToWeightMap map[v1.ResourceName]int64

// resourceAllocationScorer contains information to calculate resource allocation score.
type resourceAllocationScorer struct {
	Name                string
	scorer              func(requested, allocatable resourceToValueMap, includeVolumes bool, requestedVolumes int, allocatableVolumes int) int64
	resourceToWeightMap resourceToWeightMap
}

func (r *resourceAllocationScorer) score(
	pod *v1.Pod,
	nodeInfo *framework.NodeInfo) (int64, *framework.Status) {
	node := nodeInfo.Node()
	if node == nil {
		return 0, framework.NewStatus(framework.Error, "node not found")
	}
	if r.resourceToWeightMap == nil {
		return 0, framework.NewStatus(framework.Error, "resources not found")
	}
	// INFO: 这里requested值，得是nodeInfo.NonZeroRequested + pod_request
	requested := make(resourceToValueMap, len(r.resourceToWeightMap))
	allocatable := make(resourceToValueMap, len(r.resourceToWeightMap))
	for resource := range r.resourceToWeightMap {
		allocatable[resource], requested[resource] = calculateResourceAllocatableRequest(nodeInfo, pod, resource)
	}
	var score int64

	// INFO: 打分不考虑 volumes 因子
	score = r.scorer(requested, allocatable, false, 0, 0)

	klog.Infof(
		"%v -> %v: %v, map of allocatable resources %v, map of requested resources %v ,score %d,",
		pod.Name, node.Name, r.Name,
		allocatable, requested, score,
	)

	return score, nil
}

func calculateResourceAllocatableRequest(nodeInfo *framework.NodeInfo, pod *v1.Pod, resource v1.ResourceName) (int64, int64) {
	podRequest := calculatePodResourceRequest(pod, resource)
	switch resource {
	case v1.ResourceCPU:
		return nodeInfo.Allocatable.MilliCPU, (nodeInfo.NonZeroRequested.MilliCPU + podRequest)
	case v1.ResourceMemory:
		return nodeInfo.Allocatable.Memory, (nodeInfo.NonZeroRequested.Memory + podRequest)

	case v1.ResourceEphemeralStorage:
		return nodeInfo.Allocatable.EphemeralStorage, (nodeInfo.Requested.EphemeralStorage + podRequest)
	default:
		if v1helper.IsScalarResourceName(resource) {
			return nodeInfo.Allocatable.ScalarResources[resource], (nodeInfo.Requested.ScalarResources[resource] + podRequest)
		}
	}

	klog.Infof("requested resource %v not considered for node score calculation", resource)
	return 0, 0
}

// podResoureRequest = max(sum(podSpec.Containers), podSpec.InitContainers) + overHead
// INFO: 计算 pod request 时，如果 request.* 没有值则使用默认值
func calculatePodResourceRequest(pod *v1.Pod, resource v1.ResourceName) int64 {
	var podRequest int64
	for i := range pod.Spec.Containers {
		container := &pod.Spec.Containers[i]
		value := schedutil.GetNonzeroRequestForResource(resource, &container.Resources.Requests)
		podRequest += value
	}

	for i := range pod.Spec.InitContainers {
		initContainer := &pod.Spec.InitContainers[i]
		value := schedutil.GetNonzeroRequestForResource(resource, &initContainer.Resources.Requests)
		if podRequest < value {
			podRequest = value
		}
	}

	// If Overhead is being utilized, add to the total requests for the pod
	if pod.Spec.Overhead != nil && utilfeature.DefaultFeatureGate.Enabled(features.PodOverhead) {
		if quantity, found := pod.Spec.Overhead[resource]; found {
			podRequest += quantity.Value()
		}
	}

	return podRequest
}
