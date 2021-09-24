package backend

import (
	"sync"

	bolt "go.etcd.io/bbolt"
)

type ReadTx interface {
	Lock()
	Unlock()
	RLock()
	RUnlock()
}

type baseReadTx struct {
	// protects accesses to the txReadBuffer
	sync.RWMutex
	buf txReadBuffer

	// INFO: boltdb transaction 对象
	tx      *bolt.Tx
	buckets map[BucketID]*bolt.Bucket
}

type readTx struct {
	baseReadTx
}

type concurrentReadTx struct {
	baseReadTx
}
