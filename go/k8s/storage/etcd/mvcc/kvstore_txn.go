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

func (tw *storeTxnWrite) Put(key, value []byte, lease lease.LeaseID) (rev int64) {
	tw.put(key, value, lease)
	return tw.beginRev + 1
}

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
		klog.Infof(fmt.Sprintf("[storeTxnWrite put]KeyValue marshal error %v", err))
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
