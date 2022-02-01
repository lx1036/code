package content

import (
	"k8s-lx1036/k8s/kubelet/containerd/containerd/pkg/plugin"
	"k8s-lx1036/k8s/kubelet/containerd/containerd/pkg/services"
)

func init() {
	plugin.Register(&plugin.Registration{
		Type:   plugin.ContentPlugin,
		ID:     services.ContentService,
		InitFn: initFunc,
	})
}

func initFunc(ic *plugin.InitContext) (interface{}, error) {
	return NewStore(ic.Root)
}
