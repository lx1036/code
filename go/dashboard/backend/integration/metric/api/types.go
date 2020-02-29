package api

import (
	"k8s-lx1036/dashboard/backend/api"
	integrationapi "k8s-lx1036/dashboard/backend/integration/api"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"time"
)

// MetricClient is an interface that exposes API used by dashboard to show graphs and sparklines.
type MetricClient interface {
	// DownloadMetric returns MetricPromises for specified list of selector, for single type
	// of metric, i.e. cpu usage. Cached resources is usually list of pods as other high level
	// resources do not directly provide metrics. Only pods targeted by them.
	DownloadMetric(selectors []ResourceSelector, metricName string, cachedResources *CachedResources) MetricPromises
	// DownloadMetrics is similar to DownloadMetric method. It returns MetricPromises for
	// given list of metrics, i.e. cpu/memory usage instead of single metric type.
	DownloadMetrics(selectors []ResourceSelector, metricNames []string, cachedResources *CachedResources) MetricPromises
	// AggregateMetrics is used to aggregate previously downloaded metrics based on
	// aggregation mode (sum, min, avg). It is used to show cumulative metric graphs on
	// resource list pages.
	AggregateMetrics(metrics MetricPromises, metricName string, aggregations AggregationModes) MetricPromises

	// Implements IntegrationApp interface
	integrationapi.Integration
}

// ResourceSelector is a structure used to quickly and uniquely identify given resource.
// This struct can be later used for heapster data download etc.
type ResourceSelector struct {
	// Namespace of this resource.
	Namespace string
	// Type of this resource
	ResourceType api.ResourceKind
	// Name of this resource.
	ResourceName string
	// Selector used to identify this resource (should be used only for Deployments!).
	Selector map[string]string
	// UID is resource unique identifier.
	UID types.UID
}

// CachedResources contains all resources that may be required by DataSelect functions for metric
// gathering. Depending on the need you may have to provide DataSelect with resources it
// requires, for example resource like deployment will need Pods in order to calculate its metrics.
type CachedResources struct {
	Pods []v1.Pod
	// More cached resources can be added.
	// For example, if you want to use Events from DataSelect, you will have to add them here.
}

type MetricPromises []MetricPromise

// MetricPromise is used for parallel data extraction. Contains len 1 channels for Metric and Error.
type MetricPromise struct {
	Metric chan *Metric
	Error  chan error
}

// Metric is a format of data used in this module. This is also the format of data that is being sent by backend API.
type Metric struct {
	// DataPoints is a list of X, Y int64 data points, sorted by X.
	DataPoints `json:"dataPoints"`
	// MetricPoints is a list of value, timestamp metrics used for sparklines on a pod list page.
	MetricPoints []MetricPoint `json:"metricPoints"`
	// MetricName is the name of metric stored in this struct.
	MetricName string `json:"metricName"`
	// Label stores information about identity of resources (UIDS) described by this metric.
	Label `json:"-"`
	// Names of aggregating function used.
	Aggregate AggregationMode `json:"aggregation,omitempty"`
}

type DataPoints []DataPoint

type DataPoint struct {
	X int64 `json:"x"`
	Y int64 `json:"y"`
}

type MetricPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     uint64    `json:"value"`
}

// Label stores information about identity of resources (UIDs) described by metric.
type Label map[api.ResourceKind][]types.UID

// AggregationMode informs how data should be aggregated (sum, min, max)
type AggregationMode string
type AggregationModes []AggregationMode
