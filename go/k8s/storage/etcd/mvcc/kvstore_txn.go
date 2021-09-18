package mvcc

import (
	"fmt"

	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.etcd.io/etcd/server/v3/lease"
	"go.etcd.io/etcd/server/v3/mvcc/backend"
	"go.etcd.io/etcd/server/v3/mvcc/buckets"
	"k8s.io/klog/v2"
)

type storeTxnRead struct {
	s  *store
	tx backend.ReadTx

	firstRev int64
	rev      int64
}

func (tr *storeTxnRead) Rev() int64 {
	panic("implement me")
}

func (tr *storeTxnRead) End() {
	panic("implement me")
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

type storeTxnWrite struct {
	storeTxnRead
	tx backend.BatchTx
	// beginRev is the revision where the txn begins; it will write to the next revision.
	beginRev int64
	changes  []mvccpb.KeyValue
}

func (s *store) Write() TxnWrite {
	s.mu.RLock()
	tx := s.b.BatchTx()
	tx.Lock()

	return &storeTxnWrite{
		storeTxnRead: storeTxnRead{s, tx, 0, 0},
		tx:           tx,
		beginRev:     s.currentRev,
		changes:      make([]mvccpb.KeyValue, 0, 4),
	}
}

// INFO:
func (tw *storeTxnWrite) Rev() int64 {
	return tw.beginRev
}

func (tw *storeTxnWrite) End() {
	panic("implement me")
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

	tw.tx.UnsafeSeqPut(buckets.Key, ibytes, data)
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
	rev := tw.beginRev + 1
	c := rev
	oldLease := lease.NoLease

	// if the key exists before, use its previous created and
	// get its previous leaseID
	// INFO: 从 treeIndex 根据(key, revision.main)获取最新 revision
	_, created, ver, err := tw.s.kvindex.Get(key, rev)
	if err == nil {
		c = created.main
		oldLease = tw.s.le.GetLease(lease.LeaseItem{Key: string(key)})
	}

	ibytes := newRevBytes()
	idxRev := revision{main: rev, sub: int64(len(tw.changes))}
	revToBytes(idxRev, ibytes)

	// INFO: 构造KeyValue对象，准备要存入boltdb
	ver = ver + 1
	kv := mvccpb.KeyValue{
		Key:            key,
		Value:          value,
		CreateRevision: c,
		ModRevision:    rev,
		Version:        ver,
		Lease:          int64(leaseID),
	}
	data, err := kv.Marshal()
	if err != nil {
		klog.Infof(fmt.Sprintf("[storeTxnWrite put]key %s, KeyValue marshal error %v", string(key), err))
		return
	}

	// INFO: 这里是真正要持久化到boltdb
	tw.tx.UnsafeSeqPut(buckets.Key, ibytes, data)
	tw.s.kvindex.Put(key, idxRev) // 更新keyIndex到treeIndex
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
