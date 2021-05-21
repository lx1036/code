package metrics

import "k8s.io/component-base/metrics"

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
)

// UnschedulablePods returns the pending pods metrics with the label unschedulable
func UnschedulablePods() metrics.GaugeMetric {
	return pendingPods.With(metrics.Labels{"queue": "unschedulable"})
}
