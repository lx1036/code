package backend

import (
	"bytes"
	"fmt"
	"k8s.io/klog/v2"
	"math"
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

// BatchTx INFO: BatchTx 包含读事务 ReadTx，很重要
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

// BatchTx interface embeds ReadTx interface. But RLock() and RUnlock() do not
// have appropriate semantics in BatchTx interface. Therefore should not be called.
// TODO: might want to decouple ReadTx and BatchTx

func (t *batchTx) RLock() {
	panic("unexpected RLock")
}

func (t *batchTx) RUnlock() {
	panic("unexpected RUnlock")
}

func (t *batchTx) UnsafeCreateBucket(bucket Bucket) {
	_, err := t.tx.CreateBucket(bucket.Name())
	if err != nil && err != bolt.ErrBucketExists {
		klog.Fatal(fmt.Sprintf("[UnsafeCreateBucket]fail to create bucket %s", bucket.Name()))
	}

	t.pending++
}

func (t *batchTx) UnsafeDeleteBucket(bucket Bucket) {
	err := t.tx.DeleteBucket(bucket.Name())
	if err != nil && err != bolt.ErrBucketNotFound {
		klog.Fatal(fmt.Sprintf("[UnsafeDeleteBucket]fail to delete bucket %s", bucket.Name()))
	}

	t.pending++
}

// UnsafeRange INFO: range scan @see https://github.com/etcd-io/bbolt#range-scans
func (t *batchTx) UnsafeRange(bucketType Bucket, key, endKey []byte, limit int64) (keys [][]byte, vals [][]byte) {
	bucket := t.tx.Bucket(bucketType.Name())
	if bucket == nil {
		klog.Fatalf(fmt.Sprintf("[UnsafeRange]"))
	}

	return unsafeRange(bucket.Cursor(), key, endKey, limit)
}

// INFO: range scan
//  limit<=0就是没有限制; endKey=nil,就只查startKey;
func unsafeRange(c *bolt.Cursor, startKey, endKey []byte, limit int64) ([][]byte, [][]byte) {
	if limit <= 0 {
		limit = math.MaxInt64
	}
	var isMatch func(key []byte) bool
	if len(endKey) > 0 {
		isMatch = func(key []byte) bool { return bytes.Compare(key, endKey) < 0 } // b < endKey
	} else {
		isMatch = func(key []byte) bool { return bytes.Equal(key, startKey) }
		limit = 1
	}

	// Iterate from key to endKey
	var (
		keys   [][]byte
		values [][]byte
	)
	for key, value := c.Seek(startKey); key != nil && isMatch(startKey); key, value = c.Next() {
		keys = append(keys, key)
		values = append(values, value)
		if limit == int64(len(keys)) {
			break
		}
	}

	return keys, values
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
	//  首先 boltdb key 是版本号，put/delete 操作时，都会基于当前版本号递增生成新的版本号，因此属于顺序写入，
	//  可以调整 boltdb 的 bucket.FillPercent 参数，使每个 page 填充更多数据，减少 page 的分裂次数并降低 db 空间。
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

	// INFO: batchTx 中所有写事务都会 t.pending++，会定期 100ms 批量提交事务
	t.pending++
}

func (t *batchTx) UnsafeDelete(bucketType Bucket, key []byte) {
	bucket := t.tx.Bucket(bucketType.Name())
	if bucket == nil {
		klog.Fatalf(fmt.Sprintf("[UnsafeDelete]bucket %s in boltdb is not existed", bucketType.Name()))
	}

	err := bucket.Delete(key)
	if err != nil {
		klog.Fatalf(fmt.Sprintf("[UnsafeDelete]failed to delete a key %s from bucket %s", string(key), bucketType.Name()))
	}

	t.pending++
}

// INFO: 提交事务，数据落盘，否则在 boltdb 内存 B+tree 中
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

// INFO: batchTx 中所有写事务都会 t.pending++，会定期 100ms 批量提交事务
//  UnsafeCreateBucket()/UnsafePut()/UnsafeDeleteBucket()/UnsafeDelete()
func (t *batchTx) safePending() int {
	t.Lock()
	defer t.Unlock()

	return t.pending
}

// INFO:
//  但是这优化又引发了另外的一个问题， 因为事务未提交，读请求可能无法从 boltdb 获取到最新数据？？？
//  为了解决这个问题，etcd 引入了一个 bucket buffer 来保存暂未提交的事务数据。在更新 boltdb 的时候，etcd 也会同步数据到 bucket buffer。
//  因此 etcd 处理读请求的时候会优先从 bucket buffer 里面读取，其次再从 boltdb 读，通过 bucket buffer 实现读写性能提升，同时保证数据一致性。

// 加 buffer 的 batchTx
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

/*func (t *batchTxBuffered) RLock() {
	panic("implement me")
}

func (t *batchTxBuffered) RUnlock() {
	panic("implement me")
}*/

func (t *batchTxBuffered) CommitAndStop() {
	panic("implement me")
}

// Commit INFO: 提交事务
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
	if t.backend.readTx.tx != nil {
		// wait all store read transactions using the current boltdb tx to finish,
		// then close the boltdb tx
		go func(tx *bolt.Tx, wg *sync.WaitGroup) {
			wg.Wait()
			if err := tx.Rollback(); err != nil {
				klog.Fatal(fmt.Sprintf("failed to rollback tx err %v", err))
			}
		}(t.backend.readTx.tx, t.backend.readTx.txWg)
		t.backend.readTx.reset()
	}

	t.batchTx.commit(stop)

	if !stop {
		// only-read transaction
		t.backend.readTx.tx = t.backend.begin(false)
	}
}

// UnsafePut INFO: batchTxBuffered 除了 boltdb 写一份(key, value)数据，同时在 buffer 中写一份(key, value)数据
func (t *batchTxBuffered) UnsafePut(bucket Bucket, key []byte, value []byte) {
	t.batchTx.UnsafePut(bucket, key, value)
	t.buf.put(bucket, key, value)
}

func (t *batchTxBuffered) UnsafeSeqPut(bucket Bucket, key []byte, value []byte) {
	t.batchTx.UnsafeSeqPut(bucket, key, value)
	t.buf.putSeq(bucket, key, value)
}
