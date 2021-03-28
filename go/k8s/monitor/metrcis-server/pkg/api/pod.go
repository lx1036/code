package api

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	v1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/metrics/pkg/apis/metrics"
)

type podMetrics struct {
	groupResource schema.GroupResource
	metrics       PodMetricsGetter
	podLister     v1listers.PodLister
}

func newPodMetrics(groupResource schema.GroupResource, metrics PodMetricsGetter, podLister v1listers.PodLister) *podMetrics {
	return &podMetrics{
		groupResource: groupResource,
		metrics:       metrics,
		podLister:     podLister,
	}
}

// Storage interface
func (m *podMetrics) New() runtime.Object {
	return &metrics.PodMetrics{}
}
