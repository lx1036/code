package metric

import (
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/common/dataselect"
	"k8s.io/apimachinery/pkg/types"
	"time"
)

// MetricClient is an interface that exposes API used by dashboard to show graphs and sparklines.
type MetricClient interface {
	// DownloadMetric returns MetricPromises for specified list of selector, for single type
	// of metric, i.e. cpu usage. Cached resources is usually list of pods as other high level
	// resources do not directly provide metrics. Only pods targeted by them.
	DownloadMetric(selectors []ResourceSelector, metricName string,
		cachedResources *CachedResources) MetricPromises
	// DownloadMetrics is similar to DownloadMetric method. It returns MetricPromises for
	// given list of metrics, i.e. cpu/memory usage instead of single metric type.
	DownloadMetrics(selectors []ResourceSelector, metricNames []string,
		cachedResources *CachedResources) MetricPromises
	// AggregateMetrics is used to aggregate previously downloaded metrics based on
	// aggregation mode (sum, min, avg). It is used to show cumulative metric graphs on
	// resource list pages.
	AggregateMetrics(metrics MetricPromises, metricName string,
		aggregations dataselect.AggregationModes) MetricPromises

	// Implements IntegrationApp interface
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
	Aggregate dataselect.AggregationMode `json:"aggregation,omitempty"`
}
type DataPoints []DataPoint

type DataPoint struct {
	X int64 `json:"x"`
	Y int64 `json:"y"`
}

// Label stores information about identity of resources (UIDs) described by metric.
type Label map[ResourceKind][]types.UID

// ResourceSelector is a structure used to quickly and uniquely identify given resource.
// This struct can be later used for heapster data download etc.
type ResourceSelector struct {
	// Namespace of this resource.
	Namespace string
	// Type of this resource
	ResourceType ResourceKind
	// Name of this resource.
	ResourceName string
	// Selector used to identify this resource (should be used only for Deployments!).
	Selector map[string]string
	// UID is resource unique identifier.
	UID types.UID
}

// ResourceKind is an unique name for each resource. It can used for API discovery and generic
// code that does things based on the kind. For example, there may be a generic "deleter"
// that based on resource kind, name and namespace deletes it.
type ResourceKind string

type MetricPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     uint64    `json:"value"`
}

/*type Metric struct {
	MetricPoints []MetricPoint `json:"metricPoints"`
	MetricName string `json:"metricName"`
}*/
