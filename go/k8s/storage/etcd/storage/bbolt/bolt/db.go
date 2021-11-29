package bolt

import (
	"fmt"
	"k8s.io/klog/v2"
	"os"
	"sync"
	"time"
	"unsafe"
)

type FreelistType string

const (
	// FreelistArrayType indicates backend freelist type is array
	FreelistArrayType = FreelistType("array")
	// FreelistMapType indicates backend freelist type is hashmap
	FreelistMapType = FreelistType("hashmap")
)

// default page size for db is set to the OS page size.
var defaultPageSize = os.Getpagesize()

type Options struct {
	Timeout time.Duration // wait for file lock

	NoGrowSync bool

	NoFreelistSync bool

	FreelistType FreelistType

	ReadOnly bool

	PageSize int

	InitialMmapSize int

	MmapFlags int
}

var DefaultOptions = &Options{
	Timeout:      0,
	NoGrowSync:   false,
	FreelistType: FreelistArrayType,
}

type DB struct {
	opened   bool
	mmaplock sync.RWMutex // Protects mmap access during remapping

	// file
	readOnly bool
	path     string
	file     *os.File
	filesz   int // current on disk file size
	pageSize int
	ops      struct {
		writeAt func(b []byte, off int64) (n int, err error)
	}
	MmapFlags int
	dataref   []byte // mmap'ed readonly, write throws SEGV 指向映射到内存那块数据
	data      *[maxMapSize]byte
	datasz    int

	// page
	pagePool sync.Pool
	meta0    *meta
	meta1    *meta
}

func Open(path string, mode os.FileMode, options *Options) (*DB, error) {
	db := &DB{
		opened: true,
	}
	if options == nil {
		options = DefaultOptions
	}
	//db.NoSync = options.NoSync
	//db.NoGrowSync = options.NoGrowSync
	db.MmapFlags = options.MmapFlags
	//db.NoFreelistSync = options.NoFreelistSync
	//db.FreelistType = options.FreelistType
	//db.Mlock = options.Mlock

	flag := os.O_RDWR
	if options.ReadOnly {
		flag = os.O_RDONLY
		db.readOnly = true
	}

	var err error
	if db.file, err = os.OpenFile(path, flag|os.O_CREATE, mode); err != nil {
		_ = db.close()
		return nil, err
	}
	db.path = db.file.Name()

	// INFO: 文件锁，如果是写事务，则互斥锁；如果是读事务，则共享锁
	if err := flock(db, !db.readOnly, options.Timeout); err != nil {
		_ = db.close()
		return nil, err
	}

	// Default values for test hooks
	db.ops.writeAt = db.file.WriteAt
	if db.pageSize = options.PageSize; db.pageSize == 0 {
		db.pageSize = defaultPageSize // 4 KB
	}

	// Initialize the database if it doesn't exist.
	if info, err := db.file.Stat(); err != nil {
		_ = db.close()
		return nil, err
	} else if info.Size() == 0 {
		// Initialize new files with meta pages.
		if err := db.init(); err != nil {
			// clean up file descriptor on initialization fail
			_ = db.close()
			return nil, err
		}
	} else {
		//var buf [0x1000]byte // 0x1000=2^12=4*1024=4k, len(buf)=4096

	}

	// Initialize page pool.
	db.pagePool = sync.Pool{
		New: func() interface{} {
			return make([]byte, db.pageSize)
		},
	}

	// Memory map the data file.
	if err := db.mmap(options.InitialMmapSize); err != nil {
		_ = db.close()
		return nil, err
	}

	if db.readOnly {
		return db, nil
	}

	// Mark the database as opened and return.
	return db, nil
}

// Represents a marker value to indicate that a file is a Bolt DB.
const magic uint32 = 0xED0CDAED

// The data file format version.
const version = 2

