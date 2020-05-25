package v1alpha1

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
)

// PluginsGetter has a method to return a PluginInterface.
// A group's client should implement this interface.
type PluginsGetter interface {
	Plugins(namespace string) PluginInterface
}

// PluginInterface has methods to work with Plugin resources.
type PluginInterface interface {
	Create(*v1alpha1.Plugin) (*v1alpha1.Plugin, error)
	Update(*v1alpha1.Plugin) (*v1alpha1.Plugin, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.Plugin, error)
	List(opts v1.ListOptions) (*v1alpha1.PluginList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Plugin, err error)
	PluginExpansion
}
