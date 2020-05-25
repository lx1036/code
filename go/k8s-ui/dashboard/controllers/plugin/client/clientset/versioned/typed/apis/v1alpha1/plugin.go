package v1alpha1

import (
	"k8s-lx1036/k8s-ui/dashboard/controllers/plugin/apis/v1alpha1"
	"k8s-lx1036/k8s-ui/dashboard/controllers/plugin/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	types2 "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	watch2 "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
	"time"
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

// plugins implements PluginInterface
type plugins struct {
	client rest.Interface
	ns     string
}

func (p *plugins) Create(*v1alpha1.Plugin) (*v1alpha1.Plugin, error) {
	panic("implement me")
}

func (p *plugins) Update(*v1alpha1.Plugin) (*v1alpha1.Plugin, error) {
	panic("implement me")
}

func (p *plugins) Delete(name string, options *v12.DeleteOptions) error {
	panic("implement me")
}

func (p *plugins) DeleteCollection(options *v12.DeleteOptions, listOptions v12.ListOptions) error {
	panic("implement me")
}

func (p *plugins) Get(name string, options v12.GetOptions) (*v1alpha1.Plugin, error) {
	panic("implement me")
}

// List takes label and field selectors, and returns the list of Plugins that match those selectors.
func (p *plugins) List(opts v12.ListOptions) (result *v1alpha1.PluginList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second // e.g. => 10s
	}

	result = &v1alpha1.PluginList{}
	err = p.client.Get().Namespace(p.ns).Resource("plugins").VersionedParams(&opts, scheme.ParameterCodec).Timeout(timeout).Do().Into(result)
	return
}

func (p *plugins) Watch(opts v12.ListOptions) (watch2.Interface, error) {
	panic("implement me")
}

func (p *plugins) Patch(name string, pt types2.PatchType, data []byte, subresources ...string) (result *v1alpha1.Plugin, err error) {
	panic("implement me")
}

// newPlugins returns a Plugins
func newPlugins(c *DashboardV1alpha1Client, namespace string) *plugins {
	return &plugins{
		client: c.RESTClient(),
		ns:     namespace,
	}
}
