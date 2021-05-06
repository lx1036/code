package etcd3

import (
	"sync"
	"time"

	"go.etcd.io/etcd/clientv3"
)

const (
	defaultLeaseReuseDurationSeconds = 60
	defaultLeaseMaxObjectCount       = 1000
)

// LeaseManagerConfig is configuration for creating a lease manager.
type LeaseManagerConfig struct {
	// ReuseDurationSeconds specifies time in seconds that each lease is reused
	ReuseDurationSeconds int64
	// MaxObjectCount specifies how many objects that a lease can attach
	MaxObjectCount int64
}

// leaseManager is used to manage leases requested from etcd. If a new write
// needs a lease that has similar expiration time to the previous one, the old
// lease will be reused to reduce the overhead of etcd, since lease operations
// are expensive. In the implementation, we only store one previous lease,
// since all the events have the same ttl.
type leaseManager struct {
	client                  *clientv3.Client // etcd client used to grant leases
	leaseMu                 sync.Mutex
	prevLeaseID             clientv3.LeaseID
	prevLeaseExpirationTime time.Time
	// The period of time in seconds and percent of TTL that each lease is
	// reused. The minimum of them is used to avoid unreasonably large
	// numbers. We use var instead of const for testing purposes.
	leaseReuseDurationSeconds   int64
	leaseReuseDurationPercent   float64
	leaseMaxAttachedObjectCount int64
	leaseAttachedObjectCount    int64
}

// newDefaultLeaseManager creates a new lease manager using default setting.
func newDefaultLeaseManager(client *clientv3.Client, config LeaseManagerConfig) *leaseManager {
	if config.MaxObjectCount <= 0 {
		config.MaxObjectCount = defaultLeaseMaxObjectCount
	}
	return newLeaseManager(client, config.ReuseDurationSeconds, 0.05, config.MaxObjectCount)
}

// newLeaseManager creates a new lease manager with the number of buffered
// leases, lease reuse duration in seconds and percentage. The percentage
// value x means x*100%.
func newLeaseManager(client *clientv3.Client, leaseReuseDurationSeconds int64,
	leaseReuseDurationPercent float64, maxObjectCount int64) *leaseManager {
	return &leaseManager{
		client:                      client,
		leaseReuseDurationSeconds:   leaseReuseDurationSeconds,
		leaseReuseDurationPercent:   leaseReuseDurationPercent,
		leaseMaxAttachedObjectCount: maxObjectCount,
	}
}
