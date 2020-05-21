package versioned

import (
	"k8s.io/client-go/discovery"
)

type Interface interface {
	Discovery() discovery.DiscoveryInterface
	DashboardV1alpha1() dashboardv1alpha1.DashboardV1alpha1Interface
}
