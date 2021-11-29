package allocator

import "github.com/prometheus/client_golang/prometheus"

var (
	poolCapacity = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "metallb",
		Subsystem: "allocator",
		Name:      "addresses_total",
		Help:      "Number of usable IP addresses, per pool",
	}, []string{
		"pool",
	})

	poolActive = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "metallb",
		Subsystem: "allocator",
		Name:      "addresses_in_use_total",
		Help:      "Number of IP addresses in use, per pool",
	}, []string{
		"pool",
	})

	poolAllocated = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "metallb",
		Subsystem: "allocator",
		Name:      "services_allocated_total",
		Help:      "Number of services allocated, per pool",
	}, []string{
		"pool",
	})
)

func init() {
	prometheus.MustRegister(poolCapacity)
	prometheus.MustRegister(poolActive)
	prometheus.MustRegister(poolAllocated)
}
