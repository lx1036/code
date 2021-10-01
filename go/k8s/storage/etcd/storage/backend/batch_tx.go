package backend

import (
	"bytes"
	"fmt"
	"math"
	"sort"
	"sync"
	"sync/atomic"

	bolt "go.etcd.io/bbolt"
	"k8s.io/klog/v2"
)

// INFO: 读写事务 read-write txn

type BucketID int

type Bucket interface {
	ID() BucketID
	Name() []byte
	String() string
	IsSafeRangeBucket() bool
}

// INFO: (1) 写事务
type batchTx struct {
	sync.Mutex
	tx      *bolt.Tx // 读写事务
	backend *Backend

	pending int // pending put op 计数器
}

func (t *batchTx) Lock() {
	t.Mutex.Lock()
}

// Unlock INFO: 所有批量写事务必须先获得锁, 只有 pending put op 到了 10000，会立刻提交到 boltdb 中
func (t *batchTx) Unlock() {
	if t.pending >= t.backend.batchLimit {
		t.commit(false)
	}

	t.Mutex.Unlock()
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

// UnsafeRange
// INFO: range scan @see https://github.com/etcd-io/bbolt#range-scans
//  写事务的 range san 没有先从 buffer 里读的逻辑，直接读取磁盘文件数据
func (t *batchTx) UnsafeRange(bucketType Bucket, key, endKey []byte, limit int64) (keys [][]byte, vals [][]byte) {
	bucket := t.tx.Bucket(bucketType.Name())
	if bucket == nil {
		klog.Fatalf(fmt.Sprintf("[UnsafeRange]"))
	}

	return unsafeRange(bucket.Cursor(), key, endKey, limit)
}

// INFO: range scan
//  limit<=0就是没有限制; endKey=nil,就只查startKey;
func unsafeRange(cursor *bolt.Cursor, startKey, endKey []byte, limit int64) ([][]byte, [][]byte) {
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
	for key, value := cursor.Seek(startKey); key != nil && isMatch(startKey); key, value = cursor.Next() {
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

		// 数据落盘
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
//  UnsafeCreateBucket()/UnsafeDeleteBucket()/UnsafePut()/UnsafeDelete()
func (t *batchTx) safePending() int {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	return t.pending
}

// INFO:
//  但是这优化又引发了另外的一个问题，因为事务未提交，读请求可能无法从 boltdb 获取到最新数据？？？
//  为了解决这个问题，etcd 引入了一个 bucket buffer 来保存暂未提交的事务数据。在更新 boltdb 的时候，etcd 也会同步数据到 bucket buffer。
//  因此 etcd 处理读请求的时候会优先从 bucket buffer 里面读取，其次再从 boltdb 读，通过 bucket buffer 实现读写性能提升，同时保证数据一致性。

// INFO: (2) 写事务 buffer, 还未 committed pending put op

const bucketBufferInitialSize = 512

type kv struct {
	key []byte
	val []byte
}

// INFO: bucketBuffer 设计还需要在研究研究
//  但是这优化又引发了另外的一个问题， 因为事务未提交，读请求可能无法从 boltdb 获取到最新数据???
//  为了解决这个问题，etcd 引入了一个 bucket buffer 来保存暂未提交的事务数据。在更新 boltdb 的时候，etcd 也会同步数据到 bucket buffer。
//  因此 etcd 处理读请求的时候会优先从 bucket buffer 里面读取，其次再从 boltdb 读，通过 bucket buffer 实现读写性能提升，同时保证数据一致性。

// bucketBuffer buffers key-value pairs that are pending commit.
type bucketBuffer struct {
	buf []kv // INFO: [](key,value) 按照 key 升序
	// used tracks number of elements in use so buf can be reused without reallocation.
	used int
}

func newBucketBuffer() *bucketBuffer {
	return &bucketBuffer{
		buf:  make([]kv, bucketBufferInitialSize),
		used: 0,
	}
}

func (buffer *bucketBuffer) Len() int { return buffer.used }
func (buffer *bucketBuffer) Less(i, j int) bool {
	return bytes.Compare(buffer.buf[i].key, buffer.buf[j].key) < 0
}
func (buffer *bucketBuffer) Swap(i, j int) {
	buffer.buf[i], buffer.buf[j] = buffer.buf[j], buffer.buf[i]
}

func (buffer *bucketBuffer) Copy() *bucketBuffer {
	bufferCopy := bucketBuffer{
		buf:  make([]kv, len(buffer.buf)),
		used: buffer.used,
	}

	copy(bufferCopy.buf, buffer.buf)

	return &bufferCopy
}

// Range INFO: 这里重点是从 buffer 中查找 [key, endKey] 之间的 (key, value)
func (buffer *bucketBuffer) Range(key, endKey []byte, limit int64) (keys [][]byte, vals [][]byte) {
	// INFO: 从[0,buffer.used)中迭代查找，找到最前面 buffer.buf[i].key >= key
	f := func(i int) bool { return bytes.Compare(buffer.buf[i].key, key) >= 0 }
	idx := sort.Search(buffer.used, f)
	if idx < 0 {
		return nil, nil
	}
	// [key, nil] 就是当前 key 的 (key,value)
	if len(endKey) == 0 {
		if bytes.Equal(key, buffer.buf[idx].key) {
			keys = append(keys, buffer.buf[idx].key)
			vals = append(vals, buffer.buf[idx].val)
		}
		return keys, vals
	}

	if bytes.Compare(endKey, buffer.buf[idx].key) <= 0 {
		return nil, nil
	}

	// [idx, buffer.used) 开始查找一直到 endKey
	for i := idx; i < buffer.used && int64(len(keys)) < limit; i++ {
		if bytes.Compare(endKey, buffer.buf[i].key) <= 0 {
			break
		}

		keys = append(keys, buffer.buf[i].key)
		vals = append(vals, buffer.buf[i].val)
	}

	return keys, vals
}

func (buffer *bucketBuffer) add(key []byte, value []byte) {
	buffer.buf[buffer.used].key, buffer.buf[buffer.used].val = key, value
	buffer.used++
	// INFO: 如果满了，则buffer 1.5倍扩容
	if buffer.used == len(buffer.buf) {
		buf := make([]kv, (3*len(buffer.buf))/2)
		copy(buf, buffer.buf)
		buffer.buf = buf
	}
}

func (buffer *bucketBuffer) merge(other *bucketBuffer) {
	for i := 0; i < other.used; i++ {
		buffer.add(other.buf[i].key, other.buf[i].val)
	}
	if buffer.used == other.used {
		return
	}
	//
	if bytes.Compare(buffer.buf[(buffer.used-other.used)-1].key, other.buf[0].key) < 0 {
		return
	}
	// INFO: 按照 key 排序
	sort.Stable(buffer)

	// TODO: 没太懂这里的逻辑???, 参考 TestConcurrentReadTxn() 有重复key
	// remove duplicates, using only newest update
	widx := 0
	for ridx := 1; ridx < buffer.used; ridx++ {
		if !bytes.Equal(buffer.buf[ridx].key, buffer.buf[widx].key) {
			widx++
		}
		buffer.buf[widx] = buffer.buf[ridx]
	}
	buffer.used = widx + 1
}

// txBuffer handles functionality shared between txWriteBuffer and txReadBuffer.
type txBuffer struct {
	buckets map[BucketID]*bucketBuffer
}

func (txb *txBuffer) reset() {
	for k, v := range txb.buckets {
		if v.used == 0 {
			delete(txb.buckets, k)
		}
		v.used = 0
	}
}

// INFO: 参考 txReadBuffer struct
type txWriteBuffer struct {
	txBuffer
	// Map from bucket ID into information whether this bucket is edited
	// sequentially (i.e. keys are growing monotonically).
	bucket2seq map[BucketID]bool
}

func (writeBuffer *txWriteBuffer) put(bucketType Bucket, key []byte, value []byte) {
	writeBuffer.bucket2seq[bucketType.ID()] = false
	writeBuffer.putInternal(bucketType, key, value)
}

func (writeBuffer *txWriteBuffer) putSeq(bucketType Bucket, key []byte, value []byte) {
	writeBuffer.putInternal(bucketType, key, value)
}

// INFO: (key, value) 写到 buffer 里
func (writeBuffer *txWriteBuffer) putInternal(bucketType Bucket, key, value []byte) {
	bucketBuffer, ok := writeBuffer.buckets[bucketType.ID()]
	if !ok {
		bucketBuffer = newBucketBuffer()
		writeBuffer.buckets[bucketType.ID()] = bucketBuffer
	}

	bucketBuffer.add(key, value)
}

func (writeBuffer *txWriteBuffer) reset() {
	writeBuffer.txBuffer.reset()
	for bucketID := range writeBuffer.bucket2seq {
		value, ok := writeBuffer.buckets[bucketID]
		if !ok {
			delete(writeBuffer.bucket2seq, bucketID)
		} else if value.used == 0 {
			writeBuffer.bucket2seq[bucketID] = true
		}
	}
}

// 写事务写到 txReadBuffer 里
func (writeBuffer *txWriteBuffer) writeBack(readBuffer *txReadBuffer) {
	for id, writeBuf := range writeBuffer.buckets {
		readBuf, ok := readBuffer.buckets[id]
		if !ok {
			// 从 writeBuckets 里删除，但是保存到 readBuckets
			delete(writeBuffer.buckets, id)
			readBuffer.buckets[id] = writeBuf
			continue
		}
		if seq, ok := writeBuffer.bucket2seq[id]; ok && !seq && writeBuf.used > 1 {
			sort.Sort(writeBuf)
		}

		// read buffer 合并 write buffer
		readBuf.merge(writeBuf)
	}

	writeBuffer.reset()
	readBuffer.bufVersion++
}

// 加 buffer 的 batchTx
type batchTxBuffered struct {
	batchTx
	buf txWriteBuffer
}

func newBatchTxBuffered(backend *Backend) *batchTxBuffered {
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

// Unlock INFO: 非常重要，写事务其实是先写到 txReadBuffer 里，但是需要先 blocks txReadBuffer 阻塞读操作!!!
func (t *batchTxBuffered) Unlock() {
	if t.pending != 0 {
		t.backend.readTx.Lock()
		t.buf.writeBack(&t.backend.readTx.buf)
		t.backend.readTx.Unlock()
	}

	t.batchTx.Unlock()
}

func (t *batchTxBuffered) CommitAndStop() {
	t.Lock()
	t.commit(true)
	t.Unlock()
}

// Commit INFO: 提交事务，这里会设置 only-read transaction
//  https://github.com/etcd-io/bbolt#managing-transactions-manually
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
		// TODO: 没太懂这里的逻辑!!!
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
