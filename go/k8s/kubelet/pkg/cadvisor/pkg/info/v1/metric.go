package v1

import "time"

// Type of metric being exported.
type MetricType string

const (
	// Instantaneous value. May increase or decrease.
	MetricGauge MetricType = "gauge"

	// A counter-like value that is only expected to increase.
	MetricCumulative MetricType = "cumulative"
)

// DataType for metric being exported.
type DataType string

const (
	IntType   DataType = "int"
	FloatType DataType = "float"
)

// Spec for custom metric.
type MetricSpec struct {
	// The name of the metric.
	Name string `json:"name"`

	// Type of the metric.
	Type MetricType `json:"type"`

	// Data Type for the stats.
	Format DataType `json:"format"`

	// Display Units for the stats.
	Units string `json:"units"`
}

// An exported metric.
type MetricVal struct {
	// Label associated with a metric
	Label  string            `json:"label,omitempty"`
	Labels map[string]string `json:"labels,omitempty"`

	// Time at which the metric was queried
	Timestamp time.Time `json:"timestamp"`

	// The value of the metric at this point.
	IntValue   int64   `json:"int_value,omitempty"`
	FloatValue float64 `json:"float_value,omitempty"`
}
