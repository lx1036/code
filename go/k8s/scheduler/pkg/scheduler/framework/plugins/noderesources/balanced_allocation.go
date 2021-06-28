package noderesources

import (
	"context"
	"fmt"
	"math"

	framework "k8s-lx1036/k8s/scheduler/pkg/scheduler/framework"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// INFO: NodeResourcesBalancedAllocation 主要计算 abs(cpu_ratio - memory_ratio)，值越大就表示不平衡，0 表示最平衡。所以，
// 值越小分数越大，值越大分数越小

// BalancedAllocation is a score plugin that calculates the difference between the cpu and memory fraction
// of capacity, and prioritizes the host based on how close the two metrics are to each other.
type BalancedAllocation struct {
	handle framework.FrameworkHandle
	resourceAllocationScorer
}

const BalancedAllocationName = "NodeResourcesBalancedAllocation"

// defaultRequestedRatioResources is used to set default requestToWeight map for CPU and memory
var defaultRequestedRatioResources = resourceToWeightMap{v1.ResourceMemory: 1, v1.ResourceCPU: 1}

// Name returns name of the plugin. It is used in logs, etc.
func (ba *BalancedAllocation) Name() string {
	return BalancedAllocationName
}

func (ba *BalancedAllocation) ScoreExtensions() framework.ScoreExtensions {
	return nil
}

func (ba *BalancedAllocation) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod,
	nodeName string) (int64, *framework.Status) {
	nodeInfo, err := ba.handle.SnapshotSharedLister().NodeInfos().Get(nodeName)
	if err != nil {
		return 0, framework.NewStatus(framework.Error, fmt.Sprintf("getting node %q from Snapshot: %v", nodeName, err))
	}

	// 来源于算法：**[An Energy Efficient Virtual Machine Placement Algorithm with Balanced Resource Utilization](https://ieeexplore.ieee.org/document/6603690)**
	// 主要来源于需求：计算资源分配越平衡越好。只是，算法为何是 cpu_ratio 和 memory_ratio 计算差值？
	return ba.score(pod, nodeInfo)
}

func NewBalancedAllocation(_ runtime.Object, h framework.FrameworkHandle) (framework.Plugin, error) {
	return &BalancedAllocation{
		handle: h,
		resourceAllocationScorer: resourceAllocationScorer{
			Name:                BalancedAllocationName,
			scorer:              balancedResourceScorer,
			resourceToWeightMap: defaultRequestedRatioResources, // TODO: resource weight 还没使用
		},
	}, nil
}

// todo: use resource weights in the scorer function
// TODO: 这里可以抄一抄 NodeResourcesLeastAllocated plugin
func balancedResourceScorer(requested, allocatable resourceToValueMap, includeVolumes bool, requestedVolumes int, allocatableVolumes int) int64 {
	cpuFraction := fractionOfCapacity(requested[v1.ResourceCPU], allocatable[v1.ResourceCPU])
	memoryFraction := fractionOfCapacity(requested[v1.ResourceMemory], allocatable[v1.ResourceMemory])
	// This to find a node which has most balanced CPU, memory and volume usage.
	if cpuFraction >= 1 || memoryFraction >= 1 {
		// if requested >= capacity, the corresponding host should never be preferred.
		return 0
	}

	// INFO: cpu_ratio 和 memory_ratio 计算差值。算法来自于论文，虽然不知道为啥算法这样写，不过不重要，使用就行。
	diff := math.Abs(cpuFraction - memoryFraction)
	return int64((1 - diff) * float64(framework.MaxNodeScore))
}

// 分数占比 requested / capacity
func fractionOfCapacity(requested, capacity int64) float64 {
	if capacity == 0 {
		return 1
	}
	return float64(requested) / float64(capacity)
}
