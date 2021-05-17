package apiserver

import (
	"net/url"

	listersv1 "k8s.io/client-go/listers/core/v1"
)

// A ServiceResolver knows how to get a URL given a service.
type ServiceResolver interface {
	ResolveEndpoint(namespace, name string, port int32) (*url.URL, error)
}

// NewClusterIPServiceResolver returns a ServiceResolver that directly calls the
// service's cluster IP.
func NewClusterIPServiceResolver(services listersv1.ServiceLister) ServiceResolver {
	return &aggregatorClusterRouting{
		services: services,
	}
}
