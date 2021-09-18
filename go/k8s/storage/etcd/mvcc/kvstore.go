package mvcc

import (
	"fmt"
	"sync"

	"go.etcd.io/etcd/server/v3/lease"
	"go.etcd.io/etcd/server/v3/mvcc/backend"
	"go.etcd.io/etcd/server/v3/mvcc/buckets"
	"k8s.io/klog/v2"
)

const (
	// markedRevBytesLen is the byte length of marked revision.
	// The first `revBytesLen` bytes represents a normal revision. The last
	// one byte is the mark.
	markedRevBytesLen = revBytesLen + 1

	markTombstone byte = 't'
)

var defaultCompactBatchLimit = 1000

type StoreConfig struct {
	CompactionBatchLimit int
}

type store struct {
	// mu read locks for txns and write locks for non-txn store changes.
	mu sync.RWMutex
	// revMuLock protects currentRev and compactMainRev.
	// Locked at end of write txn and released after write txn unlock lock.
	// Locked before locking read txn and released after locking.
	revMu sync.RWMutex

	ReadView
	WriteView

	cfg StoreConfig

	b backend.Backend

	// 即 treeIndex
	kvindex Index

	le lease.Lessor

	// currentRev is the revision of the last completed transaction.
	currentRev int64
	// compactMainRev is the main revision of the last compaction.
	compactMainRev int64

	stopc chan struct{}
}

func NewStore(b backend.Backend, le lease.Lessor, cfg StoreConfig) *store {
	if cfg.CompactionBatchLimit == 0 {
		cfg.CompactionBatchLimit = defaultCompactBatchLimit
	}

	s := &store{
		cfg:     cfg,
		b:       b,
		kvindex: newTreeIndex(),

		le: le,

		currentRev:     1,
		compactMainRev: -1,

		stopc: make(chan struct{}),
	}

	s.ReadView = &readView{s}
	s.WriteView = &writeView{s}
	if s.le != nil {
		s.le.SetRangeDeleter(func() lease.TxnDelete { return s.Write() })
	}

	// INFO: batch transaction, boltdb可以批量提交 @see https://github.com/etcd-io/bbolt#batch-read-write-transactions
	tx := s.b.BatchTx()
	tx.Lock()
	tx.UnsafeCreateBucket(buckets.Key)
	tx.UnsafeCreateBucket(buckets.Meta)
	tx.Unlock()
	s.b.ForceCommit()

	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.restore(); err != nil {
		klog.Errorf(fmt.Sprintf("[NewStore]restore err %v", err))
	}

	return s
}

// TODO: restore
func (s *store) restore() error {
	return nil
}

func (s *store) Close() error {
	close(s.stopc)
	//s.fifoSched.Stop()
	return nil
}
