package bolt

import (
	"fmt"
	"hash/fnv"
	"os"
	"sync"
	"time"
	"unsafe"

	"k8s.io/klog/v2"
)

// maxMapSize represents the largest mmap size supported by Bolt.
const maxMapSize = 0xFFFFFFFFFFFF // 256TB

// Represents a marker value to indicate that a file is a Bolt DB.
const magic uint32 = 0xED0CDAED

// The data file format version.
const version = 2

type meta struct {
	magic    uint32
	version  uint32
	pageSize uint32
	flags    uint32
	root     bucket
	freelist pgid
	pgid     pgid
	txid     txid
	checksum uint64
}

// validate checks the marker bytes and version of the meta page to ensure it matches this binary.
func (m *meta) validate() error {
	if m.magic != magic {
		return ErrInvalid
	} else if m.version != version {
		return ErrVersionMismatch
	} else if m.checksum != 0 && m.checksum != m.sum64() {
		return ErrChecksum
	}
	return nil
}
func (m *meta) sum64() uint64 {
	var h = fnv.New64a()
	_, _ = h.Write((*[unsafe.Offsetof(meta{}.checksum)]byte)(unsafe.Pointer(m))[:])
	return h.Sum64()
}

// Options represents the options that can be set when opening a database.
type Options struct {
	opened bool

	// Timeout is the amount of time to wait to obtain a file lock.
	// When set to zero it will wait indefinitely. This option is only
	// available on Darwin and Linux.
	Timeout time.Duration

	// Sets the DB.NoGrowSync flag before memory mapping the file.
	NoGrowSync bool
	NoSync     bool

	// FreelistType sets the backend freelist type. There are two options. Array which is simple but endures
	// dramatic performance degradation if database is large and framentation in freelist is common.
	// The alternative one is using hashmap, it is faster in almost all circumstances
	// but it doesn't guarantee that it offers the smallest page id available. In normal case it is safe.
	// The default type is array
	FreelistType FreelistType

	OpenFile func(string, int, os.FileMode) (*os.File, error)

	MmapFlags      int
	NoFreelistSync bool
	ReadOnly       bool
}

// Stats represents statistics about the database.
type Stats struct {
	// Freelist stats
	FreePageN     int // total number of free pages on the freelist
	PendingPageN  int // total number of pending pages on the freelist
	FreeAlloc     int // total bytes allocated in free pages
	FreelistInuse int // total bytes used by the freelist

	// Transaction stats
	TxN     int // total number of started read transactions
	OpenTxN int // number of currently open read transactions

	TxStats TxStats // global, ongoing stats.
}

// DB represents a collection of buckets persisted to a file on disk.
// All data access is performed through transactions which can be obtained through the DB.
// All the functions on DB will return a ErrDatabaseNotOpen if accessed before Open() is called.
type DB struct {
	// basic
	opened     bool
	NoSync     bool
	NoGrowSync bool
	MmapFlags  int

	// data
	dataref []byte // mmap'ed readonly, write throws SEGV
	data    *[maxMapSize]byte
	datasz  int

	// file 相关字段
	file          *os.File
	openFile      func(string, int, os.FileMode) (*os.File, error)
	path          string
	MaxBatchDelay time.Duration
	MaxBatchSize  int
	AllocSize     int
	ops           struct {
		writeAt func(b []byte, off int64) (n int, err error)
	}

	// lock 相关字段
	rwlock   sync.Mutex // Allows only one writer at a time.
	metalock sync.Mutex // Protects meta page access.
	readOnly bool

	// transaction 相关字段
	rwtx *Tx

	// meta
	meta0 *meta
	meta1 *meta

	// b+tree
	freelistLoad   sync.Once
	FreelistType   FreelistType
	NoFreelistSync bool
	freelist       *freelist
	pageSize       int

	// statistics
	stats Stats
}

// Begin starts a new transaction.
// Multiple read-only transactions can be used concurrently but only one
// write transaction can be used at a time. Starting multiple write transactions
// will cause the calls to block and be serialized until the current write
// transaction finishes.
//
// Transactions should not be dependent on one another. Opening a read
// transaction and a write transaction in the same goroutine can cause the
// writer to deadlock because the database periodically needs to re-mmap itself
// as it grows and it cannot do that while a read transaction is open.
//
// If a long running read transaction (for example, a snapshot transaction) is
// needed, you might want to set DB.InitialMmapSize to a large enough value
// to avoid potential blocking of write transaction.
//
// IMPORTANT: You must close read-only transactions after you are finished or
// else the database will not reclaim old pages.
func (db *DB) Begin(writable bool) (*Tx, error) {
	if writable {
		return db.beginRWTx()
	}

	return nil, nil
	//return db.beginTx()
}

func (db *DB) beginRWTx() (*Tx, error) {

	// Obtain writer lock. This is released by the transaction when it closes.
	// This enforces only one writer transaction at a time.
	db.rwlock.Lock()
	// Once we have the writer lock then we can lock the meta pages so that
	// we can set up the transaction.
	db.metalock.Lock()
	defer db.metalock.Unlock()

	// Create a transaction associated with the database.
	transaction := &Tx{writable: true}
	transaction.init(db)
	db.rwtx = transaction
	//db.freePages()

	return transaction, nil
}

