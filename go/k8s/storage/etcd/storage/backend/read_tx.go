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

	UnsafeRange(bucket Bucket, key, endKey []byte, limit int64) (keys [][]byte, vals [][]byte)
}

type baseReadTx struct {
	// protects accesses to the txReadBuffer
	sync.RWMutex
	buf txReadBuffer

	// INFO: boltdb transaction 对象
	tx      *bolt.Tx
	buckets map[BucketID]*bolt.Bucket
}

func (baseReadTx *baseReadTx) UnsafeRange(bucketType Bucket, key, endKey []byte, limit int64) (keys [][]byte, vals [][]byte) {

	// INFO: 先查询 txReadBuffer???
	keys, vals := baseReadTx.buf.Range(bucketType, key, endKey, limit)
	if int64(len(keys)) == limit {
		return keys, vals
	}

	// find/cache bucket
	bucket, ok := baseReadTx.buckets[bn]
	// INFO: @see https://github.com/etcd-io/bbolt#range-scans
	cursor := bucket.Cursor()

	k2, v2 := unsafeRange(cursor, key, endKey, limit-int64(len(keys)))
	return append(k2, keys...), append(v2, vals...)
}

type readTx struct {
	baseReadTx
}

type concurrentReadTx struct {
	baseReadTx
}
