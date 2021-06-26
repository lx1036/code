package nodeunschedulable

import (
	"context"

	framework "k8s-lx1036/k8s/scheduler/pkg/scheduler/framework"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	v1helper "k8s.io/kubernetes/pkg/apis/core/v1/helper"
)

const Name = "NodeUnschedulable"

const (
	// ErrReasonUnknownCondition is used for NodeUnknownCondition predicate error.
	ErrReasonUnknownCondition = "node(s) had unknown conditions"
	// ErrReasonUnschedulable is used for NodeUnschedulable predicate error.
	ErrReasonUnschedulable = "node(s) were unschedulable"
)

// NodeUnschedulable is a plugin that priorities nodes according to the node annotation
// "scheduler.alpha.kubernetes.io/preferAvoidPods".
type NodeUnschedulable struct {
}

func (pl *NodeUnschedulable) Name() string {
	return Name
}

// INFO: 如果 node.Spec.Unschedulable 为 true，但是 pod 有 "node.kubernetes.io/unschedulable" taint tolerations，同样可以调度

func (pl *NodeUnschedulable) Filter(ctx context.Context, _ *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	if nodeInfo == nil || nodeInfo.Node() == nil {
		return framework.NewStatus(framework.UnschedulableAndUnresolvable, ErrReasonUnknownCondition)
	}

	// If pod tolerate unschedulable taint, it's also tolerate `node.Spec.Unschedulable`.
	podToleratesUnschedulable := v1helper.TolerationsTolerateTaint(pod.Spec.Tolerations, &v1.Taint{
		Key:    v1.TaintNodeUnschedulable,
		Effect: v1.TaintEffectNoSchedule,
	})

	if nodeInfo.Node().Spec.Unschedulable && !podToleratesUnschedulable {
		return framework.NewStatus(framework.UnschedulableAndUnresolvable, ErrReasonUnschedulable)
	}

	return nil
}

func New(_ runtime.Object, _ framework.FrameworkHandle) (framework.Plugin, error) {
	return &NodeUnschedulable{}, nil
}
