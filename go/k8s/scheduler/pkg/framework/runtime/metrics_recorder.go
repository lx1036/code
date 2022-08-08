package runtime

import (
	"time"

	k8smetrics "k8s.io/component-base/metrics"
)

// frameworkMetric is the data structure passed in the buffer channel between the main framework thread
// and the metricsRecorder goroutine.
type frameworkMetric struct {
	metric      *k8smetrics.HistogramVec
	labelValues []string
	value       float64
}

// metricRecorder records framework metrics in a separate goroutine to avoid overhead in the critical path.
type metricsRecorder struct {
	// bufferCh is a channel that serves as a metrics buffer before the metricsRecorder goroutine reports it.
	bufferCh chan *frameworkMetric
	// if bufferSize is reached, incoming metrics will be discarded.
	bufferSize int
	// how often the recorder runs to flush the metrics.
	interval time.Duration

	// stopCh is used to stop the goroutine which periodically flushes metrics. It's currently only
	// used in tests.
	stopCh chan struct{}
	// isStoppedCh indicates whether the goroutine is stopped. It's used in tests only to make sure
	// the metric flushing goroutine is stopped so that tests can collect metrics for verification.
	isStoppedCh chan struct{}
}

// run flushes buffered metrics into Prometheus every second.
func (r *metricsRecorder) run() {
	for {
		select {
		case <-r.stopCh:
			close(r.isStoppedCh)
			return
		default:
		}
		r.flushMetrics()
		time.Sleep(r.interval)
	}
}

// flushMetrics tries to clean up the bufferCh by reading at most bufferSize metrics.
func (r *metricsRecorder) flushMetrics() {
	for i := 0; i < r.bufferSize; i++ {
		select {
		case m := <-r.bufferCh:
			m.metric.WithLabelValues(m.labelValues...).Observe(m.value)
		default:
			return
		}
	}
}

func newMetricsRecorder(bufferSize int, interval time.Duration) *metricsRecorder {
	recorder := &metricsRecorder{
		bufferCh:    make(chan *frameworkMetric, bufferSize),
		bufferSize:  bufferSize,
		interval:    interval,
		stopCh:      make(chan struct{}),
		isStoppedCh: make(chan struct{}),
	}
	go recorder.run()

	return recorder
}
