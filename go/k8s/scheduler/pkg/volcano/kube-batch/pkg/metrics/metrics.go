package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"time"
)

const (
	// VolcanoNamespace - namespace in prometheus used by volcano
	VolcanoNamespace = "volcano"
)

var (
	actionSchedulingLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: VolcanoNamespace,
			Name:      "action_scheduling_latency_microseconds",
			Help:      "Action scheduling latency in microseconds",
			Buckets:   prometheus.ExponentialBuckets(5, 2, 10),
		}, []string{"action"},
	)
)

// Duration get the time since specified start
func Duration(start time.Time) time.Duration {
	return time.Since(start)
}

func UpdateActionDuration(actionName string, duration time.Duration) {
	actionSchedulingLatency.WithLabelValues(actionName).Observe(DurationInMicroseconds(duration))
}
