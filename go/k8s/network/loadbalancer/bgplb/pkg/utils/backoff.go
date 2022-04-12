package utils

import "time"

const (
	backoffMax    = 10 * time.Second
	backoffFactor = 2
)

type Backoff struct {
	nextDelay time.Duration
}

// Duration returns how long to wait before the next retry.
func (b *Backoff) Duration() time.Duration {
	ret := b.nextDelay
	if b.nextDelay == 0 {
		b.nextDelay = time.Second
	} else {
		b.nextDelay *= backoffFactor
		if b.nextDelay > backoffMax {
			b.nextDelay = backoffMax
		}
	}
	return ret
}

func (b *Backoff) Reset() {
	b.nextDelay = 0
}
