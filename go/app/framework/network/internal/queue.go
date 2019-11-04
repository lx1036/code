package internal

import "sync"

// AsyncJobQueue queues pending tasks.
type AsyncJobQueue struct {
	mu   sync.Locker
	jobs []func() error
}

