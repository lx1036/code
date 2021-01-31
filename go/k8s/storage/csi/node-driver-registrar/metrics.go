package main

import (
	"net/http"
	"time"

	"k8s.io/component-base/metrics"
)

// Server represents any type that could serve HTTP requests for the metrics
// endpoint.
type Server interface {
	Handle(pattern string, handler http.Handler)
}

// CSIMetricsManager exposes functions for recording metrics for CSI operations.
type CSIMetricsManager interface {
	// GetRegistry() returns the metrics.KubeRegistry used by this metrics manager.
	GetRegistry() metrics.KubeRegistry

	// RecordMetrics must be called upon CSI Operation completion to record
	// the operation's metric.
	// operationName - Name of the CSI operation.
	// operationErr - Error, if any, that resulted from execution of operation.
	// operationDuration - time it took for the operation to complete
	//
	// If WithLabelNames was used to define additional labels when constructing
	// the manager, then WithLabelValues should be used to create a wrapper which
	// holds the corresponding values before calling RecordMetrics of the wrapper.
	// Labels with missing values are recorded as empty.
	RecordMetrics(
		operationName string,
		operationErr error,
		operationDuration time.Duration)

	// WithLabelValues must be used to add the additional label
	// values defined via WithLabelNames. When calling RecordMetrics
	// without it or with too few values, the missing values are
	// recorded as empty. WithLabelValues can be called multiple times
	// and then accumulates values.
	WithLabelValues(labels map[string]string) (CSIMetricsManager, error)

	// SetDriverName is called to update the CSI driver name. This should be done
	// as soon as possible, otherwise metrics recorded by this manager will be
	// recorded with an "unknown-driver" driver_name.
	// driverName - Name of the CSI driver against which this operation was executed.
	SetDriverName(driverName string)

	// RegisterToServer registers an HTTP handler for this metrics manager to the
	// given server at the specified address/path.
	RegisterToServer(s Server, metricsPath string)
}
