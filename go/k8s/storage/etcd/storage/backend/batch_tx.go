package backend

import (
	"fmt"
	"k8s.io/klog/v2"
	"sync"

	bolt "go.etcd.io/bbolt"
)

type BucketID int

type Bucket interface {
	Name() []byte
}

type BatchTx interface {
	ReadTx
}

type batchTx struct {
	sync.Mutex
	tx      *bolt.Tx
	backend *backend

	pending int
}

// INFO: 所有批量写事务必须先获得锁
func (t *batchTx) Lock() {
	t.Lock()
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

type batchTxBuffered struct {
	batchTx
	buf txWriteBuffer
}
