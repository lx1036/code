package mvcc

import (
	"errors"
	"fmt"
	"sync"

	"k8s-lx1036/k8s/storage/etcd/storage/backend"

	"go.etcd.io/etcd/server/v3/lease"
	"k8s.io/klog/v2"
)

const (
	// markedRevBytesLen is the byte length of marked revision.
	// The first `revBytesLen` bytes represents a normal revision. The last
	// one byte is the mark.
	markedRevBytesLen = revBytesLen + 1
	markBytePosition  = markedRevBytesLen - 1

	markTombstone byte = 't'
)

var (
	ErrCompacted = errors.New("mvcc: required revision has been compacted")
	ErrFutureRev = errors.New("mvcc: required revision is a future revision")
)

var defaultCompactBatchLimit = 1000

type KV interface {
	ReadView
	WriteView

	// Write INFO: 创建写事务
	Write() TxnWrite

	// Read INFO: 创建读事务
	Read(mode ReadTxMode) TxnRead
}

type StoreConfig struct {
	CompactionBatchLimit int
}

// INFO: kvstore 是一个封装对象，具有事务功能，把读写分为 "读事务/写事务",主要包含了 treeIndex(keyIndex/revision) 和 Backend 对象
//
//	(1)先从B+tree treeIndex 中查找出当前 key 的 revision
//	(2)再从 Backend 中以 revision 为 key 查找出 value, 该 value 包含用户输入的 (key, value)
type store struct {
	// mu read locks for txns and write locks for non-txn store changes.
	mu sync.RWMutex
	// revMuLock protects currentRev and compactMainRev.
	// Locked at end of write txn and released after write txn unlock lock.
	// Locked before locking read txn and released after locking.
	revMu sync.RWMutex

	// INFO: store 包含读写事务功能
	ReadView
	WriteView

	cfg StoreConfig

	b backend.Backend

	treeIndex TreeIndex

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
		cfg:       cfg,
		b:         b,
		treeIndex: newTreeIndex(),

		le: le,

		currentRev:     1, // INFO: etcd 当前全局版本号，默认为 1
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
	tx.UnsafeCreateBucket(backend.Key)
	tx.UnsafeCreateBucket(backend.Meta)
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

// isTombstone INFO: revision bytes is a tombstone, "xxx_xxxt"
func isTombstone(b []byte) bool {
	return len(b) == markedRevBytesLen && b[markBytePosition] == markTombstone
}
