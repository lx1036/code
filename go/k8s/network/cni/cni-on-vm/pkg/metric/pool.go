package metric

import "github.com/prometheus/client_golang/prometheus"

var (
	// ResourcePoolIdle terway amount of idle resource in the pool
	ResourcePoolIdle = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "resource_pool_idle_count",
			Help: "amount of idle resources in the pool",
		},
		[]string{"name", "type", "capacity", "max_idle", "min_idle"},
	)

	// ResourcePoolTotal terway total source amount in the pool
	ResourcePoolTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "resource_pool_total_count",
			Help: "total resources amount in the pool",
		},
		// not accessory to put capacity, max_idle or min_idle into labels ?
		[]string{"name", "type", "capacity", "max_idle", "min_idle"},
	)
)
