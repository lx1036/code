package nodeutilization

import (
	"context"

	"k8s-lx1036/k8s/scheduler/descheduler/pkg/api"
	"k8s-lx1036/k8s/scheduler/descheduler/pkg/evictions"
	nodeutil "k8s-lx1036/k8s/scheduler/descheduler/pkg/node"
	podutil "k8s-lx1036/k8s/scheduler/descheduler/pkg/pod"
	"k8s-lx1036/k8s/scheduler/descheduler/pkg/utils"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	podResource "k8s.io/kubernetes/pkg/api/v1/resource"
)

const Name = "LowNodeUtilization"

const (
	// MinResourcePercentage is the minimum value of a resource's percentage
	MinResourcePercentage = 0
	// MaxResourcePercentage is the maximum value of a resource's percentage
	MaxResourcePercentage = 100
)

// NodeUsage stores a node's info, pods on it, thresholds and its resource usage
type NodeUsage struct {
	node    *v1.Node
	usage   map[v1.ResourceName]*resource.Quantity
	allPods []*v1.Pod

	lowResourceThreshold  map[v1.ResourceName]*resource.Quantity
	highResourceThreshold map[v1.ResourceName]*resource.Quantity
}

// INFO: LowNodeUtilization 插件主要还是根据 pod request 来区分资源利用率 high/low nodes
//  但是，我们需要一个插件，可以根据资源实际使用率 pod usage 来区分

// INFO: roadmap这块有一个计划 https://github.com/kubernetes-sigs/descheduler#roadmap ，集成 metrics provider 来获得
//  pod 的资源真实负载情况 usage: Integration with metrics providers for obtaining real load metrics

// INFO: scheduler plugins 这块已经实现了根据 node resource usage 来调度：https://github.com/kubernetes-sigs/scheduler-plugins/blob/master/kep/61-Trimaran-real-load-aware-scheduling/README.md

// LowNodeUtilization evicts pods from overutilized nodes to underutilized nodes.
// Note that CPU/Memory requests are used to calculate nodes' utilization and not the actual resource usage.
// 驱逐高负载Node的pods到低负载Node，这里只考虑pod的 cpu/memory request值，而不是根据pod真实使用率
func LowNodeUtilization(ctx context.Context, client clientset.Interface,
	strategy api.DeschedulerStrategy, nodes []*v1.Node, podEvictor *evictions.PodEvictor) {

	if err := validateLowNodeUtilizationParams(strategy.Params); err != nil {
		klog.ErrorS(err, "Invalid LowNodeUtilization parameters")
		return
	}

	// 从配置中获取 priority threshold，只有该优先级以下的pod才会被驱逐
	thresholdPriority, err := utils.GetPriorityFromStrategyParams(ctx, client, strategy.Params)
	if err != nil {
		klog.ErrorS(err, "Failed to get threshold priority from strategy's params")
		return
	}

	thresholds := strategy.Params.NodeResourceUtilizationThresholds.Thresholds
	targetThresholds := strategy.Params.NodeResourceUtilizationThresholds.TargetThresholds
	if err := validateStrategyConfig(thresholds, targetThresholds); err != nil {
		klog.ErrorS(err, "LowNodeUtilization config is not valid")
		return
	}

	// 阈值没有配置的，设置最大默认值
	if _, ok := thresholds[v1.ResourcePods]; !ok {
		thresholds[v1.ResourcePods] = MaxResourcePercentage
		targetThresholds[v1.ResourcePods] = MaxResourcePercentage
	}
	if _, ok := thresholds[v1.ResourceCPU]; !ok {
		thresholds[v1.ResourceCPU] = MaxResourcePercentage
		targetThresholds[v1.ResourceCPU] = MaxResourcePercentage
	}
	if _, ok := thresholds[v1.ResourceMemory]; !ok {
		thresholds[v1.ResourceMemory] = MaxResourcePercentage
		targetThresholds[v1.ResourceMemory] = MaxResourcePercentage
	}

	// 分类哪些是 lowNodes/highNodes, 这里不考虑不可调度node
	lowNodes, highNodes := classifyNodes(
		getNodeUsage(ctx, client, nodes, thresholds, targetThresholds),
		func(node *v1.Node, usage NodeUsage) bool { // node 是否是 低利用率 node
			if nodeutil.IsNodeUnschedulable(node) {
				klog.V(2).InfoS("Node is unschedulable, thus not considered as underutilized", "node", klog.KObj(node))
				return false
			}
			return isNodeWithLowUtilization(usage)
		},
		func(node *v1.Node, usage NodeUsage) bool { // node 是否是 高利用率 node
			return isNodeAboveTargetUtilization(usage)
		},
	)

	// INFO: 如果 "低利用率" nodes为 0, <NumberOfNodes或者全是，如果没有 "高利用率" nodes，则不驱逐任何pods
	klog.V(1).InfoS("Criteria for a node under utilization",
		"CPU", thresholds[v1.ResourceCPU], "Mem", thresholds[v1.ResourceMemory], "Pods", thresholds[v1.ResourcePods])
	if len(lowNodes) == 0 { // 没有 "低利用率" nodes
		klog.V(1).InfoS("No node is underutilized, nothing to do here, you might tune your thresholds further")
		return
	}
	klog.V(1).InfoS("Total number of underutilized nodes", "totalNumber", len(lowNodes))
	if len(lowNodes) < strategy.Params.NodeResourceUtilizationThresholds.NumberOfNodes { // 低利用率 nodes 数量如果小于 NumberOfNodes，可以不考虑驱逐
		klog.V(1).InfoS("Number of nodes underutilized is less than NumberOfNodes, nothing to do here", "underutilizedNodes", len(lowNodes), "numberOfNodes", strategy.Params.NodeResourceUtilizationThresholds.NumberOfNodes)
		return
	}
	if len(lowNodes) == len(nodes) { // 所有 nodes 都是低利用率
		klog.V(1).InfoS("All nodes are underutilized, nothing to do here")
		return
	}
	if len(highNodes) == 0 { // 没有 "高利用率" nodes
		klog.V(1).InfoS("All nodes are under target utilization, nothing to do here")
		return
	}
	klog.V(1).InfoS("Criteria for a node above target utilization",
		"CPU", targetThresholds[v1.ResourceCPU], "Mem", targetThresholds[v1.ResourceMemory], "Pods", targetThresholds[v1.ResourcePods])
	klog.V(1).InfoS("Number of nodes above target utilization", "totalNumber", len(highNodes))

	// evict驱逐pod，只有该优先级以下的pod才会被驱逐
	evictable := podEvictor.Evictable(evictions.WithPriorityThreshold(thresholdPriority))
	evictPodsFromHighNodes(ctx, highNodes, lowNodes, podEvictor, evictable.IsEvictable)
	klog.V(1).InfoS("Total number of pods evicted", "evictedPods", podEvictor.TotalEvicted())
}