// Update executes a function within the context of a read-write managed transaction.
// If no error is returned from the function then the transaction is committed.
// If an error is returned then the entire transaction is rolled back.
// Any error that is returned from the function or returned from the commit is
// returned from the Update() method.
//
// Attempting to manually commit or rollback within the function will cause a panic.
func (db *DB) Update(fn func(*Tx) error) error {
	t, err := db.Begin(true)
	if err != nil {
		return err
	}

	// Mark as a managed tx so that the inner function cannot manually commit.
	t.managed = true

	// If an error is returned from the function then rollback and return error.
	err = fn(t)
	t.managed = false
	if err != nil {
		//_ = t.Rollback()
		return err
	}

	return t.Commit()
}

// meta retrieves the current meta page reference.
func (db *DB) meta() *meta {
	// We have to return the meta with the highest txid which doesn't fail
	// validation. Otherwise, we can cause errors when in fact the database is
	// in a consistent state. metaA is the one with the higher txid.
	// 这里metaA使用更高的meta
	metaA := db.meta0
	metaB := db.meta1
	if db.meta1.txid > db.meta0.txid {
		metaA = db.meta1
		metaB = db.meta0
	}

	// Use higher meta page if valid. Otherwise fallback to previous, if valid.
	if err := metaA.validate(); err == nil {
		return metaA
	} else if err := metaB.validate(); err == nil {
		return metaB
	}

	// This should never be reached, because both meta1 and meta0 were validated
	// on mmap() and we do fsync() on every write.
	panic("bolt.DB.meta(): invalid meta pages")
}
func (db *DB) hasSyncedFreelist() bool {
	return db.meta().freelist != pgidNoFreelist
}

// loadFreelist reads tFreelistArrayTypehe freelist if it is synced, or reconstructs it
// by scanning the DB if it is not synced. It assumes there are no
// concurrent accesses being made to the freelist.
func (db *DB) loadFreelist() {
	db.freelistLoad.Do(func() {
		//db.freelist = newFreelist(db.FreelistType)
		db.freelist = newFreelist()
		if !db.hasSyncedFreelist() {
			// Reconstruct free list by scanning the DB.
			//db.freelist.readIDs(db.freepages())
		} else {
			// Read free list from freelist page.
			//db.freelist.read(db.page(db.meta().freelist))
		}
		db.stats.FreePageN = db.freelist.free_count()
	})
}
func (db *DB) close() error {
	if !db.opened {
		return nil
	}

	db.opened = false
	db.freelist = nil
	// Clear ops.
	db.ops.writeAt = nil
	// Close the mmap.
	if err := db.munmap(); err != nil {
		return err
	}

	// Close file handles.
	if db.file != nil {
		// No need to unlock read-only file.
		if !db.readOnly {
			// Unlock the file.
			if err := funlock(db); err != nil {
				klog.Infof("bolt.Close(): funlock error: %s", err)
			}
		}

		// Close the file descriptor.
		if err := db.file.Close(); err != nil {
			return fmt.Errorf("db file close: %s", err)
		}
		db.file = nil
	}

	db.path = ""
	return nil
}

// munmap unmaps the data file from memory.
func (db *DB) munmap() error {
	if err := munmap(db); err != nil {
		return fmt.Errorf("unmap error: " + err.Error())
	}
	return nil
}

// Path returns the path to currently open database file.
func (db *DB) Path() string {
	return db.path
}

// page retrieves a page reference from the mmap based on the current page size.
func (db *DB) page(id pgid) *page {
	pos := id * pgid(db.pageSize)
	return (*page)(unsafe.Pointer(&db.data[pos]))
}

// DefaultOptions represent the options used if nil options are passed into Open().
// No timeout is used which will cause Bolt to wait indefinitely for a lock.
var DefaultOptions = &Options{
	Timeout:      0,
	NoGrowSync:   false,
	FreelistType: FreelistArrayType,
}

// Default values if not set in a DB instance.
const (
	DefaultMaxBatchSize  int = 1000
	DefaultMaxBatchDelay     = 10 * time.Millisecond
	DefaultAllocSize         = 16 * 1024 * 1024
)

// 打开一个文件用来保存数据，如果不存在则创建
func Open(path string, mode os.FileMode, options *Options) (*DB, error) {
	db := &DB{
		opened: true,
	}
	// Set default options if no options are provided.
	if options == nil {
		options = DefaultOptions
	}
	db.NoSync = options.NoSync
	db.NoGrowSync = options.NoGrowSync
	db.MmapFlags = options.MmapFlags
	db.NoFreelistSync = options.NoFreelistSync
	db.FreelistType = options.FreelistType

	// Set default values for later DB operations.
	db.MaxBatchSize = DefaultMaxBatchSize
	db.MaxBatchDelay = DefaultMaxBatchDelay
	db.AllocSize = DefaultAllocSize

	flag := os.O_RDWR
	if options.ReadOnly {
		flag = os.O_RDONLY
		db.readOnly = true
	}

	db.openFile = options.OpenFile
	if db.openFile == nil {
		db.openFile = os.OpenFile
	}

	// Open data file and separate sync handler for metadata writes.
	var err error
	if db.file, err = db.openFile(path, flag|os.O_CREATE, mode); err != nil {
		_ = db.close()
		return nil, err
	}
	db.path = db.file.Name()

	// 获取 file lock
	if err := flock(db, !db.readOnly, options.Timeout); err != nil {
		_ = db.close()
		return nil, err
	}

	if db.readOnly {
		return db, nil
	}

	db.loadFreelist()

	// Mark the database as opened and return.
	return db, nil
}
