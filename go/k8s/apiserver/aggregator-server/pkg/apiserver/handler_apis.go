package apiserver

import (
	listers "k8s-lx1036/k8s/apiserver/aggregator-server/pkg/client/listers/apiregistration/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

// apisHandler serves the `/apis` endpoint.
// This is registered as a filter so that it never collides with any explicitly registered endpoints
type apisHandler struct {
	codecs         serializer.CodecFactory
	lister         listers.APIServiceLister
	discoveryGroup metav1.APIGroup
}
