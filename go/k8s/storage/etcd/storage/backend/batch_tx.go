package backend

import (
	"fmt"
	"k8s.io/klog/v2"
	"sync"
	"sync/atomic"

	bolt "go.etcd.io/bbolt"
)

type BucketID int

type Bucket interface {
	ID() BucketID
	Name() []byte
	String() string
	IsSafeRangeBucket() bool
}

type BatchTx interface {
	ReadTx

	UnsafeCreateBucket(bucket Bucket)
	UnsafeDeleteBucket(bucket Bucket)
	UnsafePut(bucket Bucket, key []byte, value []byte)
	UnsafeSeqPut(bucket Bucket, key []byte, value []byte)
	UnsafeDelete(bucket Bucket, key []byte)

	// Commit commits a previous tx and begins a new writable one.
	Commit()
	// CommitAndStop commits the previous tx and does not create a new one.
	CommitAndStop()
}

type batchTx struct {
	sync.Mutex
	tx      *bolt.Tx
	backend *backend

	pending int
}

// INFO: 所有批量写事务必须先获得锁
func (t *batchTx) Unlock() {
	if t.pending >= t.backend.batchLimit {
		t.commit(false)
	}

	t.Unlock()
}

func (t *batchTx) UnsafeCreateBucket(bucket Bucket) {
	_, err := t.tx.CreateBucket(bucket.Name())
	if err != nil && err != bolt.ErrBucketExists {
		klog.Fatal(fmt.Sprintf("[UnsafeCreateBucket]fail to create bucket %s", bucket.Name()))
	}

	t.pending++
}

func (t *batchTx) UnsafePut(bucket Bucket, key []byte, value []byte) {
	t.unsafePut(bucket, key, value, false)
}

func (t *batchTx) UnsafeSeqPut(bucket Bucket, key []byte, value []byte) {
	t.unsafePut(bucket, key, value, true)
}

func (t *batchTx) unsafePut(bucketType Bucket, key []byte, value []byte, seq bool) {
	bucket := t.tx.Bucket(bucketType.Name())
	if bucket == nil {
		klog.Fatalf(fmt.Sprintf("[unsafePut]bucket %s in boltdb is not existed", bucketType.Name()))
	}
	// INFO: 这里也是一个小重点
	if seq {
		// it is useful to increase fill percent when the workloads are mostly append-only.
		// this can delay the page split and reduce space usage.
		bucket.FillPercent = 0.9
	}

	// 这里的 key 是 revision
	err := bucket.Put(key, value)
	if err != nil {
		klog.Errorf(fmt.Sprintf("[unsafePut]write key %s into boltdb err %v", string(key), err))
		return
	}

	// TODO: pending 含义是啥???
	t.pending++
}

func (t *batchTx) commit(stop bool) {
	// commit the last tx
	if t.tx != nil {
		if t.pending == 0 && !stop {
			return
		}

		err := t.tx.Commit()
		atomic.AddInt64(&t.backend.commits, 1)
		t.pending = 0
		if err != nil {
			klog.Errorf(fmt.Sprintf("[batchTx commit]err %v", err))
		}
	}

	if !stop {
		// read-write 读写事务
		t.tx = t.backend.begin(true)
	}
}

type batchTxBuffered struct {
	batchTx
	buf txWriteBuffer
}

func newBatchTxBuffered(backend *backend) *batchTxBuffered {
	tx := &batchTxBuffered{
		batchTx: batchTx{
			backend: backend,
		},
		buf: txWriteBuffer{
			txBuffer: txBuffer{
				buckets: make(map[BucketID]*bucketBuffer),
			},
			bucket2seq: make(map[BucketID]bool),
		},
	}

	tx.Commit()

	return tx
}

func (t *batchTxBuffered) Commit() {
	t.Lock()
	t.commit(false)
	t.Unlock()
}

func (t *batchTxBuffered) commit(stop bool) {
	if t.backend.hooks != nil {
		t.backend.hooks.OnPreCommitUnsafe(t)
	}

	// all read txs must be closed to acquire boltdb commit rwlock
	t.backend.readTx.Lock()
	t.unsafeCommit(stop)
	t.backend.readTx.Unlock()
}

func (t *batchTxBuffered) unsafeCommit(stop bool) {

	t.batchTx.commit(stop)

	if !stop {
		// only-read transaction
		t.backend.readTx.tx = t.backend.begin(false)
	}
}
