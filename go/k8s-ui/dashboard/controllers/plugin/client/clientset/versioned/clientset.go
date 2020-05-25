package versioned

import (
	dashboardv1alpha1 "k8s-lx1036/k8s-ui/dashboard/controllers/plugin/client/clientset/versioned/typed/apis/v1alpha1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
)

type Interface interface {
	Discovery() discovery.DiscoveryInterface
	DashboardV1alpha1() dashboardv1alpha1.DashboardV1alpha1Interface
}

// Clientset contains the clients for groups. Each group has exactly one
// version included in a Clientset.
type Clientset struct {
	*discovery.DiscoveryClient
	dashboardV1alpha1 *dashboardv1alpha1.DashboardV1alpha1Client
}

// DashboardV1alpha1 retrieves the DashboardV1alpha1Client
func (c *Clientset) DashboardV1alpha1() dashboardv1alpha1.DashboardV1alpha1Interface {
	return c.dashboardV1alpha1
}

// Discovery retrieves the DiscoveryClient
func (c *Clientset) Discovery() discovery.DiscoveryInterface {
	if c == nil {
		return nil
	}
	return c.DiscoveryClient
}

// NewForConfig creates a new Clientset for the given config.
// If config's RateLimiter is not set and QPS and Burst are acceptable,
// NewForConfig will generate a rate-limiter in configShallowCopy.
func NewForConfig(config *rest.Config) (*Clientset, error) {
	configShallowCopy := *config

	var cs Clientset
	var err error
	cs.dashboardV1alpha1, err = dashboardv1alpha1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.DiscoveryClient, err = discovery.NewDiscoveryClientForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	return &cs, nil
}
