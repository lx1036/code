package nodeutilization

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog/v2"
	"sort"

	"k8s-lx1036/k8s/scheduler/descheduler/pkg/api"
	podutil "k8s-lx1036/k8s/scheduler/descheduler/pkg/pod"

	v1 "k8s.io/api/core/v1"
	clientset "k8s.io/client-go/kubernetes"
)

func validateLowNodeUtilizationParams(params *api.StrategyParameters) error {
	if params == nil || params.NodeResourceUtilizationThresholds == nil {
		return fmt.Errorf("NodeResourceUtilizationThresholds not set")
	}
	if params.ThresholdPriority != nil && params.ThresholdPriorityClassName != "" {
		return fmt.Errorf("only one of thresholdPriority and thresholdPriorityClassName can be set")
	}

	return nil
}

func validateStrategyConfig(thresholds, targetThresholds api.ResourceThresholds) error {
	if err := validateThresholds(thresholds); err != nil {
		return fmt.Errorf("thresholds config is not valid: %v", err)
	}
	if err := validateThresholds(targetThresholds); err != nil {
		return fmt.Errorf("targetThresholds config is not valid: %v", err)
	}

	/*
		// thresholds 值得小于 targetThresholds
		params:
			nodeResourceUtilizationThresholds:
			thresholds:
				"memory": 20
			targetThresholds:
				"memory": 70
	*/
	// validate if thresholds and targetThresholds have same resources configured
	if len(thresholds) != len(targetThresholds) {
		return fmt.Errorf("thresholds and targetThresholds configured different resources")
	}
	for resourceName, value := range thresholds {
		if targetValue, ok := targetThresholds[resourceName]; !ok {
			return fmt.Errorf("thresholds and targetThresholds configured different resources")
		} else if value > targetValue {
			return fmt.Errorf("thresholds' %v percentage is greater than targetThresholds'", resourceName)
		}
	}

	return nil
}

// 只支持 cpu/memory/pods resource
func validateThresholds(thresholds api.ResourceThresholds) error {
	if thresholds == nil || len(thresholds) == 0 {
		return fmt.Errorf("no resource threshold is configured")
	}
	for name, percent := range thresholds {
		switch name {
		case v1.ResourceCPU, v1.ResourceMemory, v1.ResourcePods:
			if percent < MinResourcePercentage || percent > MaxResourcePercentage {
				return fmt.Errorf("%v threshold not in [%v, %v] range", name, MinResourcePercentage, MaxResourcePercentage)
			}
		default:
			return fmt.Errorf("only cpu, memory, or pods thresholds can be specified")
		}
	}
	return nil
}

// 统计node信息，包括pod数量，资源使用request/limit总和
func getNodeUsage(ctx context.Context, client clientset.Interface, nodes []*v1.Node,
	lowThreshold, highThreshold api.ResourceThresholds) []NodeUsage {

	nodeUsageList := []NodeUsage{}

	for _, node := range nodes {
		pods, err := podutil.ListPodsOnNode(ctx, client, node)
		if err != nil {
			klog.V(2).InfoS("Node will not be processed, error accessing its pods", "node", klog.KObj(node), "err", err)
			continue
		}

		nodeCapacity := node.Status.Capacity
		if len(node.Status.Allocatable) > 0 {
			nodeCapacity = node.Status.Allocatable
		}

		nodeUsageList = append(nodeUsageList, NodeUsage{
			node:    node,
			usage:   nodeUtilization(node, pods),
			allPods: pods,
			/*
				// thresholds 值得小于 targetThresholds
				params:
					nodeResourceUtilizationThresholds:
					thresholds:
						"memory": 20
					targetThresholds:
						"memory": 70
			*/
			// lowThreshold阈值(0-100)得乘以 0.01，然后才能乘以当前node容量值，才是最小阈值
			lowResourceThreshold: map[v1.ResourceName]*resource.Quantity{
				v1.ResourceCPU:    resource.NewMilliQuantity(int64(float64(lowThreshold[v1.ResourceCPU])*float64(nodeCapacity.Cpu().MilliValue())*0.01), resource.DecimalSI),
				v1.ResourceMemory: resource.NewQuantity(int64(float64(lowThreshold[v1.ResourceMemory])*float64(nodeCapacity.Memory().Value())*0.01), resource.BinarySI),
				v1.ResourcePods:   resource.NewQuantity(int64(float64(lowThreshold[v1.ResourcePods])*float64(nodeCapacity.Pods().Value())*0.01), resource.DecimalSI),
			},
			highResourceThreshold: map[v1.ResourceName]*resource.Quantity{
				v1.ResourceCPU:    resource.NewMilliQuantity(int64(float64(highThreshold[v1.ResourceCPU])*float64(nodeCapacity.Cpu().MilliValue())*0.01), resource.DecimalSI),
				v1.ResourceMemory: resource.NewQuantity(int64(float64(highThreshold[v1.ResourceMemory])*float64(nodeCapacity.Memory().Value())*0.01), resource.BinarySI),
				v1.ResourcePods:   resource.NewQuantity(int64(float64(highThreshold[v1.ResourcePods])*float64(nodeCapacity.Pods().Value())*0.01), resource.DecimalSI),
			},
		})
	}

	return nodeUsageList
}

