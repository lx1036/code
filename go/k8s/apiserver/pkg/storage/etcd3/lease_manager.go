package etcd3

import (
	"context"
	"sync"
	"time"

	"k8s-lx1036/k8s/apiserver/pkg/storage/etcd3/metrics"

	clientv3 "go.etcd.io/etcd/client/v3"
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

// GetLease returns a lease based on requested ttl: if the cached previous
// lease can be reused, reuse it; otherwise request a new one from etcd.
func (l *leaseManager) GetLease(ctx context.Context, ttl int64) (clientv3.LeaseID, error) {
	now := time.Now()
	l.leaseMu.Lock()
	defer l.leaseMu.Unlock()
	// check if previous lease can be reused
	reuseDurationSeconds := l.getReuseDurationSecondsLocked(ttl)
	valid := now.Add(time.Duration(ttl) * time.Second).Before(l.prevLeaseExpirationTime)
	sufficient := now.Add(time.Duration(ttl+reuseDurationSeconds) * time.Second).After(l.prevLeaseExpirationTime)

	// We count all operations that happened in the same lease, regardless of success or failure.
	// Currently each GetLease call only attach 1 object
	l.leaseAttachedObjectCount++

	if valid && sufficient && l.leaseAttachedObjectCount <= l.leaseMaxAttachedObjectCount {
		return l.prevLeaseID, nil
	}

	// request a lease with a little extra ttl from etcd
	ttl += reuseDurationSeconds
	lcr, err := l.client.Lease.Grant(ctx, ttl)
	if err != nil {
		return clientv3.LeaseID(0), err
	}
	// cache the new lease id
	l.prevLeaseID = lcr.ID
	l.prevLeaseExpirationTime = now.Add(time.Duration(ttl) * time.Second)
	// refresh count
	metrics.UpdateLeaseObjectCount(l.leaseAttachedObjectCount)
	l.leaseAttachedObjectCount = 1
	return lcr.ID, nil
}

// getReuseDurationSecondsLocked returns the reusable duration in seconds
// based on the configuration. Lock has to be acquired before calling this
// function.
func (l *leaseManager) getReuseDurationSecondsLocked(ttl int64) int64 {
	reuseDurationSeconds := int64(l.leaseReuseDurationPercent * float64(ttl))
	if reuseDurationSeconds > l.leaseReuseDurationSeconds {
		reuseDurationSeconds = l.leaseReuseDurationSeconds
	}
	return reuseDurationSeconds
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
