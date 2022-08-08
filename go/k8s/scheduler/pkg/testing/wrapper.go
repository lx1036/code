package testing

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
)

type PodWrapper struct{ corev1.Pod }

func MakePod() *PodWrapper {
	return &PodWrapper{corev1.Pod{}}
}
func (p *PodWrapper) Name(s string) *PodWrapper {
	p.SetName(s)
	return p
}
func (p *PodWrapper) Namespace(s string) *PodWrapper {
	p.SetNamespace(s)
	return p
}
func (p *PodWrapper) UID(s string) *PodWrapper {
	p.SetUID(types.UID(s))
	return p
}
func (p *PodWrapper) Priority(val int32) *PodWrapper {
	p.Spec.Priority = &val
	return p
}
func (p *PodWrapper) Obj() *corev1.Pod {
	return &p.Pod
}
func (p *PodWrapper) Node(s string) *PodWrapper {
	p.Spec.NodeName = s
	return p
}

type NodeWrapper struct{ corev1.Node }

func MakeNode() *NodeWrapper {
	w := &NodeWrapper{corev1.Node{}}
	return w.Capacity(nil)
}
func (n *NodeWrapper) Capacity(resources map[corev1.ResourceName]string) *NodeWrapper {
	res := corev1.ResourceList{
		corev1.ResourcePods: resource.MustParse("32"), // 32 pods
	}
	for name, value := range resources {
		res[name] = resource.MustParse(value)
	}
	n.Status.Capacity, n.Status.Allocatable = res, res
	return n
}
func (n *NodeWrapper) Name(s string) *NodeWrapper {
	n.SetName(s)
	return n
}
func (n *NodeWrapper) Label(k, v string) *NodeWrapper {
	if n.Labels == nil {
		n.Labels = make(map[string]string)
	}
	n.Labels[k] = v
	return n
}
func (n *NodeWrapper) Obj() *corev1.Node {
	return &n.Node
}
