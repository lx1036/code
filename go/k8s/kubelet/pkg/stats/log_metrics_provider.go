package stats

import "k8s.io/kubernetes/pkg/volume"

// LogMetricsService defines an interface for providing LogMetrics functionality.
type LogMetricsService interface {
	createLogMetricsProvider(path string) volume.MetricsProvider
}

type logMetrics struct{}

// NewLogMetricsService returns a new LogMetricsService type struct.
func NewLogMetricsService() LogMetricsService {
	return logMetrics{}
}

func (l logMetrics) createLogMetricsProvider(path string) volume.MetricsProvider {
	return volume.NewMetricsDu(path)
}
