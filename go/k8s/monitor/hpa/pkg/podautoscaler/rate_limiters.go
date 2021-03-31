package podautoscaler

import (
	"time"

	"k8s.io/client-go/util/workqueue"
)

// FixedItemIntervalRateLimiter limits items to a fixed-rate interval
// 频率固定
type FixedItemIntervalRateLimiter struct {
	interval time.Duration
}

func (r *FixedItemIntervalRateLimiter) When(item interface{}) time.Duration {
	return r.interval
}

// NumRequeues returns back how many failures the item has had
func (r *FixedItemIntervalRateLimiter) NumRequeues(item interface{}) int {
	return 1
}

// Forget indicates that an item is finished being retried.
func (r *FixedItemIntervalRateLimiter) Forget(item interface{}) {
}

func NewFixedItemIntervalRateLimiter(interval time.Duration) workqueue.RateLimiter {
	return &FixedItemIntervalRateLimiter{
		interval: interval,
	}
}

func NewDefaultHPARateLimiter(interval time.Duration) workqueue.RateLimiter {
	return NewFixedItemIntervalRateLimiter(interval)
}