// INFO: 初始化两个 meta page, freelist page, empty leaf page
//  最开始两个 page 是 meta page, pageid={0,1}
//  pageid=2 第三页是 freelist page
//  pageid=3 第四页是 free leaf page
func (db *DB) init() error {
	// Create two meta pages on a buffer.
	buf := make([]byte, db.pageSize*4) // 4个page
	for i := 0; i < 2; i++ {
		p := db.pageInBuffer(buf, pgid(i))
		p.id = pgid(i)
		p.flags = metaPageFlag

		// Initialize the meta page.
		m := p.meta()
		m.magic = magic
		m.version = version
		m.pageSize = uint32(db.pageSize)
		m.freelist = 2
		m.root = bucket{root: 3} // INFO: 创建 meta 会自动创建一个 root bucket, 后续创建的 bucket 都是其 subbucket
		m.pgid = 4
		m.txid = txid(i)
		m.checksum = m.sum64()
	}

	// Write an empty freelist at page 3.
	p := db.pageInBuffer(buf, pgid(2))
	p.id = pgid(2)
	p.flags = freelistPageFlag
	p.count = 0

	// Write an empty leaf page at page 4.
	p = db.pageInBuffer(buf, pgid(3))
	p.id = pgid(3)
	p.flags = leafPageFlag
	p.count = 0

	// INFO: 写buffer到db.file里并落盘
	if _, err := db.ops.writeAt(buf, 0); err != nil {
		return err
	}
	if err := db.file.Sync(); err != nil { // 刷盘
		return err
	}
	db.filesz = len(buf)

	return nil
}

// pageInBuffer retrieves a page reference from a given byte array based on the current page size.
func (db *DB) pageInBuffer(b []byte, id pgid) *page {
	return (*page)(unsafe.Pointer(&b[id*pgid(db.pageSize)]))
}

// mmap opens the underlying memory-mapped file and initializes the meta references.
func (db *DB) mmap(minsz int) error {
	db.mmaplock.Lock()
	defer db.mmaplock.Unlock()

	info, err := db.file.Stat()
	if err != nil {
		return fmt.Errorf("mmap stat error: %s", err)
	} else if int(info.Size()) < db.pageSize*2 {
		return fmt.Errorf("file size too small")
	}

	// Ensure the size is at least the minimum size.
	fileSize := int(info.Size())
	var size = fileSize
	if size < minsz {
		size = minsz
	}
	size, err = db.mmapSize(size)
	if err != nil {
		return err
	}

	// unlock memory lock

	// Memory-map the data file as a byte slice.
	if err := mmap(db, size); err != nil { //INFO: mmap 从磁盘到内存中
		return err
	}

	// 指向两个meta page指针
	db.meta0 = db.page(0).meta()
	db.meta1 = db.page(1).meta()
	err0 := db.meta0.validate()
	err1 := db.meta1.validate()
	if err0 != nil && err1 != nil {
		return err0
	}

	return nil
}

// INFO: 返回pageid指向的内存地址
func (db *DB) page(id pgid) *page {
	pos := id * pgid(db.pageSize)
	return (*page)(unsafe.Pointer(&db.data[pos]))
}

const maxMapSize = 0xFFFFFFFFFFFF // 256TB = 16^12= 2^48 = 1TB * 2^8
const maxMmapStep = 1 << 30       // 1GB

func (db *DB) mmapSize(size int) (int, error) {
	// Double the size from 32KB until 1GB.
	// 如果在 [最少 32KB，最多 1GB=2^30B] 之间
	for i := uint(15); i <= 30; i++ {
		if size <= 1<<i {
			return 1 << i, nil
		}
	}

	if size > maxMapSize {
		return 0, fmt.Errorf("mmap too large")
	}

	// If larger than 1GB then grow by 1GB at a time.
	// 凑整1GB
	sz := int64(size)
	if remainder := sz % int64(maxMmapStep); remainder > 0 {
		sz += int64(maxMmapStep) - remainder
	}

	// 必须是 4KB 的整数倍
	pageSize := int64(db.pageSize)
	if (sz % pageSize) != 0 {
		sz = ((sz / pageSize) + 1) * pageSize
	}
	// If we've exceeded the max size then only grow up to the max size.
	if sz > maxMapSize {
		sz = maxMapSize
	}

	return int(sz), nil
}

// munmap unmaps the data file from memory.
func (db *DB) munmap() error {
	if err := munmap(db); err != nil {
		return fmt.Errorf("unmap error: " + err.Error())
	}
	return nil
}

// INFO: 获取 txid 更高的 meta page
func (db *DB) meta() *meta {
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

	panic("bolt.DB.meta(): invalid meta pages")
}

// INFO: 释放文件锁，close file handler, close mmap等等
func (db *DB) close() error {
	if !db.opened {
		return nil
	}

	db.opened = false
	//db.freelist = nil
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
				klog.Errorf(fmt.Sprintf("bolt.Close(): funlock error: %v", err))
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
