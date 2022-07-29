package plugins

import (
	"k8s-lx1036/k8s/scheduler/pkg/framework/plugins/nodename"
	"k8s-lx1036/k8s/scheduler/pkg/framework/runtime"
)

// NewInTreeRegistry builds the registry with all the in-tree plugins.
// A scheduler that runs out of tree plugins can register additional plugins
// through the WithFrameworkOutOfTreeRegistry option.
func NewInTreeRegistry() runtime.Registry {
	return runtime.Registry{
		nodename.Name: nodename.New,
	}
}
