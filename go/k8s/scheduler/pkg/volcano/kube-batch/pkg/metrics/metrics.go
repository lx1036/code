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
	// 一个调度周期所花费的时间，包含多个plugin和action花费的时间总和
	e2eSchedulingLatency = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Subsystem: VolcanoNamespace,
			Name:      "e2e_scheduling_latency_milliseconds",
			Help:      "E2e scheduling latency in milliseconds (scheduling algorithm + binding)",
			Buckets:   prometheus.ExponentialBuckets(5, 2, 10),
		},
	)

	// 一个 action 花费的时间
	actionSchedulingLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: VolcanoNamespace,
			Name:      "action_scheduling_latency_microseconds",
			Help:      "Action scheduling latency in microseconds",
			Buckets:   prometheus.ExponentialBuckets(5, 2, 10),
		}, []string{"action"},
	)

	// 一个 plugin 花费的时间
	pluginSchedulingLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: VolcanoNamespace,
			Name:      "plugin_scheduling_latency_microseconds",
			Help:      "Plugin scheduling latency in microseconds",
			Buckets:   prometheus.ExponentialBuckets(5, 2, 10),
		}, []string{"plugin", "OnSession"},
	)
)

// Duration get the time since specified start
func Duration(start time.Time) time.Duration {
	return time.Since(start)
}

func DurationInMicroseconds(duration time.Duration) float64 {
	return float64(duration.Microseconds())
}

func DurationInMilliseconds(duration time.Duration) float64 {
	return float64(duration.Milliseconds())
}

// 一个调度周期的总时间
func UpdateE2eDuration(duration time.Duration) {
	e2eSchedulingLatency.Observe(DurationInMilliseconds(duration))
}

func UpdateActionDuration(actionName string, duration time.Duration) {
	actionSchedulingLatency.WithLabelValues(actionName).Observe(DurationInMicroseconds(duration))
}

func UpdatePluginDuration(pluginName, onSessionStatus string, duration time.Duration) {
	pluginSchedulingLatency.WithLabelValues(pluginName, onSessionStatus).Observe(DurationInMicroseconds(duration))
}
