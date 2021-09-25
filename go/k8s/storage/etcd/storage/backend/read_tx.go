package backend

import (
	"math"
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

	// txWg protects tx from being rolled back at the end of a batch interval until all reads using this tx are done.
	txWg *sync.WaitGroup
}

// UnsafeRange INFO: 先 range scan 下 buffer，然后 boltdb
func (baseReadTx *baseReadTx) UnsafeRange(bucketType Bucket, key, endKey []byte, limit int64) (keys [][]byte, vals [][]byte) {
	if endKey == nil {
		// forbid duplicates for single keys
		limit = 1
	}
	if limit <= 0 {
		limit = math.MaxInt64
	}
	if limit > 1 && !bucketType.IsSafeRangeBucket() {
		panic("do not use unsafeRange on non-keys bucket")
	}

	// INFO: 先查询 txReadBuffer 中查找 [key, endKey] 之间的 (key, value)
	keys, vals = baseReadTx.buf.Range(bucketType, key, endKey, limit)
	if int64(len(keys)) == limit {
		return keys, vals
	}

	// TODO: 这里到处加锁???

	// find/cache bucket
	bucketID := bucketType.ID()
	baseReadTx.RLock() // only-read lock
	bucket, ok := baseReadTx.buckets[bucketID]
	baseReadTx.RUnlock()
	lockHeld := false
	if !ok {
		baseReadTx.Lock()
		lockHeld = true
		bucket = baseReadTx.tx.Bucket(bucketType.Name()) // INFO: 如果不存在返回nil
		baseReadTx.buckets[bucketID] = bucket
	}
	if bucket == nil { // ignore missing bucket since may have been created in this batch
		if lockHeld {
			baseReadTx.Unlock()
		}
		return keys, vals
	}
	if !lockHeld {
		baseReadTx.Lock()
	}

	// INFO: @see https://github.com/etcd-io/bbolt#range-scans
	cursor := bucket.Cursor()
	baseReadTx.Unlock()

	k2, v2 := unsafeRange(cursor, key, endKey, limit-int64(len(keys)))
	return append(k2, keys...), append(v2, vals...)
}

type readTx struct {
	baseReadTx
}

func (rt *readTx) reset() {
	rt.buf.reset()
	rt.buckets = make(map[BucketID]*bolt.Bucket)
	rt.tx = nil
	rt.txWg = new(sync.WaitGroup)
}

type concurrentReadTx struct {
	baseReadTx
}
