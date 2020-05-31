package node

import (
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/common"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/common/metric"
	corev1 "k8s.io/api/core/v1"
)

type NodeList struct {
	ListMeta common.ListMeta `json:"listMeta"`
	Metrics []metric.Metric `json:"metrics"`

	Nodes []Node `json:"nodes"`
	Errors []error `json:"errors"`
}

type Node struct {
	ObjectMeta common.ObjectMeta `json:"objectMeta"`
	TypeMeta common.TypeMeta `json:"typeMeta"`

	Ready corev1.ConditionStatus `json:"ready"`
	AllocatedResources NodeAllocatedResources `json:"allocatedResources"`
}

type NodeAllocatedResources struct {
	// CPURequests is number of allocated milicores.
	CPURequests int64 `json:"cpuRequests"`
	// CPURequestsFraction is a fraction of CPU, that is allocated.
	CPURequestsFraction float64 `json:"cpuRequestsFraction"`
	// CPULimits is defined CPU limit.
	CPULimits int64 `json:"cpuLimits"`
	// CPULimitsFraction is a fraction of defined CPU limit, can be over 100%, i.e.
	// overcommitted.
	CPULimitsFraction float64 `json:"cpuLimitsFraction"`
	// CPUCapacity is specified node CPU capacity in milicores.
	CPUCapacity int64 `json:"cpuCapacity"`

	// MemoryRequests is a fraction of memory, that is allocated.
	MemoryRequests int64 `json:"memoryRequests"`
	// MemoryRequestsFraction is a fraction of memory, that is allocated.
	MemoryRequestsFraction float64 `json:"memoryRequestsFraction"`
	// MemoryLimits is defined memory limit.
	MemoryLimits int64 `json:"memoryLimits"`
	// MemoryLimitsFraction is a fraction of defined memory limit, can be over 100%, i.e.
	// overcommitted.
	MemoryLimitsFraction float64 `json:"memoryLimitsFraction"`
	// MemoryCapacity is specified node memory capacity in bytes.
	MemoryCapacity int64 `json:"memoryCapacity"`

	// AllocatedPods in number of currently allocated pods on the node.
	AllocatedPods int `json:"allocatedPods"`
	// PodCapacity is maximum number of pods, that can be allocated on the node.
	PodCapacity int64 `json:"podCapacity"`
	// PodFraction is a fraction of pods, that can be allocated on given node.
	PodFraction float64 `json:"podFraction"`
}
