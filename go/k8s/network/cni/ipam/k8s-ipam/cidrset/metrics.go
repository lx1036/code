package cidrset

import "k8s.io/component-base/metrics"

const nodeIpamSubsystem = "node_ipam_controller"

var (
	cidrSetAllocations = metrics.NewCounterVec(
		&metrics.CounterOpts{
			Subsystem:      nodeIpamSubsystem,
			Name:           "cidrset_cidrs_allocations_total",
			Help:           "Counter measuring total number of CIDR allocations.",
			StabilityLevel: metrics.ALPHA,
		},
		[]string{"clusterCIDR"},
	)

	cidrSetAllocationTriesPerRequest = metrics.NewHistogramVec(
		&metrics.HistogramOpts{
			Subsystem:      nodeIpamSubsystem,
			Name:           "cidrset_allocation_tries_per_request",
			Help:           "Number of endpoints added on each Service sync",
			StabilityLevel: metrics.ALPHA,
			Buckets:        metrics.ExponentialBuckets(1, 5, 5),
		},
		[]string{"clusterCIDR"},
	)

	cidrSetUsage = metrics.NewGaugeVec(
		&metrics.GaugeOpts{
			Subsystem:      nodeIpamSubsystem,
			Name:           "cidrset_usage_cidrs",
			Help:           "Gauge measuring percentage of allocated CIDRs.",
			StabilityLevel: metrics.ALPHA,
		},
		[]string{"clusterCIDR"},
	)

	cidrSetReleases = metrics.NewCounterVec(
		&metrics.CounterOpts{
			Subsystem:      nodeIpamSubsystem,
			Name:           "cidrset_cidrs_releases_total",
			Help:           "Counter measuring total number of CIDR releases.",
			StabilityLevel: metrics.ALPHA,
		},
		[]string{"clusterCIDR"},
	)
)
