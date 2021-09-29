package mvcc

import (
	"context"
	"fmt"

	"k8s-lx1036/k8s/storage/etcd/storage/backend"

	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.etcd.io/etcd/server/v3/lease"
	"k8s.io/klog/v2"
)

type ReadTxMode uint32

const (
	// Use ConcurrentReadTx and the txReadBuffer is copied
	ConcurrentReadTxMode = ReadTxMode(1)
	// Use backend ReadTx and txReadBuffer is not copied
	SharedBufReadTxMode = ReadTxMode(2)
)

type RangeOptions struct {
	Limit int64
	Rev   int64
	Count bool // 如果是 true，只是返回 revisions 个数
}

type RangeResult struct {
	KVs   []mvccpb.KeyValue
	Rev   int64
	Count int
}

type ReadView interface {
	// Range gets the keys in the range at rangeRev.
	// The returned rev is the current revision of the KV when the operation is executed.
	// If rangeRev <=0, range gets the keys at currentRev.
	// If `end` is nil, the request returns the key.
	// If `end` is not nil and not empty, it gets the keys in range [key, range_end).
	// If `end` is not nil and empty, it gets the keys greater than or equal to key.
	// Limit limits the number of keys returned.
	// If the required rev is compacted, ErrCompacted will be returned.
	Range(ctx context.Context, key, end []byte, ro RangeOptions) (r *RangeResult, err error)

	// Rev returns the revision of the KV at the time of opening the txn.
	Rev() int64
}

type WriteView interface {
	Put(key, value []byte, lease lease.LeaseID) (rev int64)

	// DeleteRange
	// INFO: 范围删除，会触发 delete event
	//  delete keys in [key, end)
	//  如果end is nil, delete the key
	DeleteRange(key, end []byte) (n, rev int64)
}

// INFO: 只读事务
type TxnRead interface {
	ReadView

	// End marks the transaction is complete and ready to commit.
	End()
}

// INFO: 读写事务
type TxnWrite interface {
	TxnRead

	WriteView

	// Changes gets the changes made since opening the write txn.
	Changes() []mvccpb.KeyValue
}

type readView struct{ kv KV }

func (rv *readView) Rev() int64 {
	tr := rv.kv.Read(ConcurrentReadTxMode)
	defer tr.End()
	return tr.Rev()
}

func (rv *readView) Range(ctx context.Context, key, end []byte, ro RangeOptions) (r *RangeResult, err error) {
	tr := rv.kv.Read(ConcurrentReadTxMode)
	defer tr.End()
	return tr.Range(ctx, key, end, ro)
}

// INFO: "读事务"
type storeTxnRead struct {
	s       *store
	readTxn backend.ReadTx

	firstRev int64

	// INFO: 当前 revision, 也是当前 store revision
	rev int64
}

func (s *store) Read(mode ReadTxMode) TxnRead {
	s.mu.RLock()
	s.revMu.RLock()

	// INFO: 如果是并发读事务，没有锁，并且从 buffer 里读，提高性能
	var txn backend.ReadTx
	if mode == ConcurrentReadTxMode {
		txn = s.b.ConcurrentReadTx()
	} else {
		txn = s.b.ReadTx()
	}

	// INFO: 如果是 concurrentReadTx, tx.RLock() 是 no-op，这样可以实现并发读
	txn.RLock()
	s.revMu.RUnlock()

	return &storeTxnRead{
		s:        s,
		readTxn:  txn,
		firstRev: s.compactMainRev,
		rev:      s.currentRev,
	}
}

// Range INFO: 读事务 range read
func (tr *storeTxnRead) Range(ctx context.Context, key, end []byte, rangeOptions RangeOptions) (r *RangeResult, err error) {
	return tr.rangeKeys(ctx, key, end, tr.Rev(), rangeOptions)
}

