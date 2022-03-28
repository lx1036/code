package service

import "time"

const (
	backoffMax    = 60 * time.Minute
	backoffFactor = 2
)

type backoff struct {
	nextDelay time.Duration
}

// Duration returns how long to wait before the next retry.
func (b *backoff) Duration() time.Duration {
	ret := b.nextDelay
	if b.nextDelay == 0 {
		b.nextDelay = time.Minute
	} else {
		b.nextDelay *= backoffFactor
		if b.nextDelay > backoffMax {
			b.nextDelay = backoffMax
		}
	}
	return ret
}

func (b *backoff) Reset() {
	b.nextDelay = 0
}