// nodeUsages 已经拿到了所有nodes的 total_request_limit 资源使用总量，这时可以调用 lowThresholdFilter/highThresholdFilter
// function 进行分类，哪些是 lowNodes/highNodes
func classifyNodes(nodeUsages []NodeUsage, lowThresholdFilter,
	highThresholdFilter func(node *v1.Node, usage NodeUsage) bool) ([]NodeUsage, []NodeUsage) {
	var lowNodes, highNodes []NodeUsage

	for _, nodeUsage := range nodeUsages {
		if lowThresholdFilter(nodeUsage.node, nodeUsage) {
			klog.V(2).InfoS("Node is underutilized", "node", klog.KObj(nodeUsage.node),
				"usage", nodeUsage.usage, "usagePercentage", resourceUsagePercentages(nodeUsage))
			lowNodes = append(lowNodes, nodeUsage)
		} else if highThresholdFilter(nodeUsage.node, nodeUsage) {
			klog.V(2).InfoS("Node is overutilized", "node", klog.KObj(nodeUsage.node),
				"usage", nodeUsage.usage, "usagePercentage", resourceUsagePercentages(nodeUsage))
			highNodes = append(highNodes, nodeUsage)
		} else {
			klog.V(2).InfoS("Node is appropriately utilized", "node", klog.KObj(nodeUsage.node),
				"usage", nodeUsage.usage, "usagePercentage", resourceUsagePercentages(nodeUsage))
		}
	}

	return lowNodes, highNodes
}

// 开始驱逐highNodes上pod，根据优先级从低到高驱逐
func evictPodsFromHighNodes(ctx context.Context, highNodesUsage, lowNodesUsage []NodeUsage,
	podEvictor *evictions.PodEvictor, podFilter func(pod *v1.Pod) bool) {
	// 按照资源和降序排序
	sortNodesByUsage(highNodesUsage)

	// 这里统计下 lowNodes 还可以多少资源可用
	lowNodesTotalAvailableUsage := map[v1.ResourceName]*resource.Quantity{
		v1.ResourcePods:   {},
		v1.ResourceCPU:    {},
		v1.ResourceMemory: {},
	}
	var taintsOfLowNodes = make(map[string][]v1.Taint, len(lowNodesUsage))
	for _, lowNodeUsage := range lowNodesUsage {
		// 把 利用率低
		taintsOfLowNodes[lowNodeUsage.node.Name] = lowNodeUsage.node.Spec.Taints
		for name := range lowNodesTotalAvailableUsage {
			lowNodesTotalAvailableUsage[name].Add(*lowNodeUsage.highResourceThreshold[name])
			lowNodesTotalAvailableUsage[name].Sub(*lowNodeUsage.usage[name]) // INFO: 最高阈值-已经使用量=剩余量，这里用最高阈值而不是node allocatable来统计，很精妙
		}
	}
	klog.V(1).InfoS(
		"Total capacity to be moved",
		"CPU", lowNodesTotalAvailableUsage[v1.ResourceCPU].MilliValue(),
		"Mem", lowNodesTotalAvailableUsage[v1.ResourceMemory].Value(),
		"Pods", lowNodesTotalAvailableUsage[v1.ResourcePods].Value(),
	)

	// INFO: 开始驱逐highNodes上的pods
	for _, highNodeUsage := range highNodesUsage {
		klog.V(3).InfoS("Evicting pods from node", "node", klog.KObj(highNodeUsage.node), "usage", highNodeUsage.usage)
		// podFilter函数会判断哪些pod是需要驱逐的
		nonRemovablePods, removablePods := classifyPods(highNodeUsage.allPods, podFilter)
		klog.V(2).InfoS("Pods on node", "node", klog.KObj(highNodeUsage.node),
			"allPods", len(highNodeUsage.allPods), "nonRemovablePods", len(nonRemovablePods), "removablePods", len(removablePods))
		if len(removablePods) == 0 {
			klog.V(1).InfoS("No removable pods on node, try next node", "node", klog.KObj(highNodeUsage.node))
			continue
		}

		// 开始驱逐那些priority小于指定阈值threshold的pod
		klog.V(1).InfoS("Evicting pods based on priority, if they have same priority, they'll be evicted based on QoS tiers")
		// 根据 pod 优先级升序排序，优先级相等则根据QoS(BestEffort, Burstable, Guaranteed)升序排序
		podutil.SortPodsBasedOnPriorityLowToHigh(removablePods)
		evictPods(ctx, removablePods, highNodeUsage, lowNodesTotalAvailableUsage, taintsOfLowNodes, podEvictor)
		klog.V(1).InfoS("Evicted pods from node", "node", klog.KObj(highNodeUsage.node),
			"evictedPods", podEvictor.NodeEvicted(highNodeUsage.node), "usage", highNodeUsage.usage)
	}
}

