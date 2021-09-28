package mvcc

import (
	"context"
	"fmt"

	"k8s-lx1036/k8s/storage/etcd/storage/backend"

	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.etcd.io/etcd/server/v3/lease"
	"k8s.io/klog/v2"
)

// INFO: "读事务"
type storeTxnRead struct {
	s  *store
	tx backend.ReadTx

	firstRev int64
	rev      int64
}

func (s *store) Read(mode ReadTxMode) TxnRead {
	s.mu.RLock()
	s.revMu.RLock()

	// TODO: 没太明白这里的意思
	// For read-only workloads, we use shared buffer by copying transaction read buffer
	// for higher concurrency with ongoing blocking writes.
	// For write/write-read transactions, we use the shared buffer
	// rather than duplicating transaction read buffer to avoid transaction overhead.
	var tx backend.ReadTx
	if mode == ConcurrentReadTxMode {
		tx = s.b.ConcurrentReadTx()
	} else {
		tx = s.b.ReadTx()
	}

	tx.RLock() // RLock is no-op. concurrentReadTx does not need to be locked after it is created.
	firstRev, rev := s.compactMainRev, s.currentRev
	s.revMu.RUnlock()

	return &storeTxnRead{
		s:        s,
		tx:       tx,
		firstRev: firstRev,
		rev:      rev,
	}
}

// Range INFO: 读事务 range read
func (tr *storeTxnRead) Range(ctx context.Context, key, end []byte, ro RangeOptions) (r *RangeResult, err error) {
	return tr.rangeKeys(ctx, key, end, tr.Rev(), ro)
}

func (tr *storeTxnRead) rangeKeys(ctx context.Context, key, end []byte, curRev int64, ro RangeOptions) (r *RangeResult, err error) {
	rev := ro.Rev
	if rev > curRev {
		return &RangeResult{KVs: nil, Count: -1, Rev: curRev}, ErrFutureRev
	}
	if rev <= 0 {
		rev = curRev
	}
	if rev < tr.s.compactMainRev {
		return &RangeResult{KVs: nil, Count: -1, Rev: 0}, ErrCompacted
	}
	if ro.Count {
		total := tr.s.kvindex.CountRevisions(key, end, rev)
		return &RangeResult{KVs: nil, Count: total, Rev: curRev}, nil
	}

	revisions, total := tr.s.kvindex.Revisions(key, end, rev, int(ro.Limit))
	if len(revisions) == 0 {
		return &RangeResult{KVs: nil, Count: total, Rev: curRev}, nil
	}
	limit := int(ro.Limit)
	if limit <= 0 || limit > len(revisions) {
		limit = len(revisions)
	}
	kvs := make([]mvccpb.KeyValue, limit)
	revBytes := newRevBytes()
	for i, revision := range revisions[:len(kvs)] {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		revToBytes(revision, revBytes)
		_, vs := tr.tx.UnsafeRange(backend.Key, revBytes, nil, 0)
		if err := kvs[i].Unmarshal(vs[0]); err != nil {
			// TODO: 可以直接这样 Unmarshal 么???
			klog.Fatalf(fmt.Sprintf("[rangeKeys]failed to unmarshal mvccpb.KeyValue err: %v", err))
		}
	}

	return &RangeResult{KVs: kvs, Count: total, Rev: curRev}, nil
}

func (tr *storeTxnRead) Rev() int64 {
	return tr.rev
}

func (tr *storeTxnRead) End() {
	tr.tx.RUnlock() // RUnlock signals the end of concurrentReadTx.
	tr.s.mu.RUnlock()
}

// INFO: "写事务"
type storeTxnWrite struct {
	storeTxnRead
	tx backend.BatchTx

	// INFO: 这是个全局版本，写事务时自增
	beginRev int64

	// INFO: 见 put()，存入 boltdb 里每个 (key,value)
	changes []mvccpb.KeyValue
}

func (s *store) Write() TxnWrite {
	s.mu.RLock()
	// INFO: 所有批量写事务必须先获得锁，见 backend.batchTx::Lock()
	tx := s.b.BatchTx()
	tx.Lock()

	return &storeTxnWrite{
		storeTxnRead: storeTxnRead{s, tx, 0, 0},
		tx:           tx,
		beginRev:     s.currentRev,
		changes:      make([]mvccpb.KeyValue, 0, 4),
	}
}

func (tw *storeTxnWrite) Changes() []mvccpb.KeyValue {
	return tw.changes
}

// INFO:
func (tw *storeTxnWrite) Rev() int64 {
	return tw.beginRev
}

// TODO: 没太懂这块知识
func (tw *storeTxnWrite) End() {
	// only update index if the txn modifies the mvcc state.
	if len(tw.changes) != 0 {
		// hold revMu lock to prevent new read txns from opening until writeback.
		tw.s.revMu.Lock()
		tw.s.currentRev++
	}
	// INFO: 所有批量写事务必须先获得锁，见 backend.batchTx::Lock()
	tw.tx.Unlock()
	if len(tw.changes) != 0 {
		tw.s.revMu.Unlock()
	}
	tw.s.mu.RUnlock()
}

