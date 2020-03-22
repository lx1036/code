package v1alpha1

import "k8s.io/client-go/rest"

type DashboardV1alpha1Interface interface {
	RESTClient() rest.Interface
	PluginsGetter
}

// DashboardV1alpha1Client is used to interact with features provided by the dashboard.k8s.io group.
type DashboardV1alpha1Client struct {
	restClient rest.Interface
}

func (c *DashboardV1alpha1Client) Plugins(namespace string) PluginInterface {
	return newPlugins(c, namespace)
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *DashboardV1alpha1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}

// NewForConfig creates a new DashboardV1alpha1Client for the given config.
func NewForConfig(c *rest.Config) (*DashboardV1alpha1Client, error) {
	config := *c

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &DashboardV1alpha1Client{client}, nil
}
