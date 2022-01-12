package bolt_store

import (
	"encoding/json"
	"errors"
	"io/fs"
	"strconv"

	"github.com/hashicorp/raft"
	bolt "go.etcd.io/bbolt"
)

// INFO: https://github.com/hashicorp/raft-boltdb/blob/master/v2/README.md

var (
	// Bucket names we perform transactions in
	dbLogs = []byte("logs")
	dbConf = []byte("conf")

	// ErrKeyNotFound INFO: @see https://github.com/hashicorp/raft/blob/v1.3.3/api.go#L480-L484
	ErrKeyNotFound = errors.New("not found")
)

// BoltStore provides access to bbolt for Raft to store and retrieve log entries.
//  INFO: 内存 store @see github.com/hashicorp/raft@v1.3.3/inmem_store.go
type BoltStore struct {
	db *bolt.DB

	// The path to the Bolt database file
	path string
}

type Options struct {
	Path string

	BoltOptions *bolt.Options

	// NoSync causes the database to skip fsync calls after each
	// write to the log. This is unsafe, so it should be used
	// with caution.
	NoSync bool
}

func New(options Options) (*BoltStore, error) {
	db, err := bolt.Open(options.Path, fs.ModePerm, options.BoltOptions)
	if err != nil {
		return nil, err
	}

	db.NoSync = options.NoSync

	store := &BoltStore{
		db:   db,
		path: options.Path,
	}

	// If the store was opened read-only, don't try and create buckets
	if !options.readOnly() {
		// Set up our buckets
		if err := store.initialize(); err != nil {
			_ = store.Close()
			return nil, err
		}
	}

	return store, nil
}

// NewBoltStore takes a file path and returns a connected Raft backend.
func NewBoltStore(path string) (*BoltStore, error) {
	return New(Options{Path: path})
}

func (o *Options) readOnly() bool {
	return o != nil && o.BoltOptions != nil && o.BoltOptions.ReadOnly
}

// 创建 logs/conf bucket
func (b *BoltStore) initialize() error {
	return b.db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(dbLogs); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(dbConf); err != nil {
			return err
		}

		return nil
	})
}

// Close is used to gracefully close the DB connection.
func (b *BoltStore) Close() error {
	return b.db.Close()
}

/////////////////////////LogStore interface////////////////////////////////////////

// StoreLog is used to store a single raft log
func (b *BoltStore) StoreLog(log *raft.Log) error {
	return b.StoreLogs([]*raft.Log{log})
}

// StoreLogs is used to store a set of raft logs
func (b *BoltStore) StoreLogs(logs []*raft.Log) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		for _, log := range logs {
			key := []byte(strconv.FormatUint(log.Index, 10))
			val, err := json.Marshal(log)
			if err != nil {
				return err
			}
			if err := tx.Bucket(dbLogs).Put(key, val); err != nil {
				return err
			}
		}

		return nil
	})
}

// GetLog is used to retrieve a log from bbolt at a given index.
func (b *BoltStore) GetLog(idx uint64, log *raft.Log) error {
	return b.db.View(func(tx *bolt.Tx) error {
		val := tx.Bucket(dbLogs).Get([]byte(strconv.FormatUint(idx, 10)))
		if val == nil {
			// INFO: raft.ErrLogNotFound 会触发 snapshot，表示 follower log entry 离 leader too far behind
			//  所以需要 leader 发送 snapshot 给 follower
			//  @see https://github.com/hashicorp/raft/blob/v1.3.3/replication.go#L219-L222
			return raft.ErrLogNotFound
		}

		return json.Unmarshal(val, log)
	})
}

// FirstIndex returns the first known index from the Raft log.
func (b *BoltStore) FirstIndex() (uint64, error) {
	var firstIndex []byte
	_ = b.db.View(func(tx *bolt.Tx) error {
		cur := tx.Bucket(dbLogs).Cursor()
		firstIndex, _ = cur.First()
		return nil
	})
	if firstIndex == nil {
		return 0, nil
	} else {
		return strconv.ParseUint(string(firstIndex), 10, 64)
	}
}

// LastIndex returns the last known index from the Raft log.
func (b *BoltStore) LastIndex() (uint64, error) {
	var lastIndex []byte
	_ = b.db.View(func(tx *bolt.Tx) error {
		cur := tx.Bucket(dbLogs).Cursor()
		lastIndex, _ = cur.Last()
		return nil
	})
	if lastIndex == nil {
		// INFO: @see https://github.com/hashicorp/raft/blob/v1.3.3/api.go#L486-L490
		return 0, nil
	} else {
		return strconv.ParseUint(string(lastIndex), 10, 64)
	}
}

// DeleteRange INFO: compact logs in [min, max) after snapshot @see https://github.com/hashicorp/raft/blob/v1.3.3/snapshot.go#L243-L246
func (b *BoltStore) DeleteRange(min, max uint64) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		minKey := []byte(strconv.FormatUint(min, 10))
		cur := tx.Bucket(dbLogs).Cursor()
		for k, _ := cur.Seek(minKey); k != nil; k, _ = cur.Next() {
			key, _ := strconv.ParseUint(string(k), 10, 64)
			if key > max {
				break
			}

			if err := cur.Delete(); err != nil {
				return err
			}
		}

		return nil
	})
}

/////////////////////////LogStore interface////////////////////////////////////////

/////////////////////////StableStore interface////////////////////////////////////////
// 持久化 term,candidate

// Set is used to set a key/value set outside of the raft log
func (b *BoltStore) Set(key []byte, val []byte) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(dbConf).Put(key, val)
	})
}

func (b *BoltStore) Get(key []byte) ([]byte, error) {
	var value []byte
	err := b.db.View(func(tx *bolt.Tx) error {
		value = tx.Bucket(dbConf).Get(key)
		if value == nil {
			// INFO: @see https://github.com/hashicorp/raft/blob/v1.3.3/raft.go#L1502-L1512
			//  @see https://github.com/hashicorp/raft/blob/v1.3.3/api.go#L480-L484
			return ErrKeyNotFound
		}
		return nil
	})

	return value, err
}

func (b *BoltStore) SetUint64(key []byte, val uint64) error {
	return b.Set(key, []byte(strconv.FormatUint(val, 10)))
}

func (b *BoltStore) GetUint64(key []byte) (uint64, error) {
	val, err := b.Get(key)
	if err != nil {
		return 0, err
	}

	return strconv.ParseUint(string(val), 10, 64)
}

/////////////////////////StableStore interface////////////////////////////////////////
