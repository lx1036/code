package nodename

import (
	framework "k8s-lx1036/k8s/scheduler/pkg/scheduler/framework/v1alpha1"

	"k8s.io/apimachinery/pkg/runtime"
)

const (
	// Name is the name of the plugin used in the plugin registry and configurations.
	Name = "NodeName"

	// ErrReason returned when node name doesn't match.
	ErrReason = "node(s) didn't match the requested hostname"
)

// NodeName is a plugin that checks if a pod spec node name matches the current node.
type NodeName struct{}

// Name returns name of the plugin. It is used in logs, etc.
func (pl *NodeName) Name() string {
	return Name
}

// New initializes a new plugin and returns it.
func New(_ runtime.Object, _ framework.FrameworkHandle) (framework.Plugin, error) {
	return &NodeName{}, nil
}
