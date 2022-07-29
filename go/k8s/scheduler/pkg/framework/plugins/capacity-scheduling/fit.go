package capacity_scheduling

import (
	//framework "k8s-lx1036/k8s/scheduler/pkg/framework"

	"k8s.io/api/core/v1"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	kubefeatures "k8s.io/kubernetes/pkg/features"
)

// PreFilterState computed at PreFilter and used at PostFilter or Reserve.
type PreFilterState struct {
	framework.Resource
}

// Clone the preFilter state.
func (s *PreFilterState) Clone() framework.StateData {
	return s
}

// INFO: @see go/k8s/scheduler/pkg/scheduler/framework/plugins/noderesources/fit.go
// computePodResourceRequest returns a framework.Resource that covers the largest
// width in each resource dimension. Because init-containers run sequentially, we collect
// the max in each dimension iteratively. In contrast, we sum the resource vectors for
// regular containers since they run simultaneously.
//
// If Pod Overhead is specified and the feature gate is set, the resources defined for Overhead
// are added to the calculated Resource request sum
//
// Example:
//
// Pod:
//   InitContainers
//     IC1:
//       CPU: 2
//       Memory: 1G
//     IC2:
//       CPU: 2
//       Memory: 3G
//   Containers
//     C1:
//       CPU: 2
//       Memory: 1G
//     C2:
//       CPU: 1
//       Memory: 1G
//
// Result: CPU: 3, Memory: 3G
func computePodResourceRequest(pod *v1.Pod) *PreFilterState {
	result := &PreFilterState{}
	for _, container := range pod.Spec.Containers {
		result.Add(container.Resources.Requests)
	}

	// take max_resource(sum_pod, any_init_container)
	for _, container := range pod.Spec.InitContainers {
		result.SetMaxResource(container.Resources.Requests)
	}

	// If Overhead is being utilized, add to the total requests for the pod
	if pod.Spec.Overhead != nil && utilfeature.DefaultFeatureGate.Enabled(kubefeatures.PodOverhead) {
		result.Add(pod.Spec.Overhead)
	}

	return result
}
