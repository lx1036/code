package versioned

import (
	dashboardv1alpha1 "k8s-lx1036/dashboard/backend/plugin/client/clientset/versioned/typed/apis/v1alpha1"
	"k8s.io/client-go/discovery"
)

type Interface interface {
	Discovery() discovery.DiscoveryInterface
	DashboardV1alpha1() dashboardv1alpha1.DashboardV1alpha1Interface
}
