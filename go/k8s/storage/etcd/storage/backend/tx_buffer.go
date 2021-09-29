package backend

import (
	"bytes"
	"sort"
	"sync"
)

const bucketBufferInitialSize = 512

// txBuffer handles functionality shared between txWriteBuffer and txReadBuffer.
type txBuffer struct {
	buckets map[BucketID]*bucketBuffer
}

func (txb *txBuffer) reset() {
	for k, v := range txb.buckets {
		if v.used == 0 {
			// demote
			delete(txb.buckets, k)
		}
		v.used = 0
	}
}

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
	buf []kv
	// used tracks number of elements in use so buf can be reused without reallocation.
	used int
}

func newBucketBuffer() *bucketBuffer {
	return &bucketBuffer{
		buf:  make([]kv, bucketBufferInitialSize),
		used: 0,
	}
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
		if bytes.Compare(endKey, buffer.buf[idx].key) <= 0 {
			break
		}

		keys = append(keys, buffer.buf[idx].key)
		vals = append(vals, buffer.buf[idx].val)
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

// txWriteBuffer buffers writes of pending updates that have not yet committed.
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

type txReadBufferCache struct {
	mu         sync.Mutex
	buf        *txReadBuffer
	bufVersion uint64
}

// txReadBuffer accesses buffered updates.
type txReadBuffer struct {
	txBuffer
	// bufVersion is used to check if the buffer is modified recently
	bufVersion uint64
}

// INFO: 这里 copy 时 bufVersion=0
func (txr *txReadBuffer) unsafeCopy() txReadBuffer {
	txrCopy := txReadBuffer{
		txBuffer: txBuffer{
			buckets: make(map[BucketID]*bucketBuffer, len(txr.txBuffer.buckets)),
		},
		bufVersion: 0, // 这里可以看 backend.ConcurrentReadTx() 里会重置
	}
	for bucketName, bucket := range txr.txBuffer.buckets {
		txrCopy.txBuffer.buckets[bucketName] = bucket.Copy()
	}

	return txrCopy
}

func (txr *txReadBuffer) Range(bucketType Bucket, key, endKey []byte, limit int64) ([][]byte, [][]byte) {
	if buffer := txr.buckets[bucketType.ID()]; buffer != nil {
		return buffer.Range(key, endKey, limit)
	}

	return nil, nil
}