func (tr *storeTxnRead) rangeKeys(ctx context.Context, key, end []byte, curRev int64, rangeOptions RangeOptions) (r *RangeResult, err error) {
	rev := rangeOptions.Rev
	if rev > curRev {
		return &RangeResult{KVs: nil, Count: -1, Rev: curRev}, ErrFutureRev
	}
	if rev <= 0 {
		rev = curRev
	}
	if rev < tr.s.compactMainRev {
		return &RangeResult{KVs: nil, Count: -1, Rev: 0}, ErrCompacted
	}
	if rangeOptions.Count {
		total := tr.s.treeIndex.CountRevisions(key, end, rev)
		return &RangeResult{KVs: nil, Count: total, Rev: curRev}, nil
	}

	revisions, total := tr.s.treeIndex.Revisions(key, end, rev, int(rangeOptions.Limit))
	if len(revisions) == 0 {
		return &RangeResult{KVs: nil, Count: total, Rev: curRev}, nil
	}
	limit := int(rangeOptions.Limit)
	if limit <= 0 || limit > len(revisions) {
		limit = len(revisions)
	}
	keyValues := make([]mvccpb.KeyValue, limit)
	revBytes := newRevBytes()
	for i, revision := range revisions[:len(keyValues)] {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		revToBytes(revision, revBytes)
		_, keyValue := tr.readTxn.UnsafeRange(backend.Key, revBytes, nil, 0)
		if err := keyValues[i].Unmarshal(keyValue[0]); err != nil {
			// INFO: 居然可以直接这样 Unmarshal!!!可以抄一抄!!!
			klog.Fatalf(fmt.Sprintf("[rangeKeys]failed to unmarshal mvccpb.KeyValue err: %v", err))
		}
	}

	return &RangeResult{KVs: keyValues, Count: total, Rev: curRev}, nil
}

func (tr *storeTxnRead) Rev() int64 {
	return tr.rev
}

func (tr *storeTxnRead) End() {
	tr.readTxn.RUnlock() // RUnlock signals the end of concurrentReadTx.
	tr.s.mu.RUnlock()
}

type writeView struct{ kv KV }

func (wv *writeView) Put(key, value []byte, lease lease.LeaseID) (rev int64) {
	tw := wv.kv.Write()
	defer tw.End()

	return tw.Put(key, value, lease)
}

func (wv *writeView) DeleteRange(key, end []byte) (n, rev int64) {
	tw := wv.kv.Write()
	defer tw.End()

	return tw.DeleteRange(key, end)
}

// INFO: "写事务"
type storeTxnWrite struct {
	storeTxnRead
	writeTxn backend.BatchTx

	// INFO: 这是个全局版本，写事务时自增
	beginRev int64

	// INFO: 见 put()，存入 boltdb 里每个 (key,value)
	changes []mvccpb.KeyValue
}

func (s *store) Write() TxnWrite {
	s.mu.RLock()
	// INFO: 所有批量写事务必须先获得锁，见 backend.batchTx::Lock()
	writeTxn := s.b.BatchTx()
	writeTxn.Lock()

	return &storeTxnWrite{
		storeTxnRead: storeTxnRead{s, writeTxn, 0, 0},
		writeTxn:     writeTxn,
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
		tw.s.currentRev++ // INFO: 当前 revision, 也是当前 store revision，会在 End() 之后 +1
	}
	// INFO: 所有批量写事务必须先获得锁，见 backend.batchTx::Lock()
	tw.writeTxn.Unlock()
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
	keys, _ := tw.s.treeIndex.Range(key, end, rrev)
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

	tw.writeTxn.UnsafeSeqPut(backend.Key, ibytes, data)
	err = tw.s.treeIndex.Tombstone(key, idxRev) // 删除keyIndex，关闭这一代 generation revisions
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
	createRevision := rev
	oldLease := lease.NoLease

	// INFO: 从 treeIndex 根据(key, revision.main)获取最新 revision
	_, created, ver, err := tw.s.treeIndex.Get(key, rev)
	if err == nil {
		createRevision = created.main
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
		CreateRevision: createRevision,
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
	tw.writeTxn.UnsafeSeqPut(backend.Key, boltdbKey, data)
	tw.s.treeIndex.Put(key, newRev) // 更新keyIndex到treeIndex
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
