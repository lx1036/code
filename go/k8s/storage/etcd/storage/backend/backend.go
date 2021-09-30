package backend

import (
	"fmt"
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

type Backend struct {
	sync.RWMutex

	batchInterval time.Duration
	batchLimit    int
	batchTx       *batchTxBuffered

	readTx *readTx
	// txReadBufferCache mirrors "txReadBuffer" within "readTx" -- readTx.baseReadTx.buf.
	// When creating "concurrentReadTx":
	// - if the cache is up-to-date, "readTx.baseReadTx.buf" copy can be skipped
	// - if the cache is empty or outdated, "readTx.baseReadTx.buf" copy is required
	txReadBufferCache txReadBufferCache

	db *bolt.DB
	// Size returns current database size in bytes as seen by this transaction.
	size int64
	// sizeInUse is the number of bytes actually used in the Backend
	sizeInUse int64
	// openReadTxN is the number of currently open read transactions in the Backend
	openReadTxN int64

	// mlock prevents Backend database file to be swapped
	mlock bool
	// commits counts number of commits since start
	commits int64

	hooks Hooks

	stopc chan struct{}
	donec chan struct{}
}

type Config struct {
	// Path is the file path to the Backend file.
	Path string

	// BatchInterval is the maximum time before flushing the BatchTx.
	BatchInterval time.Duration

	// BatchLimit is the maximum puts before flushing the BatchTx.
	BatchLimit int

	// MmapSize is the number of bytes to mmap for the Backend.
	MmapSize uint64

	// BackendFreelistType is the Backend boltdb's freelist type.
	BackendFreelistType bolt.FreelistType

	// UnsafeNoFsync disables all uses of fsync.
	UnsafeNoFsync bool `json:"unsafe-no-fsync"`

	// Mlock prevents Backend database file to be swapped
	Mlock bool

	// Hooks are getting executed during lifecycle of Backend's transactions.
	Hooks Hooks
}

func DefaultBackendConfig() Config {
	return Config{
		BatchInterval: defaultBatchInterval,
		BatchLimit:    defaultBatchLimit,
		MmapSize:      initialMmapSize,
	}
}

func New(cfg Config) *Backend {
	return newBackend(cfg)
}

func NewDefaultBackend(path string) *Backend {
	bcfg := DefaultBackendConfig()
	bcfg.Path = path
	return newBackend(bcfg)
}

func newBackend(cfg Config) *Backend {
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

	b := &Backend{
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

	// INFO: Backend 异步批量提交多个写事务请求
	go b.run()
	return b
}

// INFO: 定时任务，每 batchInterval 内去批量提交所有事务
//  etcd 通过合并多个写事务请求，是异步机制定时（默认每隔 100ms）将批量事务一次性提交（pending 事务过多才会触发同步提交），从而大大提高吞吐量
func (b *Backend) run() {
	defer close(b.donec)
	tick := time.Tick(b.batchInterval)
	for {
		select {
		case <-tick:
			if b.batchTx.safePending() != 0 {
				klog.Infof(fmt.Sprintf("[run]batchTx Commit"))
				b.batchTx.Commit()
			}
		case <-b.stopc:
			b.batchTx.CommitAndStop()
			return
		}
	}

}

func (b *Backend) BatchTx() *batchTxBuffered {
	return b.batchTx
}

func (b *Backend) ReadTx() *readTx {
	return b.readTx
}

// ConcurrentReadTx
//  INFO: @see https://github.com/etcd-io/etcd/commit/9c82e8c72b96eec1e7667a0e139a07b944c33b75
// ConcurrentReadTx creates and returns a new ReadTx, which:
// A) creates and keeps a copy of Backend.readTx.txReadBuffer,
// B) references the boltdb read Tx (and its bucket cache) of current batch interval.
func (b *Backend) ConcurrentReadTx() *concurrentReadTx {
	b.readTx.RLock()
	defer b.readTx.RUnlock()
	// prevent boltdb read Tx from been rolled back until store read Tx is done. Needs to be called when holding readTx.RLock().
	b.readTx.txWg.Add(1)

	// TODO: might want to copy the read buffer lazily - create copy when A) end of a write transaction B) end of a batch interval.

	b.txReadBufferCache.mu.Lock()

	currentCache := b.txReadBufferCache.buf
	currentCacheVersion := b.txReadBufferCache.bufVersion
	currentBufVersion := b.readTx.buf.bufVersion

	isEmptyCache := currentCache == nil
	isStaleCache := currentCacheVersion != currentBufVersion

	var buf *txReadBuffer
	switch {
	case isEmptyCache:
		curBuf := b.readTx.buf.unsafeCopy()
		buf = &curBuf
	case isStaleCache:
		// to maximize the concurrency, try unsafe copy of buffer
		// release the lock while copying buffer -- cache may become stale again and
		// get overwritten by someone else.
		// therefore, we need to check the readTx buffer version again
		b.txReadBufferCache.mu.Unlock()
		curBuf := b.readTx.buf.unsafeCopy()
		b.txReadBufferCache.mu.Lock()
		buf = &curBuf
	default:
		// neither empty nor stale cache, just use the current buffer
		buf = currentCache
	}

	if isEmptyCache || currentCacheVersion == b.txReadBufferCache.bufVersion {
		// continue if the cache is never set or no one has modified the cache
		b.txReadBufferCache.buf = buf
		b.txReadBufferCache.bufVersion = currentBufVersion
	}

	b.txReadBufferCache.mu.Unlock()

	// concurrentReadTx is not supposed to write to its txReadBuffer
	return &concurrentReadTx{
		baseReadTx: baseReadTx{
			buf:     *buf,
			tx:      b.readTx.tx,
			buckets: b.readTx.buckets,
			txWg:    b.readTx.txWg,
		},
	}
}

// INFO: transaction begin
func (b *Backend) begin(write bool) *bolt.Tx {
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
func (b *Backend) unsafeBegin(write bool) *bolt.Tx {
	tx, err := b.db.Begin(write)
	if err != nil {
		klog.Fatalf(fmt.Sprintf("[unsafeBegin]boltdb begin transaction for db file %s err: %v", b.db.Path(), err))
	}

	return tx
}

// ForceCommit forces the current batching tx to commit.
func (b *Backend) ForceCommit() {
	b.batchTx.Commit()
}

func (b *Backend) Close() error {
	close(b.stopc)
	<-b.donec
	return b.db.Close()
}
