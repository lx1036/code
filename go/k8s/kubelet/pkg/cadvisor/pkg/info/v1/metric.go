package v1

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
