package defaultbinder

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s-lx1036/k8s/scheduler/pkg/framework"
	frameworkruntime "k8s-lx1036/k8s/scheduler/pkg/framework/runtime"

	"k8s.io/apimachinery/pkg/runtime"
)

const (
	Name = "DefaultBinder"
)

type DefaultBinder struct {
	framework *frameworkruntime.Framework
}

func New(_ runtime.Object, framework *frameworkruntime.Framework) (framework.Plugin, error) {
	return &DefaultBinder{framework: framework}, nil
}

func (b DefaultBinder) Name() string {
	return Name
}

// Bind 创建 Kind: Bind 对象
func (b DefaultBinder) Bind(ctx context.Context, state *framework.CycleState, p *corev1.Pod, nodeName string) *framework.Status {
	binding := &corev1.Binding{
		ObjectMeta: metav1.ObjectMeta{Namespace: p.Namespace, Name: p.Name, UID: p.UID},
		Target:     corev1.ObjectReference{Kind: "Node", Name: nodeName},
	}
	err := b.framework.ClientSet().CoreV1().Pods(binding.Namespace).Bind(ctx, binding, metav1.CreateOptions{})
	if err != nil {
		return framework.AsStatus(err)
	}
	return nil
}
