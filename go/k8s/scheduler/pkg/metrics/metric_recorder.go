package metrics

import "k8s.io/component-base/metrics"

// MetricRecorder represents a metric recorder which takes action when the
// metric Inc(), Dec() and Clear()
type MetricRecorder interface {
	Inc()
	Dec()
	Clear()
}

// PendingPodsRecorder is an implementation of MetricRecorder
type PendingPodsRecorder struct {
	recorder metrics.GaugeMetric
}

// Inc increases a metric counter by 1, in an atomic way
func (r *PendingPodsRecorder) Inc() {
	r.recorder.Inc()
}

// Dec decreases a metric counter by 1, in an atomic way
func (r *PendingPodsRecorder) Dec() {
	r.recorder.Dec()
}

// Clear set a metric counter to 0, in an atomic way
func (r *PendingPodsRecorder) Clear() {
	r.recorder.Set(float64(0))
}

// NewUnschedulablePodsRecorder returns UnschedulablePods in a Prometheus metric fashion
func NewUnschedulablePodsRecorder() *PendingPodsRecorder {
	return &PendingPodsRecorder{
		recorder: UnschedulablePods(),
	}
}