// 判断是否处于lowResourceThreshold低阈值之上
func isNodeWithLowUtilization(usage NodeUsage) bool {
	for name, nodeValue := range usage.usage {
		// usage.lowResourceThreshold[name] < nodeValue, 在低阈值之上
		if usage.lowResourceThreshold[name].Cmp(*nodeValue) == -1 {
			return false
		}
	}

	return true
}

// 判断是否处于 highResourceThreshold 高阈值之下
func isNodeAboveTargetUtilization(usage NodeUsage) bool {
	for name, nodeValue := range usage.usage {
		// usage.highResourceThreshold[name] < nodeValue, 在高阈值之下
		if usage.highResourceThreshold[name].Cmp(*nodeValue) == -1 {
			return true
		}
	}

	return false
}

// 计算 node 上 cpu/memory/pod 资源使用总和，计算的是request/limit值，而不是pod真实使用率
func nodeUtilization(node *v1.Node, pods []*v1.Pod) map[v1.ResourceName]*resource.Quantity {
	totalReqs := map[v1.ResourceName]*resource.Quantity{
		v1.ResourceCPU:    resource.NewMilliQuantity(0, resource.DecimalSI),
		v1.ResourceMemory: resource.NewQuantity(0, resource.BinarySI),
		v1.ResourcePods:   resource.NewQuantity(int64(len(pods)), resource.DecimalSI),
	}
	for _, pod := range pods {
		req, _ := podutil.PodRequestsAndLimits(pod)
		for name, quantity := range req {
			if name == v1.ResourceCPU || name == v1.ResourceMemory {
				// As Quantity.Add says: Add adds the provided y quantity to the current value. If the current value is zero,
				// the format of the quantity will be updated to the format of y.
				totalReqs[name].Add(quantity)
			}
		}
	}

	return totalReqs
}

// sortNodesByUsage sorts nodes based on usage in descending order
func sortNodesByUsage(nodes []NodeUsage) {
	sort.Slice(nodes, func(i, j int) bool {
		// INFO: 直接 cpu/memory/pods 求和？？？
		ti := nodes[i].usage[v1.ResourceMemory].Value() + nodes[i].usage[v1.ResourceCPU].MilliValue() + nodes[i].usage[v1.ResourcePods].Value()
		tj := nodes[j].usage[v1.ResourceMemory].Value() + nodes[j].usage[v1.ResourceCPU].MilliValue() + nodes[j].usage[v1.ResourcePods].Value()
		// To return sorted in descending order
		return ti > tj
	})
}

// 根据filter函数过滤哪些pod是需要驱逐的
func classifyPods(pods []*v1.Pod, filter func(pod *v1.Pod) bool) ([]*v1.Pod, []*v1.Pod) {
	var nonRemovablePods, removablePods []*v1.Pod

	for _, pod := range pods {
		if !filter(pod) {
			nonRemovablePods = append(nonRemovablePods, pod)
		} else {
			removablePods = append(removablePods, pod)
		}
	}

	return nonRemovablePods, removablePods
}

// 计算各个资源使用百分比
func resourceUsagePercentages(nodeUsage NodeUsage) map[v1.ResourceName]float64 {
	nodeCapacity := nodeUsage.node.Status.Capacity
	if len(nodeUsage.node.Status.Allocatable) > 0 {
		nodeCapacity = nodeUsage.node.Status.Allocatable
	}

	resourceUsagePercentage := map[v1.ResourceName]float64{}
	for resourceName, resourceUsage := range nodeUsage.usage {
		c := nodeCapacity[resourceName]
		if !c.IsZero() {
			resourceUsagePercentage[resourceName] = 100 * float64(resourceUsage.Value()) / float64(c.Value())
		}
	}

	return resourceUsagePercentage
}
