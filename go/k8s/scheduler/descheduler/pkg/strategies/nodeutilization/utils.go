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
			// A threshold is in percentages but in <0;100> interval.
			// Performing `threshold * 0.01` will convert <0;100> interval into <0;1>.
			// Multiplying it with capacity will give fraction of the capacity corresponding to the given high/low resource threshold in Quantity units.
			lowResourceThreshold: map[v1.ResourceName]*resource.Quantity{
				v1.ResourceCPU:    resource.NewMilliQuantity(int64(float64(lowThreshold[v1.ResourceCPU])*float64(nodeCapacity.Cpu().MilliValue())*0.01), resource.DecimalSI),
				v1.ResourceMemory: resource.NewQuantity(int64(float64(lowThreshold[v1.ResourceMemory])*float64(nodeCapacity.Memory().Value())*0.01), resource.BinarySI),
				v1.ResourcePods:   resource.NewQuantity(int64(float64(lowThreshold[v1.ResourcePods])*float64(nodeCapacity.Pods().Value())*0.01), resource.DecimalSI),
			},
			highResourceThreshold: map[v1.ResourceName]*resource.Quantity{
				// TODO: 这里 `threshold * 0.01` ???
				v1.ResourceCPU:    resource.NewMilliQuantity(int64(float64(highThreshold[v1.ResourceCPU])*float64(nodeCapacity.Cpu().MilliValue())*0.01), resource.DecimalSI),
				v1.ResourceMemory: resource.NewQuantity(int64(float64(highThreshold[v1.ResourceMemory])*float64(nodeCapacity.Memory().Value())*0.01), resource.BinarySI),
				v1.ResourcePods:   resource.NewQuantity(int64(float64(highThreshold[v1.ResourcePods])*float64(nodeCapacity.Pods().Value())*0.01), resource.DecimalSI),
			},
		})
	}

	return nodeUsageList
}

// 计算 node 上 cpu/memory/pod 资源使用总和
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
		// TODO: 直接 cpu/memory/pods 求和？？？
		ti := nodes[i].usage[v1.ResourceMemory].Value() + nodes[i].usage[v1.ResourceCPU].MilliValue() + nodes[i].usage[v1.ResourcePods].Value()
		tj := nodes[j].usage[v1.ResourceMemory].Value() + nodes[j].usage[v1.ResourceCPU].MilliValue() + nodes[j].usage[v1.ResourcePods].Value()
		// To return sorted in descending order
		return ti > tj
	})
}
