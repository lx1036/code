package bolt

import (
	"bytes"
	"unsafe"
)

const (
	minFillPercent = 0.1
	maxFillPercent = 1.0
)

// bucket represents the on-file representation of a bucket.
// This is stored as the "value" of a bucket key. If the bucket is small enough,
// then its root page can be stored inline in the "value", after the bucket
// header. In the case of inline buckets, the "root" will be 0.
type bucket struct {
	root     pgid   // page id of the bucket's root-level page
	sequence uint64 // monotonically incrementing, used by NextSequence()
}

// Bucket represents a collection of key/value pairs inside the database.
type Bucket struct {
	*bucket
	tx       *Tx                // the associated transaction
	buckets  map[string]*Bucket // subbucket cache
	page     *page              // inline page reference
	rootNode *node              // materialized node for the root page.
	nodes    map[pgid]*node     // node cache

	// Sets the threshold for filling nodes when they split. By default,
	// the bucket will fill to 50% but it can be useful to increase this
	// amount if you know that your write workloads are mostly append-only.
	//
	// This is non-persisted across transactions so it must be set in every Tx.
	FillPercent float64
}

// ForEach executes a function for each key/value pair in a bucket.
// If the provided function returns an error then the iteration is stopped and
// the error is returned to the caller. The provided function must not modify
// the bucket; this will result in undefined behavior.
func (b *Bucket) ForEach(fn func(k, v []byte) error) error {
	if b.tx.db == nil {
		return ErrTxClosed
	}
	c := b.Cursor()
	for k, v := c.First(); k != nil; k, v = c.Next() {
		if err := fn(k, v); err != nil {
			return err
		}
	}
	return nil
}

// Cursor creates a cursor associated with the bucket.
// The cursor is only valid as long as the transaction is open.
// Do not use a cursor after the transaction is closed.
func (b *Bucket) Cursor() *Cursor {
	// Update transaction statistics.
	b.tx.stats.CursorCount++

	// Allocate and return a cursor.
	return &Cursor{
		bucket: b,
		stack:  make([]elemRef, 0),
	}
}

// Bucket retrieves a nested bucket by name.
// Returns nil if the bucket does not exist.
// The bucket instance is only valid for the lifetime of the transaction.
func (b *Bucket) Bucket(name []byte) *Bucket {
	if b.buckets != nil {
		if child := b.buckets[string(name)]; child != nil {
			return child
		}
	}

	// Move cursor to key.
	c := b.Cursor()
	k, v, flags := c.seek(name)

	// Return nil if the key doesn't exist or it is not a bucket.
	if !bytes.Equal(name, k) || (flags&bucketLeafFlag) == 0 {
		return nil
	}

	// Otherwise create a bucket and cache it.
	var child = b.openBucket(v)
	if b.buckets != nil {
		b.buckets[string(name)] = child
	}

	return child
}

// Helper method that re-interprets a sub-bucket value
// from a parent into a Bucket
func (b *Bucket) openBucket(value []byte) *Bucket {
	panic("not implemented")
}

// CreateBucket creates a new bucket at the given key and returns the new bucket.
// Returns an error if the key already exists, if the bucket name is blank, or if the bucket name is too long.
// The bucket instance is only valid for the lifetime of the transaction.
func (b *Bucket) CreateBucket(key []byte) (*Bucket, error) {
	if b.tx.db == nil {
		return nil, ErrTxClosed
	} else if !b.tx.writable {
		return nil, ErrTxNotWritable
	} else if len(key) == 0 {
		return nil, ErrBucketNameRequired
	}

	// Move cursor to correct position.
	c := b.Cursor()
	k, _, flags := c.seek(key)

	// Return an error if there is an existing key.
	if bytes.Equal(key, k) {
		if (flags & bucketLeafFlag) != 0 {
			return nil, ErrBucketExists
		}
		return nil, ErrIncompatibleValue
	}

	// Create empty, inline bucket.
	var bucket = Bucket{
		bucket:      &bucket{},
		rootNode:    &node{isLeaf: true},
		FillPercent: DefaultFillPercent,
	}
	var value = bucket.write()

	// Insert into node.
	key = cloneBytes(key)
	c.node().put(key, key, value, 0, bucketLeafFlag)

	// Since subbuckets are not allowed on inline buckets, we need to
	// dereference the inline page, if it exists. This will cause the bucket
	// to be treated as a regular, non-inline bucket for the rest of the tx.
	b.page = nil

	return b.Bucket(key), nil
}

// cloneBytes returns a copy of a given slice.
func cloneBytes(v []byte) []byte {
	var clone = make([]byte, len(v))
	copy(clone, v)
	return clone
}

const bucketHeaderSize = int(unsafe.Sizeof(bucket{}))

// write allocates and writes a bucket to a byte slice.
func (b *Bucket) write() []byte {
	// Allocate the appropriate size.
	var n = b.rootNode
	var value = make([]byte, bucketHeaderSize+n.size())

	// Write a bucket header.
	var bucket = (*bucket)(unsafe.Pointer(&value[0]))
	*bucket = *b.bucket

	// Convert byte slice to a fake page and write the root node.
	var p = (*page)(unsafe.Pointer(&value[bucketHeaderSize]))
	n.write(p)

	return value
}

// DefaultFillPercent is the percentage that split pages are filled.
// This value can be changed by setting Bucket.FillPercent.
const DefaultFillPercent = 0.5

// newBucket returns a new bucket associated with a transaction.
func newBucket(tx *Tx) Bucket {
	var b = Bucket{
		tx:          tx,
		FillPercent: DefaultFillPercent,
	}
	if tx.writable {
		b.buckets = make(map[string]*Bucket)
		b.nodes = make(map[pgid]*node)
	}
	return b
}
