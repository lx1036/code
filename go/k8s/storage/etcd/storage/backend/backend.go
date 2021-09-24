package backend

import (
	"fmt"
	"k8s.io/klog/v2"
	"sync"
	"sync/atomic"
	"time"

	bolt "go.etcd.io/bbolt"
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

	//Snapshot() Snapshot

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

func (b *backend) BatchTx() BatchTx {
	return b.batchTx
}

func (b *backend) ReadTx() ReadTx {
	return b.readTx
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

		readTx: &readTx{
			baseReadTx: baseReadTx{
				buf: txReadBuffer{
					txBuffer: txBuffer{
						buckets: make(map[BucketID]*bucketBuffer),
					},
					bufVersion: 0,
				},
				buckets: make(map[BucketID]*bolt.Bucket),
			},
		},

		stopc: make(chan struct{}),
		donec: make(chan struct{}),
	}

	b.batchTx = newBatchTxBuffered(b)
	// We set it after newBatchTxBuffered to skip the 'empty' commit.
	b.hooks = cfg.Hooks

	go b.run()
	return b
}

// ConcurrentReadTx INFO: @see https://github.com/etcd-io/etcd/commit/9c82e8c72b96eec1e7667a0e139a07b944c33b75
// ConcurrentReadTx creates and returns a new ReadTx, which:
// A) creates and keeps a copy of backend.readTx.txReadBuffer,
// B) references the boltdb read Tx (and its bucket cache) of current batch interval.
/*func (b *backend) ConcurrentReadTx() ReadTx {

}*/

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

func (b *backend) Close() error {
	close(b.stopc)
	<-b.donec
	return b.db.Close()
}
