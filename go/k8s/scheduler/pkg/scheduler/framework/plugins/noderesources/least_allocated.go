package noderesources

import (
	"context"
	"fmt"

	"k8s-lx1036/k8s/scheduler/pkg/scheduler/apis/config"
	"k8s-lx1036/k8s/scheduler/pkg/scheduler/apis/config/validation"
	framework "k8s-lx1036/k8s/scheduler/pkg/scheduler/framework/v1alpha1"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const LeastAllocatedName = "NodeResourcesLeastAllocated"

// LeastAllocated is a score plugin that favors nodes with fewer allocation requested resources based on requested resources.
type LeastAllocated struct {
	handle framework.FrameworkHandle
	resourceAllocationScorer
}

// Name returns name of the plugin. It is used in logs, etc.
func (la *LeastAllocated) Name() string {
	return LeastAllocatedName
}

func (la *LeastAllocated) ScoreExtensions() framework.ScoreExtensions {
	return nil
}

func (la *LeastAllocated) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	nodeInfo, err := la.handle.SnapshotSharedLister().NodeInfos().Get(nodeName)
	if err != nil {
		return 0, framework.NewStatus(framework.Error, fmt.Sprintf("getting node %q from Snapshot: %v", nodeName, err))
	}

	// Details:
	// (cpu((capacity-sum(requested))*MaxNodeScore/capacity) + memory((capacity-sum(requested))*MaxNodeScore/capacity))/weightSum
	return la.score(pod, nodeInfo)
}

// INFO: 这个字段会被 NodeResourcesLeastAllocated plugin 使用，和 Requested 字段意思类似，但是如果 pod request 没有设置值，也需要根据一个默认值去
// 统计，所以 (allocatable - NonZeroRequested[cpu]) / allocatable 就表示 sum(request) 占该 node allocatable 资源比率，哪个 node 比率最小分数最高
func leastRequestedScore(requested, capacity int64) int64 {
	if capacity == 0 {
		return 0
	}
	if requested > capacity {
		return 0
	}

	return ((capacity - requested) * int64(framework.MaxNodeScore)) / capacity
}

func (la *LeastAllocated) leastResourceScorer(requested, allocable resourceToValueMap, includeVolumes bool,
	requestedVolumes int, allocatableVolumes int) int64 {
	var nodeScore, weightSum int64
	for resource, weight := range la.resourceToWeightMap {
		resourceScore := leastRequestedScore(requested[resource], allocable[resource])
		nodeScore += resourceScore * weight
		weightSum += weight
	}

	return nodeScore / weightSum
}

func NewLeastAllocated(laArgs runtime.Object, h framework.FrameworkHandle) (framework.Plugin, error) {
	args, ok := laArgs.(*config.NodeResourcesLeastAllocatedArgs)
	if !ok {
		return nil, fmt.Errorf("want args to be of type NodeResourcesLeastAllocatedArgs, got %T", laArgs)
	}
	if err := validation.ValidateNodeResourcesLeastAllocatedArgs(args); err != nil {
		return nil, err
	}

	resToWeightMap := make(resourceToWeightMap)
	for _, resource := range (*args).Resources {
		resToWeightMap[v1.ResourceName(resource.Name)] = resource.Weight
	}

	leastAllocated := &LeastAllocated{
		handle: h,
		resourceAllocationScorer: resourceAllocationScorer{
			Name: LeastAllocatedName,
			//scorer:              leastResourceScorer(resToWeightMap),
			resourceToWeightMap: resToWeightMap,
		},
	}

	leastAllocated.resourceAllocationScorer.scorer = leastAllocated.leastResourceScorer

	return leastAllocated, nil
}
