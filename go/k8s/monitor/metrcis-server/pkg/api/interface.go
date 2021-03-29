package api

import (
	"time"

	corev1 "k8s.io/api/core/v1"
	apitypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/metrics/pkg/apis/metrics"
)

type MetricsGetter interface {
	PodMetricsGetter
	NodeMetricsGetter
}

type PodMetricsGetter interface {
	GetContainerMetrics(pods ...apitypes.NamespacedName) ([]TimeInfo, [][]metrics.ContainerMetrics)
}

type NodeMetricsGetter interface {
	GetNodeMetrics(nodes ...string) ([]TimeInfo, []corev1.ResourceList)
}

type TimeInfo struct {
	Timestamp time.Time

	Window time.Duration
}
