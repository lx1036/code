package backend

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	bolt "go.etcd.io/bbolt"
	"k8s.io/klog/v2"
)

var (
	defaultBatchLimit    = 10000
	defaultBatchInterval = 100 * time.Millisecond

	// initialMmapSize is the initial size of the mmapped region. Setting this larger than
	// the potential max db size can prevent writer from blocking reader.
	// This only works for linux.
	initialMmapSize = uint64(10 * 1024 * 1024 * 1024)
)

type Backend interface {
	BatchTx() BatchTx
	// ReadTx 读事务/并发读
	ReadTx() ReadTx
	// ConcurrentReadTx non-blocking read transaction
	ConcurrentReadTx() ReadTx

	Snapshot() Snapshot

	ForceCommit()
	Close() error
}

type backend struct {
	sync.RWMutex

	batchTx *batchTxBuffered
	readTx  *readTx

	db *bolt.DB
	// Size returns current database size in bytes as seen by this transaction.
	size int64
	// sizeInUse is the number of bytes actually used in the backend
	sizeInUse int64
	// openReadTxN is the number of currently open read transactions in the backend
	openReadTxN   int64
	batchInterval time.Duration
	batchLimit    int
	// mlock prevents backend database file to be swapped
	mlock bool
	// commits counts number of commits since start
	commits int64

	hooks Hooks

	stopc chan struct{}
	donec chan struct{}
}

type BackendConfig struct {
	// Path is the file path to the backend file.
	Path string

	// BatchInterval is the maximum time before flushing the BatchTx.
	BatchInterval time.Duration

	// BatchLimit is the maximum puts before flushing the BatchTx.
	BatchLimit int

	// MmapSize is the number of bytes to mmap for the backend.
	MmapSize uint64

	// BackendFreelistType is the backend boltdb's freelist type.
	BackendFreelistType bolt.FreelistType

	// UnsafeNoFsync disables all uses of fsync.
	UnsafeNoFsync bool `json:"unsafe-no-fsync"`

	// Mlock prevents backend database file to be swapped
	Mlock bool

	// Hooks are getting executed during lifecycle of Backend's transactions.
	Hooks Hooks
}

func DefaultBackendConfig() BackendConfig {
	return BackendConfig{
		BatchInterval: defaultBatchInterval,
		BatchLimit:    defaultBatchLimit,
		MmapSize:      initialMmapSize,
	}
}

func New(cfg BackendConfig) Backend {
	return newBackend(cfg)
}

func NewDefaultBackend(path string) Backend {
	bcfg := DefaultBackendConfig()
	bcfg.Path = path
	return newBackend(bcfg)
}

func newBackend(cfg BackendConfig) *backend {
	boltOptions := &bolt.Options{
		NoGrowSync:     cfg.UnsafeNoFsync,
		NoFreelistSync: true,
		FreelistType:   cfg.BackendFreelistType,
		//MmapFlags:       0,
		InitialMmapSize: int(cfg.MmapSize),
		NoSync:          cfg.UnsafeNoFsync,
		Mlock:           cfg.Mlock,
	}
	db, err := bolt.Open(cfg.Path, 0600, boltOptions)
	if err != nil {
		klog.Fatalf(fmt.Sprintf("[newBackend]bolt open file %s err: %v", cfg.Path, err))
	}

	b := &backend{
		db: db,

		batchInterval: cfg.BatchInterval,
		batchLimit:    cfg.BatchLimit,
		mlock:         cfg.Mlock,

		readTx: newReadTx(),

		stopc: make(chan struct{}),
		donec: make(chan struct{}),
	}

	b.batchTx = newBatchTxBuffered(b)
	// We set it after newBatchTxBuffered to skip the 'empty' commit.
	b.hooks = cfg.Hooks

	// INFO: backend 异步批量提交多个写事务请求
	go b.run()
	return b
}

// INFO: 定时任务，每 batchInterval 内去批量提交所有事务
//  etcd 通过合并多个写事务请求，是异步机制定时（默认每隔 100ms）将批量事务一次性提交（pending 事务过多才会触发同步提交），从而大大提高吞吐量
func (b *backend) run() {
	defer close(b.donec)
	tick := time.Tick(b.batchInterval)
	for {
		select {
		case <-tick:
			if b.batchTx.safePending() != 0 {
				b.batchTx.Commit()
			}
		case <-b.stopc:
			b.batchTx.CommitAndStop()
			return
		}
	}

}

func (b *backend) BatchTx() BatchTx {
	return b.batchTx
}

func (b *backend) ReadTx() ReadTx {
	return b.readTx
}

// ConcurrentReadTx INFO: @see https://github.com/etcd-io/etcd/commit/9c82e8c72b96eec1e7667a0e139a07b944c33b75
// ConcurrentReadTx creates and returns a new ReadTx, which:
// A) creates and keeps a copy of backend.readTx.txReadBuffer,
// B) references the boltdb read Tx (and its bucket cache) of current batch interval.
func (b *backend) ConcurrentReadTx() ReadTx {
	panic("ConcurrentReadTx")
}

// INFO: transaction begin
func (b *backend) begin(write bool) *bolt.Tx {
	// 只读锁
	b.RLock()
	tx := b.unsafeBegin(write)
	b.RUnlock()

	size := tx.Size()
	db := tx.DB()
	stats := db.Stats()
	atomic.StoreInt64(&b.size, size)
	atomic.StoreInt64(&b.sizeInUse, size-(int64(stats.FreePageN)*int64(db.Info().PageSize)))
	atomic.StoreInt64(&b.openReadTxN, int64(stats.OpenTxN))

	return tx
}

// INFO: https://github.com/etcd-io/bbolt#managing-transactions-manually
func (b *backend) unsafeBegin(write bool) *bolt.Tx {
	tx, err := b.db.Begin(write)
	if err != nil {
		klog.Fatalf(fmt.Sprintf("[unsafeBegin]boltdb begin transaction for db file %s err: %v", b.db.Path(), err))
	}

	return tx
}

// ForceCommit forces the current batching tx to commit.
func (b *backend) ForceCommit() {
	b.batchTx.Commit()
}

func (b *backend) Close() error {
	close(b.stopc)
	<-b.donec
	return b.db.Close()
}

type Snapshot interface {
	// Size gets the size of the snapshot.
	Size() int64
	// WriteTo writes the snapshot into the given writer.
	WriteTo(w io.Writer) (n int64, err error)
	// Close closes the snapshot.
	Close() error
}

type snapshot struct {
	*bolt.Tx
	stopc chan struct{}
	donec chan struct{}
}

func (b *backend) Snapshot() Snapshot {
	// TODO: 为何先 commit, commit 其实也是 begin transaction
	b.batchTx.Commit()

	// read-only lock
	b.RLock()
	defer b.RUnlock()
	tx, err := b.db.Begin(false) // read-only
	if err != nil {
		klog.Fatalf(fmt.Sprintf("[Snapshot]begin transaction err %v", err))
	}
	stopc, donec := make(chan struct{}), make(chan struct{})
	dbBytes := tx.Size() // returns current database size in bytes as seen by this transaction
	mb := 100 * 1024 * 1024
	klog.Infof(fmt.Sprintf("[Snapshot]db size %d MB", int64(float64(dbBytes)/float64(mb))))

	return &snapshot{
		Tx:    tx,
		stopc: stopc,
		donec: donec,
	}
}

// Close INFO: Close 里去 Rollback
func (s *snapshot) Close() error {
	close(s.stopc)
	<-s.donec

	return s.Tx.Rollback()
}
