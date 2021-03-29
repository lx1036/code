package priorityclass

import (
	"context"
	"fmt"
	"strconv"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	frameworkRuntime "k8s.io/kubernetes/pkg/scheduler/framework/runtime"
	"k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"
)

const Name = "priority-class-fit"

const (
	CPUAllocateAnnotation    = "allocatable.resource/cpu"
	MemoryAllocateAnnotation = "allocatable.resource/memory"
)

type Args struct {
	PriorityClassName string `json:"priorityClassName,omitempty"`
	Ratio             int    `json:"ratio,omitempty"`
}

// @see https://github.com/kubernetes/kubernetes/blob/master/pkg/scheduler/framework/plugins/noderesources/fit.go
type PriorityClassFit struct {
	args   *Args
	handle v1alpha1.FrameworkHandle
}

func (s *PriorityClassFit) Name() string {
	return Name
}

func (s *PriorityClassFit) Filter(c context.Context, state *v1alpha1.CycleState, pod *v1.Pod, nodeInfo *v1alpha1.NodeInfo) *v1alpha1.Status {
	if pod.Spec.PriorityClassName != s.args.PriorityClassName {
		return nil
	}

	// calculate sum * ratio cpu/memory
	node := nodeInfo.Node()
	cpuStr, ok := node.Annotations[CPUAllocateAnnotation] // e.g. "30"
	if !ok {
		return nil
	}
	cpu, err := strconv.ParseInt(cpuStr, 10, 0)
	if err != nil {
		klog.Errorf("parse int err %v", err)
		return nil
	}
	cpu = cpu * 1000 * int64(s.args.Ratio)
	sumCores := resource.MustParse(fmt.Sprintf("%dm", cpu))
	klog.Infof("sum * ratio: milliCores = %v (%v) in node %s", sumCores.MilliValue(), sumCores.Format, nodeInfo.Node().Name)
	memoryStr, ok := node.Annotations[MemoryAllocateAnnotation] // e.g. "122991640Ki"
	if !ok {
		return nil
	}
	memory := resource.MustParse(memoryStr)
	sumMemory := resource.NewQuantity(memory.Value()*int64(s.args.Ratio), resource.BinarySI)
	klog.Infof("sum * ratio: memorySize = %v (%v) in node %s", sumMemory.Value(), sumMemory.Format, nodeInfo.Node().Name)

	consumedCores := resource.NewMilliQuantity(0, resource.DecimalSI)
	consumedMemory := resource.NewQuantity(0, resource.BinarySI)
	for _, item := range nodeInfo.Pods {
		if item.Pod.Spec.NodeName == node.Name && item.Pod.Spec.PriorityClassName == s.args.PriorityClassName {
			// calculate consumed cpu/memory
			for _, container := range item.Pod.Spec.Containers {
				for name, quantity := range container.Resources.Requests {
					switch name {
					case v1.ResourceCPU:
						consumedCores.Add(quantity)
					case v1.ResourceMemory:
						consumedMemory.Add(quantity)
					}
				}
			}
		}
	}

	// add current pod
	for _, container := range pod.Spec.Containers {
		for name, quantity := range container.Resources.Requests {
			switch name {
			case v1.ResourceCPU:
				consumedCores.Add(quantity)
			case v1.ResourceMemory:
				consumedMemory.Add(quantity)
			}
		}
	}

	if consumedCores.Cmp(sumCores) == -1 && consumedMemory.Cmp(*sumMemory) == -1 {
		return nil
	}

	return v1alpha1.NewStatus(v1alpha1.Unschedulable, "cpu or memory resource is insufficient")
}

// 在 pkg/scheduler/framework/v1alpha1/interface.go 定义
func (s *PriorityClassFit) Score(ctx context.Context, state *v1alpha1.CycleState, pod *v1.Pod, nodeName string) (int64, *v1alpha1.Status) {
	// 只考虑 "PriorityClassName" pod
	if pod.Spec.PriorityClassName != s.args.PriorityClassName {
		return 0, nil
	}

	nodeInfo, err := s.handle.SnapshotSharedLister().NodeInfos().Get(nodeName)
	if err != nil || nodeInfo.Node() == nil {
		return 0, v1alpha1.NewStatus(v1alpha1.Error, fmt.Sprintf("getting node %q from Snapshot: %v", nodeName, err))
	}
	node := nodeInfo.Node()

	score := v1alpha1.MaxNodeScore
	for _, item := range nodeInfo.Pods {
		if item.Pod.Spec.NodeName == node.Name && item.Pod.Spec.PriorityClassName == s.args.PriorityClassName {
			// TODO: 通过 pod resource 来打分，而不是个数，参考下 kube-scheduler 源码
			score-- // 高优先级pod数量越多，分数越低
		}
	}

	return score, nil
}

// ScoreExtensions of the Score plugin.
func (s *PriorityClassFit) ScoreExtensions() v1alpha1.ScoreExtensions {
	return nil
}

func New(configuration runtime.Object, handle v1alpha1.FrameworkHandle) (v1alpha1.Plugin, error) {
	args := &Args{}
	if err := frameworkRuntime.DecodeInto(configuration, args); err != nil {
		return nil, err
	}

	if len(args.PriorityClassName) == 0 {
		args.PriorityClassName = "system-cluster-critical"
	}
	if args.Ratio == 0 {
		args.Ratio = 1
	}

	klog.Infof("get plugin config args: %v", args)
	return &PriorityClassFit{
		args:   args,
		handle: handle,
	}, nil
}
