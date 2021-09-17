package mvcc

import (
	"sync"

	"go.etcd.io/etcd/server/v3/lease"
	"go.etcd.io/etcd/server/v3/mvcc/backend"
)

const (
	// markedRevBytesLen is the byte length of marked revision.
	// The first `revBytesLen` bytes represents a normal revision. The last
	// one byte is the mark.
	markedRevBytesLen = revBytesLen + 1
)

var defaultCompactBatchLimit = 1000

type StoreConfig struct {
	CompactionBatchLimit int
}

type store struct {
	// mu read locks for txns and write locks for non-txn store changes.
	mu sync.RWMutex

	ReadView
	WriteView

	cfg StoreConfig

	b backend.Backend

	// 即 treeIndex
	kvindex Index

	le lease.Lessor

	// currentRev is the revision of the last completed transaction.
	currentRev int64

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
	}

	s.ReadView = &readView{s}
	s.WriteView = &writeView{s}

}

func (s *store) Close() error {
	close(s.stopc)
	//s.fifoSched.Stop()
	return nil
}
