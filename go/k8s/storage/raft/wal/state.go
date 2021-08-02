package wal

import "sync"

type RaftState struct {
	sync.RWMutex
}
