package metrics

import (
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"time"

	"k8s.io/component-base/metrics"
)

const (
	// SubsystemSidecar is the default subsystem name in a metrics
	// (= the prefix in the final metrics name). It is to be used
	// by CSI sidecars. Using the same subsystem in different CSI
	// drivers makes it possible to reuse dashboards because
	// the metrics names will be identical. Data from different
	// drivers can be selected via the "driver_name" tag.
	SubsystemSidecar = "csi_sidecar"
	// SubsystemPlugin is what CSI driver's should use as
	// subsystem name.
	SubsystemPlugin = "csi_plugin"

	// Common metric strings
	labelCSIDriverName    = "driver_name"
	labelCSIOperationName = "method_name"
	labelGrpcStatusCode   = "grpc_status_code"
	unknownCSIDriverName  = "unknown-driver"

	// CSI Operation Latency with status code total - Histogram Metric
	operationsLatencyMetricName = "operations_seconds"
	operationsLatencyHelp       = "Container Storage Interface operation duration with gRPC error code status total"
)

var (
	operationsLatencyBuckets = []float64{.1, .25, .5, 1, 2.5, 5, 10, 15, 25, 50, 120, 300, 600}
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

// NewCSIMetricsManager is provided for backwards-compatibility.
var NewCSIMetricsManager = NewCSIMetricsManagerForSidecar

// NewCSIMetricsManagerForSidecar creates and registers metrics for CSI Sidecars and
// returns an object that can be used to trigger the metrics. It uses "csi_sidecar"
// as subsystem.
//
// driverName - Name of the CSI driver against which this operation was executed.
//
//	If unknown, leave empty, and use SetDriverName method to update later.
func NewCSIMetricsManagerForSidecar(driverName string) CSIMetricsManager {
	return NewCSIMetricsManagerWithOptions(driverName)
}

// NewCSIMetricsManagerWithOptions is a customizable constructor, to be used only
// if there are special needs like changing the default subsystems.
//
// driverName - Name of the CSI driver against which this operation was executed.
//
//	If unknown, leave empty, and use SetDriverName method to update later.
func NewCSIMetricsManagerWithOptions(driverName string, options ...MetricsManagerOption) CSIMetricsManager {
	cmm := csiMetricsManager{
		registry:                 metrics.NewKubeRegistry(),
		subsystem:                SubsystemSidecar,
		stabilityLevel:           metrics.ALPHA,
		registerProcessStartTime: true,
	}

	for _, option := range options {
		option(&cmm)
	}

	if cmm.registerProcessStartTime {
		// https://github.com/open-telemetry/opentelemetry-collector/issues/969
		// Add process_start_time_seconds into the metric to let the start time be parsed correctly
		metrics.RegisterProcessStartTime(cmm.registry.Register)
		// INFO: This is a bug in component-base library. We need to remove this after upgrade component-base dependency
		// BugFix: https://github.com/kubernetes/kubernetes/pull/96435
		// The first call to RegisterProcessStartTime can only create the metric, so we need a second call to actually
		// register the metric.
		metrics.RegisterProcessStartTime(cmm.registry.Register)
	}

	labels := []string{labelCSIDriverName, labelCSIOperationName, labelGrpcStatusCode}
	labels = append(labels, cmm.additionalLabelNames...)
	for _, label := range cmm.additionalLabels {
		labels = append(labels, label.name)
	}
	cmm.csiOperationsLatencyMetric = metrics.NewHistogramVec(
		&metrics.HistogramOpts{
			Subsystem:      cmm.subsystem,
			Name:           operationsLatencyMetricName,
			Help:           operationsLatencyHelp,
			Buckets:        operationsLatencyBuckets,
			StabilityLevel: cmm.stabilityLevel,
		},
		labels,
	)

	cmm.SetDriverName(driverName)
	cmm.registerMetrics()

	return &cmm
}

// MetricsManagerOption is used to pass optional configuration to a
// new metrics manager.
type MetricsManagerOption func(*csiMetricsManager)

var _ CSIMetricsManager = &csiMetricsManager{}

type csiMetricsManager struct {
	registry                   metrics.KubeRegistry
	subsystem                  string
	stabilityLevel             metrics.StabilityLevel
	driverName                 string
	additionalLabelNames       []string
	additionalLabels           []label
	csiOperationsLatencyMetric *metrics.HistogramVec
	registerProcessStartTime   bool
}

type label struct {
	name, value string
}

func (cmm *csiMetricsManager) GetRegistry() metrics.KubeRegistry {
	return cmm.registry
}

// RecordMetrics implements CSIMetricsManager.RecordMetrics.
func (cmm *csiMetricsManager) RecordMetrics(
	operationName string,
	operationErr error,
	operationDuration time.Duration) {
	cmm.recordMetricsWithLabels(operationName, operationErr, operationDuration, nil)
}

// recordMetricsWithLabels is the internal implementation of RecordMetrics.
func (cmm *csiMetricsManager) recordMetricsWithLabels(
	operationName string,
	operationErr error,
	operationDuration time.Duration,
	labelValues map[string]string) {
	values := []string{cmm.driverName, operationName, getErrorCode(operationErr)}
	for _, name := range cmm.additionalLabelNames {
		values = append(values, labelValues[name])
	}
	for _, label := range cmm.additionalLabels {
		values = append(values, label.value)
	}
	cmm.csiOperationsLatencyMetric.WithLabelValues(values...).Observe(operationDuration.Seconds())
}

// SetDriverName is called to update the CSI driver name. This should be done
// as soon as possible, otherwise metrics recorded by this manager will be
// recorded with an "unknown-driver" driver_name.
func (cmm *csiMetricsManager) SetDriverName(driverName string) {
	if driverName == "" {
		cmm.driverName = unknownCSIDriverName
	} else {
		cmm.driverName = driverName
	}
}

// RegisterToServer registers an HTTP handler for this metrics manager to the
// given server at the specified address/path.
func (cmm *csiMetricsManager) RegisterToServer(s Server, metricsPath string) {
	s.Handle(metricsPath, metrics.HandlerFor(
		cmm.GetRegistry(),
		metrics.HandlerOpts{
			ErrorHandling: metrics.ContinueOnError}))
}

// WithLabelValues in the base metrics manager creates a fresh wrapper with no labels and let's
// that deal with adding the label values.
func (cmm *csiMetricsManager) WithLabelValues(labels map[string]string) (CSIMetricsManager, error) {
	cmmv := &csiMetricsManagerWithValues{
		csiMetricsManager: cmm,
		additionalValues:  map[string]string{},
	}

	return cmmv.WithLabelValues(labels)
}

func (cmm *csiMetricsManager) registerMetrics() {
	cmm.registry.MustRegister(cmm.csiOperationsLatencyMetric)
}

func (cmm *csiMetricsManager) haveAdditionalLabel(name string) bool {
	for _, n := range cmm.additionalLabelNames {
		if n == name {
			return true
		}
	}
	return false
}

type csiMetricsManagerWithValues struct {
	*csiMetricsManager

	// additionalValues holds the values passed via WithLabelValues.
	additionalValues map[string]string
}

// WithLabelValues in the wrapper creates a wrapper which has all existing labels and
// adds the new ones, with error checking. Can be called multiple times. Each call then
// can add some new value(s). It is an error to overwrite an already set value.
// If RecordMetrics is called before setting all additional values, the missing ones will
// be empty.
func (cmmv *csiMetricsManagerWithValues) WithLabelValues(labels map[string]string) (CSIMetricsManager, error) {
	extended := &csiMetricsManagerWithValues{
		csiMetricsManager: cmmv.csiMetricsManager,
		additionalValues:  map[string]string{},
	}
	// We need to copy the old values to avoid modifying the map in cmmv.
	for name, value := range cmmv.additionalValues {
		extended.additionalValues[name] = value
	}
	// Now add all new values.
	for name, value := range labels {
		if !extended.haveAdditionalLabel(name) {
			return nil, fmt.Errorf("label %q was not defined via WithLabelNames", name)
		}
		if v, ok := extended.additionalValues[name]; ok {
			return nil, fmt.Errorf("label %q already has value %q", name, v)
		}
		extended.additionalValues[name] = value
	}
	return extended, nil
}

// RecordMetrics passes the stored values as to the implementation.
func (cmmv *csiMetricsManagerWithValues) RecordMetrics(
	operationName string,
	operationErr error,
	operationDuration time.Duration) {
	cmmv.recordMetricsWithLabels(operationName, operationErr, operationDuration, cmmv.additionalValues)
}

func getErrorCode(err error) string {
	if err == nil {
		return codes.OK.String()
	}

	st, ok := status.FromError(err)
	if !ok {
		// This is not gRPC error. The operation must have failed before gRPC
		// method was called, otherwise we would get gRPC error.
		return "unknown-non-grpc"
	}

	return st.Code().String()
}
