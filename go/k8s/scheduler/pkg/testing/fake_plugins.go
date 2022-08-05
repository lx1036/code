package testing

import (
	"context"

	"k8s-lx1036/k8s/scheduler/pkg/framework"
	frameworkruntime "k8s-lx1036/k8s/scheduler/pkg/framework/runtime"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/runtime"
)

const ErrReasonFake = "Nodes failed the fake plugin"

type FalseFilterPlugin struct{}

func NewFalseFilterPlugin(_ runtime.Object, _ *frameworkruntime.Framework) (framework.Plugin, error) {
	return &FalseFilterPlugin{}, nil
}
func (pl *FalseFilterPlugin) Name() string {
	return "FalseFilter"
}
func (pl *FalseFilterPlugin) Filter(_ context.Context, _ *framework.CycleState, pod *corev1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	return framework.NewStatus(framework.Unschedulable, ErrReasonFake)
}
