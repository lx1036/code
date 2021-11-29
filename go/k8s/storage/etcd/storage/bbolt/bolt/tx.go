package bolt

// txid represents the internal transaction identifier.
type txid uint64

type Tx struct {
	writable bool
	managed  bool
	db       *DB
	meta     *meta
	root     Bucket
	pages    map[pgid]*page
	//stats          TxStats
	commitHandlers []func()

	// WriteFlag specifies the flag for write-related methods like WriteTo().
	// Tx opens the database file with the specified flag to copy the data.
	//
	// By default, the flag is unset, which works well for mostly in-memory
	// workloads. For databases that are much larger than available RAM,
	// set the flag to syscall.O_DIRECT to avoid trashing the page cache.
	WriteFlag int
}

func (tx *Tx) init(db *DB) {
	tx.db = db
	tx.pages = nil

	// Copy the meta page since it can be changed by the writer.
	tx.meta = &meta{}
	db.meta().copy(tx.meta)

	// Copy over the root bucket.
	tx.root = newBucket(tx)
	tx.root.bucket = &bucket{}
	*tx.root.bucket = tx.meta.root // INFO: 每一个 meta page 初始化时就有一个 root bucket

	// Increment the transaction id and add a page cache for writable transactions.
	if tx.writable { // INFO: 开启一个写事务，事务号递增+1
		tx.pages = make(map[pgid]*page)
		tx.meta.txid += txid(1) // 事务id +1
	}
}

func (tx *Tx) CreateBucket(name []byte) (*Bucket, error) {
	return tx.root.CreateBucket(name) // root bucket: {root:3, sequence: 0}
}

func (tx *Tx) Bucket(name []byte) *Bucket {
	return tx.root.Bucket(name)
}

func (tx *Tx) page(id pgid) *page {
	// Check the dirty pages first.
	if tx.pages != nil {
		if p, ok := tx.pages[id]; ok {
			return p
		}
	}

	// Otherwise return directly from the mmap.
	return tx.db.page(id)
}
