package util

import (
	v1 "k8s.io/api/core/v1"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	v1helper "k8s.io/kubernetes/pkg/apis/core/v1/helper"
	"k8s.io/kubernetes/pkg/features"
)

// INFO: 这个字段会被 NodeResourcesLeastAllocated plugin 使用，和 Requested 字段意思类似，但是如果 pod request 没有设置值，也需要根据一个默认值去
// 统计，所以 (allocatable - NonZeroRequested[cpu]) / allocatable 就表示 sum(request) 占该 node allocatable 资源比率，哪个 node 比率最小分数最高

const (
	// DefaultMilliCPURequest defines default milli cpu request number.
	DefaultMilliCPURequest int64 = 100 // 0.1 core
	// DefaultMemoryRequest defines default memory request size.
	DefaultMemoryRequest int64 = 200 * 1024 * 1024 // 200 MB
)

// GetNonzeroRequests returns the default cpu and memory resource request if none is found or
// what is provided on the request.
func GetNonzeroRequests(requests *v1.ResourceList) (int64, int64) {
	return GetNonzeroRequestForResource(v1.ResourceCPU, requests),
		GetNonzeroRequestForResource(v1.ResourceMemory, requests)
}

// GetNonzeroRequestForResource returns the default resource request if none is found or
// what is provided on the request.
func GetNonzeroRequestForResource(resource v1.ResourceName, requests *v1.ResourceList) int64 {
	switch resource {
	case v1.ResourceCPU:
		// Override if un-set, but not if explicitly set to zero
		if _, found := (*requests)[v1.ResourceCPU]; !found {
			return DefaultMilliCPURequest
		}
		return requests.Cpu().MilliValue()
	case v1.ResourceMemory:
		// Override if un-set, but not if explicitly set to zero
		if _, found := (*requests)[v1.ResourceMemory]; !found {
			return DefaultMemoryRequest
		}
		return requests.Memory().Value()
	case v1.ResourceEphemeralStorage:
		// if the local storage capacity isolation feature gate is disabled, pods request 0 disk.
		if !utilfeature.DefaultFeatureGate.Enabled(features.LocalStorageCapacityIsolation) {
			return 0
		}

		quantity, found := (*requests)[v1.ResourceEphemeralStorage]
		if !found {
			return 0
		}
		return quantity.Value()
	default:
		if v1helper.IsScalarResourceName(resource) {
			quantity, found := (*requests)[resource]
			if !found {
				return 0
			}
			return quantity.Value()
		}
	}
	return 0
}
