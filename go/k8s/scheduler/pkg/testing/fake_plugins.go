package testing

import (
	"context"
	"fmt"
	"sync/atomic"

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

type FakeFilterPlugin struct {
	NumFilterCalled         int32
	FailedNodeReturnCodeMap map[string]framework.Code
}

func NewFakeFilterPlugin(_ runtime.Object, _ *frameworkruntime.Framework) (framework.Plugin, error) {
	return &FakeFilterPlugin{}, nil
}
func (pl *FakeFilterPlugin) Name() string {
	return "FakeFilter"
}
func (pl *FakeFilterPlugin) Filter(_ context.Context, _ *framework.CycleState, pod *corev1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	atomic.AddInt32(&pl.NumFilterCalled, 1) // pl.NumFilterCalled++
	if returnCode, ok := pl.FailedNodeReturnCodeMap[nodeInfo.Node().Name]; ok {
		return framework.NewStatus(returnCode, fmt.Sprintf("injecting failure for pod %v", pod.Name))
	}

	return nil
}

type TestPlugin struct {
	name string
}

func NewTestPlugin(_ runtime.Object, _ *frameworkruntime.Framework) (framework.Plugin, error) {
	return &TestPlugin{name: "test-plugin"}, nil
}
func (pl *TestPlugin) Name() string {
	return pl.name
}
func (pl *TestPlugin) PreFilter(ctx context.Context, state *framework.CycleState, p *corev1.Pod) (*framework.PreFilterResult, *framework.Status) {
	return nil, nil
}

func (pl *TestPlugin) PreFilterExtensions() framework.PreFilterExtensions {
	return pl
}

func (pl *TestPlugin) AddPod(ctx context.Context, state *framework.CycleState, podToSchedule *corev1.Pod, podInfoToAdd *framework.PodInfo, nodeInfo *framework.NodeInfo) *framework.Status {
	if value, ok := nodeInfo.Node().GetLabels()["error"]; ok && value == "true" {
		return framework.AsStatus(fmt.Errorf("failed to add pod: %v", podToSchedule.Name))
	}
	return nil
}

func (pl *TestPlugin) RemovePod(ctx context.Context, state *framework.CycleState, podToSchedule *corev1.Pod, podInfoToRemove *framework.PodInfo, nodeInfo *framework.NodeInfo) *framework.Status {
	if value, ok := nodeInfo.Node().GetLabels()["error"]; ok && value == "true" {
		return framework.AsStatus(fmt.Errorf("failed to remove pod: %v", podToSchedule.Name))
	}
	return nil
}
