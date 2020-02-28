package v1alpha1

import "k8s.io/client-go/rest"

type DashboardV1alpha1Interface interface {
	RESTClient() rest.Interface
	PluginsGetter
}
