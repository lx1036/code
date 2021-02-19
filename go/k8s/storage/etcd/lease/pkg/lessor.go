package pkg

import (
	"math"
	"sync"
	"time"

	store "go.etcd.io/etcd/mvcc/backend"
)

var (
	// maximum number of leases to revoke per second; configurable for tests
	// lease过期淘汰会默认限速每秒1000个
	leaseRevokeRate = 1000

	// the default interval to check if the expired lease is revoked
	defaultExpiredleaseRetryInterval = 3 * time.Second
)

type LeaseID int64

type LeaseItem struct {
	Key string
}

type Lessor interface {
}

type Lease struct {
	ID           LeaseID
	ttl          int64 // time to live of the lease in seconds
	remainingTTL int64 // remaining time to live in seconds, if zero valued it is considered unset and the full ttl should be used
	// expiryMu protects concurrent accesses to expiry
	expiryMu sync.RWMutex

	// expiry is time when lease should expire. no expiration when expiry.IsZero() is true
	expiry time.Time

	// mu protects concurrent accesses to itemSet
	mu      sync.RWMutex
	itemSet map[LeaseItem]struct{}
	revokeC chan struct{}
}

type LessorConfig struct {
	MinLeaseTTL                int64
	CheckpointInterval         time.Duration
	ExpiredLeasesRetryInterval time.Duration
}

type lessor struct {
	mu sync.RWMutex

	leaseMap             map[LeaseID]*Lease
	leaseExpiredNotifier *LeaseExpiredNotifier
	leaseCheckpointHeap  LeaseQueue
	itemMap              map[LeaseItem]LeaseID

	backend store.Backend

	// minLeaseTTL is the minimum lease TTL that can be granted for a lease.
	minLeaseTTL int64

	// Wait duration between lease checkpoints.
	checkpointInterval time.Duration
	// the interval to check if the expired lease is revoked
	expiredLeaseRetryInterval time.Duration

	expiredC chan []*Lease
	// stopC is a channel whose closure indicates that the lessor should be stopped.
	stopC chan struct{}
	// doneC is a channel whose closure indicates that the lessor is stopped.
	doneC chan struct{}

	// demotec is set when the lessor is the primary.
	// demotec will be closed if the lessor is demoted.
	// lessor 降级
	demotec chan struct{}
}

// isPrimary indicates if this lessor is the primary lessor. The primary
// lessor manages lease expiration and renew.
func (le *lessor) isPrimary() bool {
	return le.demotec != nil
}

// expireExists returns true if expiry items exist.
// It pops only when expiry item exists.
// "next" is true, to indicate that it may exist in next attempt.
func (le *lessor) expireExists() (l *Lease, ok bool, next bool) {
	if le.leaseExpiredNotifier.Len() == 0 {
		return nil, false, false
	}

	item := le.leaseExpiredNotifier.Poll()
	l = le.leaseMap[item.id]
	if l == nil {
		// lease has expired or been revoked
		// no need to revoke (nothing is expiry)
		le.leaseExpiredNotifier.Unregister() // O(log N)
		return nil, false, true
	}

	now := time.Now()
	if now.UnixNano() < item.time /* expiration time */ {
		// Candidate expirations are caught up, reinsert this item
		// and no need to revoke (nothing is expiry)
		return l, false, false
	}

	// ??? expiredLeaseRetryInterval

	// recheck if revoke is complete after retry interval
	item.time = now.Add(le.expiredLeaseRetryInterval).UnixNano()
	le.leaseExpiredNotifier.RegisterOrUpdate(item)
	return l, true, false
}

func (l *Lease) expired() bool {
	return l.Remaining() <= 0
}

// Remaining returns the remaining time of the lease.
func (l *Lease) Remaining() time.Duration {
	l.expiryMu.RLock()
	defer l.expiryMu.RUnlock()
	if l.expiry.IsZero() { // expiry为0表示永不过期
		return time.Duration(math.MaxInt64)
	}

	return time.Until(l.expiry)
}

// findExpiredLeases loops leases in the leaseMap until reaching expired limit
// and returns the expired leases that needed to be revoked.
func (le *lessor) findExpiredLeases(limit int) []*Lease {
	leases := make([]*Lease, 0, 16)

	for {
		l, ok, next := le.expireExists()
		if !ok && !next { // 最小堆首元素没有过期
			break
		}

		if !ok {
			continue
		}
		if next {
			continue
		}

		if l.expired() {
			leases = append(leases, l)

			// reach expired limit
			if len(leases) == limit {
				break
			}
		}
	}

	return leases
}

// revokeExpiredLeases finds all leases past their expiry and sends them to expired channel for to be revoked.
func (le *lessor) revokeExpiredLeases() {
	var leases []*Lease

	// rate limit
	// lease过期淘汰会默认限速每秒1000个
	revokeLimit := leaseRevokeRate / 2

	le.mu.RLock()
	if le.isPrimary() {
		leases = le.findExpiredLeases(revokeLimit)
	}
	le.mu.RUnlock()

	if len(leases) != 0 {
		select {
		case <-le.stopC:
			return
		case le.expiredC <- leases:
		default:
			// the receiver of expiredC is probably busy handling
			// other stuff
			// let's try this next time after 500ms
		}
	}
}

func (le *lessor) ExpiredLeasesC() <-chan []*Lease {
	return le.expiredC
}

func (le *lessor) runLoop() {
	defer close(le.doneC)

	for {
		le.revokeExpiredLeases()

		select {
		case <-time.After(500 * time.Millisecond):
		case <-le.stopC:
			return
		}
	}
}

func (le *lessor) Stop() {
	close(le.stopC)
	<-le.doneC
}

func NewLessor(backend store.Backend, cfg LessorConfig) *lessor {
	expiredLeaseRetryInterval := cfg.ExpiredLeasesRetryInterval
	if expiredLeaseRetryInterval == 0 {
		expiredLeaseRetryInterval = defaultExpiredleaseRetryInterval
	}

	l := &lessor{
		leaseMap:             make(map[LeaseID]*Lease),
		itemMap:              make(map[LeaseItem]LeaseID),
		leaseExpiredNotifier: newLeaseExpiredNotifier(),
		leaseCheckpointHeap:  make(LeaseQueue, 0),
		backend:              backend,
		minLeaseTTL:          cfg.MinLeaseTTL,
		//checkpointInterval:        checkpointInterval,
		expiredLeaseRetryInterval: expiredLeaseRetryInterval,
		// expiredC is a small buffered chan to avoid unnecessary blocking.
		expiredC: make(chan []*Lease, 16),
		stopC:    make(chan struct{}),
		doneC:    make(chan struct{}),
	}

	l.initAndRecover()

	go l.runLoop()

	return l
}
