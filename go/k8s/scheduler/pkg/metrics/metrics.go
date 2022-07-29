package metrics

import (
	"sync"

	"k8s.io/component-base/metrics"
	"k8s.io/component-base/metrics/legacyregistry"
)

const (
	// SchedulerSubsystem - subsystem name used by scheduler
	SchedulerSubsystem = "scheduler"
)

var (
	pendingPods = metrics.NewGaugeVec(
		&metrics.GaugeOpts{
			Subsystem: SchedulerSubsystem,
			Name:      "pending_pods",
			Help: `Number of pending pods, by the queue type. 'active' means number of pods in activeQ;
'backoff' means number of pods in backoffQ; 'unschedulable' means number of pods in unschedulableQ.`,
			StabilityLevel: metrics.ALPHA,
		}, []string{"queue"})

	SchedulerQueueIncomingPods = metrics.NewCounterVec(
		&metrics.CounterOpts{
			Subsystem:      SchedulerSubsystem,
			Name:           "queue_incoming_pods_total",
			Help:           "Number of pods added to scheduling queues by event and queue type.",
			StabilityLevel: metrics.STABLE,
		}, []string{"queue", "event"})

	unschedulableReasons = metrics.NewGaugeVec(
		&metrics.GaugeOpts{
			Subsystem:      SchedulerSubsystem,
			Name:           "unschedulable_pods",
			Help:           "The number of unschedulable pods broken down by plugin name. A pod will increment the gauge for all plugins that caused it to not schedule and so this metric have meaning only when broken down by plugin.",
			StabilityLevel: metrics.ALPHA,
		}, []string{"plugin", "profile"})
)

// UnschedulablePods returns the pending pods metrics with the label unschedulable
func UnschedulablePods() metrics.GaugeMetric {
	return pendingPods.With(metrics.Labels{"queue": "unschedulable"})
}

func UnschedulableReason(plugin string, profile string) metrics.GaugeMetric {
	return unschedulableReasons.With(metrics.Labels{"plugin": plugin, "profile": profile})
}

var registerMetrics sync.Once

var (
	scheduleAttempts = metrics.NewCounterVec(
		&metrics.CounterOpts{
			Subsystem:      SchedulerSubsystem,
			Name:           "schedule_attempts_total",
			Help:           "Number of attempts to schedule pods, by the result. 'unschedulable' means a pod could not be scheduled, while 'error' means an internal scheduler problem.",
			StabilityLevel: metrics.ALPHA,
		}, []string{"result", "profile"})

	metricsList = []metrics.Registerable{
		scheduleAttempts,
	}
)

func RegisterMetrics(extraMetrics ...metrics.Registerable) {
	for _, metric := range extraMetrics {
		legacyregistry.MustRegister(metric)
	}
}

// Register all metrics.
func Register() {
	// Register the metrics.
	registerMetrics.Do(func() {
		RegisterMetrics(metricsList...)
		//volumeschedulingmetrics.RegisterVolumeSchedulingMetrics()
	})
}
