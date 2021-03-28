package api

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	v1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/metrics/pkg/apis/metrics"
)

type nodeMetrics struct {
	groupResource schema.GroupResource
	metrics       NodeMetricsGetter
	nodeLister    v1listers.NodeLister
}

func newNodeMetrics(groupResource schema.GroupResource, metrics NodeMetricsGetter, nodeLister v1listers.NodeLister) *nodeMetrics {
	return &nodeMetrics{
		groupResource: groupResource,
		metrics:       metrics,
		nodeLister:    nodeLister,
	}
}

func (m *nodeMetrics) New() runtime.Object {
	return &metrics.NodeMetrics{}
}
