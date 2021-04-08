package nodeutilization

import (
	"context"

	"k8s-lx1036/k8s/scheduler/descheduler/pkg/api"
	"k8s-lx1036/k8s/scheduler/descheduler/pkg/evictions"
	nodeutil "k8s-lx1036/k8s/scheduler/descheduler/pkg/node"
	"k8s-lx1036/k8s/scheduler/descheduler/pkg/utils"

	podutil "k8s-lx1036/k8s/scheduler/descheduler/pkg/pod"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
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

// LowNodeUtilization evicts pods from overutilized nodes to underutilized nodes.
// Note that CPU/Memory requests are used to calculate nodes' utilization and not the actual resource usage.
// 驱逐高负载Node的pods到低负载Node，这里只考虑pod的 cpu/memory request值，而不是根据pod真实使用率
func LowNodeUtilization(ctx context.Context, client clientset.Interface,
	strategy api.DeschedulerStrategy, nodes []*v1.Node, podEvictor *evictions.PodEvictor) {

	if err := validateLowNodeUtilizationParams(strategy.Params); err != nil {
		klog.ErrorS(err, "Invalid LowNodeUtilization parameters")
		return
	}

	// 从配置中获取 priority threshold
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
	// check if Pods/CPU/Mem are set, if not, set them to 100
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

	lowNodes, highNodes := classifyNodes(
		getNodeUsage(ctx, client, nodes, thresholds, targetThresholds),
		// The node has to be schedulable (to be able to move workload there)
		func(node *v1.Node, usage NodeUsage) bool {
			if nodeutil.IsNodeUnschedulable(node) {
				klog.V(2).InfoS("Node is unschedulable, thus not considered as underutilized", "node", klog.KObj(node))
				return false
			}
			return isNodeWithLowUtilization(usage)
		},
		func(node *v1.Node, usage NodeUsage) bool {
			return isNodeAboveTargetUtilization(usage)
		},
	)

	klog.V(1).InfoS("Criteria for a node under utilization",
		"CPU", thresholds[v1.ResourceCPU], "Mem", thresholds[v1.ResourceMemory], "Pods", thresholds[v1.ResourcePods])
	if len(lowNodes) == 0 {
		klog.V(1).InfoS("No node is underutilized, nothing to do here, you might tune your thresholds further")
		return
	}
	klog.V(1).InfoS("Total number of underutilized nodes", "totalNumber", len(lowNodes))
	if len(lowNodes) < strategy.Params.NodeResourceUtilizationThresholds.NumberOfNodes {
		klog.V(1).InfoS("Number of nodes underutilized is less than NumberOfNodes, nothing to do here", "underutilizedNodes", len(lowNodes), "numberOfNodes", strategy.Params.NodeResourceUtilizationThresholds.NumberOfNodes)
		return
	}
	if len(lowNodes) == len(nodes) {
		klog.V(1).InfoS("All nodes are underutilized, nothing to do here")
		return
	}
	if len(highNodes) == 0 {
		klog.V(1).InfoS("All nodes are under target utilization, nothing to do here")
		return
	}
	klog.V(1).InfoS("Criteria for a node above target utilization",
		"CPU", targetThresholds[v1.ResourceCPU], "Mem", targetThresholds[v1.ResourceMemory], "Pods", targetThresholds[v1.ResourcePods])
	klog.V(1).InfoS("Number of nodes above target utilization", "totalNumber", len(highNodes))

	// evict驱逐 pod
	evictable := podEvictor.Evictable(evictions.WithPriorityThreshold(thresholdPriority))
	evictPodsFromHighNodes(ctx, highNodes, lowNodes, podEvictor, evictable.IsEvictable)
	klog.V(1).InfoS("Total number of pods evicted", "evictedPods", podEvictor.TotalEvicted())
}

// classifyNodes classifies the nodes into low-utilization or high-utilization nodes. If a node lies between
// low and high thresholds, it is simply ignored.
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

// evictPodsFromTargetNodes evicts pods based on priority, if all the pods on the node have priority, if not
// evicts them based on QoS as fallback option.
func evictPodsFromHighNodes(ctx context.Context, targetNodes, lowNodes []NodeUsage,
	podEvictor *evictions.PodEvictor, podFilter func(pod *v1.Pod) bool) {
	// 按照资源和降序排序
	sortNodesByUsage(targetNodes)

	// upper bound on total number of pods/cpu/memory to be moved
	// 日志一下可用资源总和
	totalAvailableUsage := map[v1.ResourceName]*resource.Quantity{
		v1.ResourcePods:   {},
		v1.ResourceCPU:    {},
		v1.ResourceMemory: {},
	}
	var taintsOfLowNodes = make(map[string][]v1.Taint, len(lowNodes))
	for _, node := range lowNodes {
		taintsOfLowNodes[node.node.Name] = node.node.Spec.Taints
		for name := range totalAvailableUsage {
			totalAvailableUsage[name].Add(*node.highResourceThreshold[name])
			totalAvailableUsage[name].Sub(*node.usage[name])
		}
	}
	klog.V(1).InfoS(
		"Total capacity to be moved",
		"CPU", totalAvailableUsage[v1.ResourceCPU].MilliValue(),
		"Mem", totalAvailableUsage[v1.ResourceMemory].Value(),
		"Pods", totalAvailableUsage[v1.ResourcePods].Value(),
	)

	for _, node := range targetNodes {
		klog.V(3).InfoS("Evicting pods from node", "node", klog.KObj(node.node), "usage", node.usage)
		// podFilter函数会判断哪些pod是需要驱逐的
		nonRemovablePods, removablePods := classifyPods(node.allPods, podFilter)
		klog.V(2).InfoS("Pods on node", "node", klog.KObj(node.node), "allPods", len(node.allPods), "nonRemovablePods", len(nonRemovablePods), "removablePods", len(removablePods))
		if len(removablePods) == 0 {
			klog.V(1).InfoS("No removable pods on node, try next node", "node", klog.KObj(node.node))
			continue
		}

		// 开始驱逐那些priority小于指定阈值threshold的pod
		klog.V(1).InfoS("Evicting pods based on priority, if they have same priority, they'll be evicted based on QoS tiers")
		// sort the evictable Pods based on priority. This also sorts them based on QoS. If there are multiple pods with same priority, they are sorted based on QoS tiers.
		podutil.SortPodsBasedOnPriorityLowToHigh(removablePods)
		evictPods(ctx, removablePods, node, totalAvailableUsage, taintsOfLowNodes, podEvictor)
		klog.V(1).InfoS("Evicted pods from node", "node", klog.KObj(node.node), "evictedPods", podEvictor.NodeEvicted(node.node), "usage", node.usage)
	}
}
