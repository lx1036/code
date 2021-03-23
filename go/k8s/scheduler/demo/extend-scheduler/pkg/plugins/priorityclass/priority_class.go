package priorityclass

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	frameworkRuntime "k8s.io/kubernetes/pkg/scheduler/framework/runtime"
	"k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"
)

const Name = "sample-plugin"

type Args struct {
	FavoriteColor     string `json:"favorite_color,omitempty"`
	FavoriteNumber    int    `json:"favorite_number,omitempty"`
	ThanksTo          string `json:"thanks_to,omitempty"`
	PriorityClassName string `json:"priorityClassName,omitempty"`
}

// @see k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1/interface.go
type Sample struct {
	args   *Args
	handle v1alpha1.FrameworkHandle
}

func (s *Sample) Name() string {
	return Name
}

func (s *Sample) PreFilter(pc context.Context, state *v1alpha1.CycleState, pod *v1.Pod) *v1alpha1.Status {
	klog.V(3).Infof("prefilter pod: %v", pod.Name)
	return v1alpha1.NewStatus(v1alpha1.Success, "")
}

func (s *Sample) PreFilterExtensions() v1alpha1.PreFilterExtensions {
	return nil
}

func (s *Sample) Filter(pc context.Context, state *v1alpha1.CycleState, pod *v1.Pod, nodeName *v1alpha1.NodeInfo) *v1alpha1.Status {
	klog.V(3).Infof("filter pod: %v, node: %v", pod.Name, nodeName.Node().Name)
	return v1alpha1.NewStatus(v1alpha1.Success, "")
}

func (s *Sample) PreBind(pc context.Context, state *v1alpha1.CycleState, pod *v1.Pod, nodeName string) *v1alpha1.Status {
	if nodeInfo, err := s.handle.SnapshotSharedLister().NodeInfos().Get(nodeName); err != nil {
		return v1alpha1.NewStatus(v1alpha1.Error, fmt.Sprintf("prebind get node info error: %+v", nodeName))
	} else {
		klog.Infof("prebind node info: %+v", nodeInfo.Node().Name)
		return v1alpha1.NewStatus(v1alpha1.Success, "")
	}
}

// 在 pkg/scheduler/framework/v1alpha1/interface.go 定义
func (s *Sample) Score(ctx context.Context, state *v1alpha1.CycleState, pod *v1.Pod, nodeName string) (int64, *v1alpha1.Status) {
	if pod.Spec.PriorityClassName != s.args.PriorityClassName {
		return 0, v1alpha1.NewStatus(v1alpha1.Success)
	}

	nodeInfo, err := s.handle.SnapshotSharedLister().NodeInfos().Get(nodeName)
	if err != nil || nodeInfo.Node() == nil {
		return 0, v1alpha1.NewStatus(v1alpha1.Error, fmt.Sprintf("getting node %q from Snapshot: %v", nodeName, err))
	}
	node := nodeInfo.Node()

	score := 0
	for _, item := range nodeInfo.Pods {
		if item.Pod.Spec.NodeName == node.Name && item.Pod.Spec.PriorityClassName == s.args.PriorityClassName {
			score++
		}
	}

	return int64(score), v1alpha1.NewStatus(v1alpha1.Success)
}

func New(configuration runtime.Object, handle v1alpha1.FrameworkHandle) (v1alpha1.Plugin, error) {
	args := &Args{}
	if err := frameworkRuntime.DecodeInto(configuration, args); err != nil {
		return nil, err
	}

	if len(args.PriorityClassName) == 0 {
		return nil, fmt.Errorf("")
	}

	klog.Infof("get plugin config args: %+v", args)
	return &Sample{
		args:   args,
		handle: handle,
	}, nil
}
