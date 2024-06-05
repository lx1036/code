package lock

import "sync"

type internalRWMutex struct {
    sync.RWMutex
}

type internalMutex struct {
    sync.Mutex
}
