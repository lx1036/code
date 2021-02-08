package metrics

import "github.com/prometheus/client_golang/prometheus"

const (
	// ControllerSubsystem is prometheus subsystem name.
	ControllerSubsystem = "controller"
)

// These variables are defined merely for API compatibility.
var (
	// PersistentVolumeClaimProvisionTotal is used to collect accumulated count of persistent volumes provisioned.
	PersistentVolumeClaimProvisionTotal = M.PersistentVolumeClaimProvisionTotal
	// PersistentVolumeClaimProvisionFailedTotal is used to collect accumulated count of persistent volume provision failed attempts.
	PersistentVolumeClaimProvisionFailedTotal = M.PersistentVolumeClaimProvisionFailedTotal
	// PersistentVolumeClaimProvisionDurationSeconds is used to collect latency in seconds to provision persistent volumes.
	PersistentVolumeClaimProvisionDurationSeconds = M.PersistentVolumeClaimProvisionDurationSeconds
	// PersistentVolumeDeleteTotal is used to collect accumulated count of persistent volumes deleted.
	PersistentVolumeDeleteTotal = M.PersistentVolumeDeleteTotal
	// PersistentVolumeDeleteFailedTotal is used to collect accumulated count of persistent volume delete failed attempts.
	PersistentVolumeDeleteFailedTotal = M.PersistentVolumeDeleteFailedTotal
	// PersistentVolumeDeleteDurationSeconds is used to collect latency in seconds to delete persistent volumes.
	PersistentVolumeDeleteDurationSeconds = M.PersistentVolumeDeleteDurationSeconds
)

// Metrics contains the metrics for a certain subsystem name.
type Metrics struct {
	// PersistentVolumeClaimProvisionTotal is used to collect accumulated count of persistent volumes provisioned.
	PersistentVolumeClaimProvisionTotal *prometheus.CounterVec
	// PersistentVolumeClaimProvisionFailedTotal is used to collect accumulated count of persistent volume provision failed attempts.
	PersistentVolumeClaimProvisionFailedTotal *prometheus.CounterVec
	// PersistentVolumeClaimProvisionDurationSeconds is used to collect latency in seconds to provision persistent volumes.
	PersistentVolumeClaimProvisionDurationSeconds *prometheus.HistogramVec
	// PersistentVolumeDeleteTotal is used to collect accumulated count of persistent volumes deleted.
	PersistentVolumeDeleteTotal *prometheus.CounterVec
	// PersistentVolumeDeleteFailedTotal is used to collect accumulated count of persistent volume delete failed attempts.
	PersistentVolumeDeleteFailedTotal *prometheus.CounterVec
	// PersistentVolumeDeleteDurationSeconds is used to collect latency in seconds to delete persistent volumes.
	PersistentVolumeDeleteDurationSeconds *prometheus.HistogramVec
}

// New creates a new set of metrics with the goven subsystem name.
func New(subsystem string) Metrics {
	return Metrics{
		PersistentVolumeClaimProvisionTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Subsystem: subsystem,
				Name:      "persistentvolumeclaim_provision_total",
				Help:      "Total number of persistent volumes provisioned succesfully. Broken down by storage class name.",
			},
			[]string{"class"},
		),
		PersistentVolumeClaimProvisionFailedTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Subsystem: subsystem,
				Name:      "persistentvolumeclaim_provision_failed_total",
				Help:      "Total number of persistent volume provision failed attempts. Broken down by storage class name.",
			},
			[]string{"class"},
		),
		PersistentVolumeClaimProvisionDurationSeconds: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Subsystem: subsystem,
				Name:      "persistentvolumeclaim_provision_duration_seconds",
				Help:      "Latency in seconds to provision persistent volumes. Failed provisioning attempts are ignored. Broken down by storage class name.",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"class"},
		),
		PersistentVolumeDeleteTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Subsystem: subsystem,
				Name:      "persistentvolume_delete_total",
				Help:      "Total number of persistent volumes deleted succesfully. Broken down by storage class name.",
			},
			[]string{"class"},
		),
		PersistentVolumeDeleteFailedTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Subsystem: subsystem,
				Name:      "persistentvolume_delete_failed_total",
				Help:      "Total number of persistent volume delete failed attempts. Broken down by storage class name.",
			},
			[]string{"class"},
		),
		PersistentVolumeDeleteDurationSeconds: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Subsystem: subsystem,
				Name:      "persistentvolume_delete_duration_seconds",
				Help:      "Latency in seconds to delete persistent volumes. Failed deletion attempts are ignored. Broken down by storage class name.",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"class"},
		),
	}
}

// M contains the metrics with ControllerSubsystem as subsystem name.
var M = New(ControllerSubsystem)
