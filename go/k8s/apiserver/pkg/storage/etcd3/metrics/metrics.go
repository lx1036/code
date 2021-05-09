package metrics

import (
	"sync"
	"time"

	compbasemetrics "k8s.io/component-base/metrics"
	"k8s.io/component-base/metrics/legacyregistry"
)

var (
	etcdLeaseObjectCounts = compbasemetrics.NewHistogramVec(
		&compbasemetrics.HistogramOpts{
			Name:           "etcd_lease_object_counts",
			Help:           "Number of objects attached to a single etcd lease.",
			Buckets:        []float64{10, 50, 100, 500, 1000, 2500, 5000},
			StabilityLevel: compbasemetrics.ALPHA,
		},
		[]string{},
	)
	etcdRequestLatency = compbasemetrics.NewHistogramVec(
		&compbasemetrics.HistogramOpts{
			Name:           "etcd_request_duration_seconds",
			Help:           "Etcd request latency in seconds for each operation and object type.",
			StabilityLevel: compbasemetrics.ALPHA,
		},
		[]string{"operation", "type"},
	)
)

var registerMetrics sync.Once

func init() {
	registerMetrics.Do(func() {
		legacyregistry.MustRegister(etcdRequestLatency)
		legacyregistry.MustRegister(etcdLeaseObjectCounts)

	})
}

// UpdateLeaseObjectCount sets the etcd_lease_object_counts metric.
func UpdateLeaseObjectCount(count int64) {
	// Currently we only store one previous lease, since all the events have the same ttl.
	// See pkg/storage/etcd3/lease_manager.go
	etcdLeaseObjectCounts.WithLabelValues().Observe(float64(count))
}

// RecordEtcdRequestLatency sets the etcd_request_duration_seconds metrics.
func RecordEtcdRequestLatency(verb, resource string, startTime time.Time) {
	etcdRequestLatency.WithLabelValues(verb, resource).Observe(sinceInSeconds(startTime))
}

// sinceInSeconds gets the time since the specified start in seconds.
func sinceInSeconds(start time.Time) float64 {
	return time.Since(start).Seconds()
}