// INFO: range delete
func (tw *storeTxnWrite) DeleteRange(key, end []byte) (int64, int64) {
	if n := tw.deleteRange(key, end); n != 0 || len(tw.changes) > 0 {
		return n, tw.beginRev + 1
	}
	return 0, tw.beginRev
}

func (tw *storeTxnWrite) deleteRange(key, end []byte) int64 {
	rrev := tw.beginRev
	if len(tw.changes) > 0 {
		rrev++
	}
	keys, _ := tw.s.kvindex.Range(key, end, rrev)
	if len(keys) == 0 {
		return 0
	}

	// 返回删除的keys个数，从boltdb中删除
	for _, key := range keys {
		tw.delete(key)
	}

	return int64(len(keys))
}

func (tw *storeTxnWrite) delete(key []byte) {
	ibytes := newRevBytes()
	idxRev := revision{main: tw.beginRev + 1, sub: int64(len(tw.changes))}
	revToBytes(idxRev, ibytes)

	if len(ibytes) != revBytesLen {
		klog.Errorf(fmt.Sprintf("[storeTxnWrite delete]cannot append tombstone mark to non-normal revision bytes, expected-revision-bytes-size %d given-revision-bytes-size %d",
			revBytesLen, len(ibytes)))
		return
	}
	ibytes = append(ibytes, markTombstone) // 加个 't' suffix

	kv := mvccpb.KeyValue{Key: key}

	data, err := kv.Marshal()
	if err != nil {
		klog.Errorf(fmt.Sprintf("[storeTxnWrite delete]key %s, KeyValue marshal error %v", string(key), err))
		return
	}

	tw.tx.UnsafeSeqPut(backend.Key, ibytes, data)
	err = tw.s.kvindex.Tombstone(key, idxRev) // 删除keyIndex，关闭这一代 generation revisions
	if err != nil {
		klog.Errorf(fmt.Sprintf("[storeTxnWrite delete]key %s, tombstore err %v", string(key), err))
		return
	}
	tw.changes = append(tw.changes, kv)

	item := lease.LeaseItem{Key: string(key)}
	leaseID := tw.s.le.GetLease(item)
	if leaseID != lease.NoLease {
		err = tw.s.le.Detach(leaseID, []lease.LeaseItem{{Key: string(key)}})
		if err != nil {
			klog.Errorf(fmt.Sprintf("[storeTxnWrite put]unexpected error from lease Detach %v", err))
		}
	}
}

func (tw *storeTxnWrite) Put(key, value []byte, lease lease.LeaseID) (rev int64) {
	tw.put(key, value, lease)
	return tw.beginRev + 1
}

// INFO: 这个函数会把(key,value)构造出KeyValue结构体，并持久化到boltdb中
func (tw *storeTxnWrite) put(key, value []byte, leaseID lease.LeaseID) {
	rev := tw.beginRev + 1 // INFO: beginRev 是全局版本号，每次写事务 +1
	c := rev
	oldLease := lease.NoLease

	// INFO: 从 treeIndex 根据(key, revision.main)获取最新 revision
	_, created, ver, err := tw.s.kvindex.Get(key, rev)
	if err == nil {
		c = created.main
		oldLease = tw.s.le.GetLease(lease.LeaseItem{Key: string(key)})
	}

	boltdbKey := newRevBytes()
	newRev := revision{main: rev, sub: int64(len(tw.changes))}
	revToBytes(newRev, boltdbKey) // INFO: 存储在boltdb里的key: "${main}_${sub}"

	// INFO: 构造KeyValue对象，准备要存入boltdb
	ver = ver + 1 // 修改次数，版本号
	kv := mvccpb.KeyValue{
		Key:            key,
		Value:          value,
		CreateRevision: c,
		ModRevision:    rev, // INFO: key 最后一次修改时的版本，这个版本也是etcd写事务版本，全局的!!!
		Version:        ver,
		Lease:          int64(leaseID),
	}
	data, err := kv.Marshal()
	if err != nil {
		klog.Infof(fmt.Sprintf("[storeTxnWrite put]key %s, KeyValue marshal error %v", string(key), err))
		return
	}

	// INFO: 这里是真正要持久化到boltdb
	tw.tx.UnsafeSeqPut(backend.Key, boltdbKey, data)
	tw.s.kvindex.Put(key, newRev) // 更新keyIndex到treeIndex
	tw.changes = append(tw.changes, kv)

	if oldLease != lease.NoLease {
		err = tw.s.le.Detach(oldLease, []lease.LeaseItem{{Key: string(key)}})
		if err != nil {
			klog.Errorf(fmt.Sprintf("[storeTxnWrite put]unexpected error from lease Detach %v", err))
		}
	}

	if leaseID != lease.NoLease {
		err = tw.s.le.Attach(leaseID, []lease.LeaseItem{{Key: string(key)}})
		if err != nil {
			klog.Errorf(fmt.Sprintf("[storeTxnWrite put]unexpected error from lease Attach %v", err))
		}
	}
}
