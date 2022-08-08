package nodeaffinity

import (
	framework "k8s-lx1036/k8s/scheduler/pkg/framework"
	"k8s.io/apimachinery/pkg/runtime"
)

const Name = "NodeAffinity"

type NodeAffinity struct {
	handle framework.FrameworkHandle
}

// Name returns name of the plugin. It is used in logs, etc.
func (pl *NodeAffinity) Name() string {
	return Name
}

func New(_ runtime.Object, h framework.FrameworkHandle) (framework.Plugin, error) {
	return &NodeAffinity{handle: h}, nil
}
