package lock

import "sync"

type internalRWMutex struct {
	sync.RWMutex
}