// 驱逐pod时也讲究策略，每次驱逐一个pod，判断下是否处于 high threshold 之下，达到了 < highThreshold 了就停止驱逐pod
func evictPods(ctx context.Context, removablePods []*v1.Pod, highNodeUsage NodeUsage,
	lowNodesTotalAvailableUsage map[v1.ResourceName]*resource.Quantity, taintsOfLowNodes map[string][]v1.Taint,
	podEvictor *evictions.PodEvictor) {
	continueCond := func() bool {
		// 判断是否处于lowResourceThreshold高阈值之上
		if !isNodeAboveTargetUtilization(highNodeUsage) {
			return false
		}
		if lowNodesTotalAvailableUsage[v1.ResourcePods].CmpInt64(0) < 1 {
			return false
		}
		if lowNodesTotalAvailableUsage[v1.ResourceCPU].CmpInt64(0) < 1 {
			return false
		}
		if lowNodesTotalAvailableUsage[v1.ResourceMemory].CmpInt64(0) < 1 {
			return false
		}
		return true
	}

	// true表示就继续驱逐pod
	if continueCond() {
		for _, pod := range removablePods { // 一个个驱逐pod，每次判断 nodeUsage 是否 < highThreshold，达到了就不驱逐了
			if !podutil.PodToleratesTaints(pod, taintsOfLowNodes) {
				// INFO: 这里有个重要逻辑，如果所有的 lowNodes，pod都不能容忍其 taints，则这个pod不用驱逐了
				klog.V(3).InfoS("Skipping eviction for pod, doesn't tolerate node taint", "pod", klog.KObj(pod))

				continue
			}

			success, err := podEvictor.EvictPod(ctx, pod, highNodeUsage.node, "LowNodeUtilization")
			if err != nil {
				klog.ErrorS(err, "Error evicting pod", "pod", klog.KObj(pod))
				break
			}
			if success { // 如果驱逐这个pod成功，重新判断是否继续驱逐下一个pod
				klog.V(3).InfoS("Evicted pods", "pod", klog.KObj(pod), "err", err)

				// INFO: 为何是 Sub() ???
				// 因为lowNodesTotalAvailableUsage表示的是 lowNodes 的剩余可用资源量，这里从 highNodes 驱逐pods，当然lowNodesTotalAvailableUsage 可用量就会减少

				// 获取 pod cpu request quantity值
				cpuQuantity := podResource.GetResourceRequestQuantity(pod, v1.ResourceCPU)
				highNodeUsage.usage[v1.ResourceCPU].Sub(cpuQuantity)
				lowNodesTotalAvailableUsage[v1.ResourceCPU].Sub(cpuQuantity)

				// 获取 pod memory request quantity值
				memoryQuantity := podResource.GetResourceRequestQuantity(pod, v1.ResourceMemory)
				highNodeUsage.usage[v1.ResourceMemory].Sub(memoryQuantity)
				lowNodesTotalAvailableUsage[v1.ResourceMemory].Sub(memoryQuantity)

				highNodeUsage.usage[v1.ResourcePods].Sub(*resource.NewQuantity(1, resource.DecimalSI))         // pod - 1
				lowNodesTotalAvailableUsage[v1.ResourcePods].Sub(*resource.NewQuantity(1, resource.DecimalSI)) // pod - 1

				klog.V(3).InfoS("Updated node usage", "updatedUsage", highNodeUsage)

				if !continueCond() { // 如果已经达到了highThreshold就不需要驱逐
					break
				}
			}
		}
	}
}
