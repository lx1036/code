package controller

import (
	"math/rand"
	"sync"
	"time"

	"k8s.io/client-go/util/workqueue"
)

type rateLimiterWithJitter struct {
	workqueue.RateLimiter
	baseDelay time.Duration
	rd        *rand.Rand
	mutex     sync.Mutex
}

func (r *rateLimiterWithJitter) When(item interface{}) time.Duration {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	delay := r.RateLimiter.When(item).Nanoseconds()
	percentage := r.rd.Float64()
	jitter := int64(float64(r.baseDelay.Nanoseconds()) * percentage)
	if jitter > delay {
		return 0
	}
	return time.Duration(delay - jitter)
}

func newItemExponentialFailureRateLimiterWithJitter(baseDelay time.Duration, maxDelay time.Duration) workqueue.RateLimiter {
	return &rateLimiterWithJitter{
		RateLimiter: workqueue.NewItemExponentialFailureRateLimiter(baseDelay, maxDelay),
		baseDelay:   baseDelay,
		rd:          rand.New(rand.NewSource(time.Now().UTC().UnixNano())),
	}
}
