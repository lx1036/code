package raftstore

import (
	"errors"
	"fmt"
	"os"

	"github.com/tecbot/gorocksdb"
)

// RocksDBStore is a wrapper of the gorocksdb.DB
type RocksDBStore struct {
	dir string
	db  *gorocksdb.DB
}

// Open opens the RocksDB instance.
func (rs *RocksDBStore) Open(lruCacheSize, writeBufferSize int) error {
	basedTableOptions := gorocksdb.NewDefaultBlockBasedTableOptions()
	basedTableOptions.SetBlockCache(gorocksdb.NewLRUCache(uint64(lruCacheSize)))
	opts := gorocksdb.NewDefaultOptions()
	opts.SetBlockBasedTableFactory(basedTableOptions)
	opts.SetCreateIfMissing(true)
	opts.SetWriteBufferSize(writeBufferSize)
	opts.SetMaxWriteBufferNumber(2)
	opts.SetCompression(gorocksdb.NoCompression)
	db, err := gorocksdb.OpenDb(opts, rs.dir)
	if err != nil {
		err = fmt.Errorf("action[openRocksDB],err:%v", err)
		return err
	}
	rs.db = db
	return nil
}

// NewRocksDBStore returns a new RocksDB instance.
func NewRocksDBStore(dir string, lruCacheSize, writeBufferSize int) (store *RocksDBStore,
	err error) {
	var fi os.FileInfo
	if fi, err = os.Stat(dir); err != nil {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return
		}
	} else {
		if !fi.IsDir() {
			return nil, errors.New(dir + " is not a directory")
		}
	}
	store = &RocksDBStore{dir: dir}
	err = store.Open(lruCacheSize, writeBufferSize)
	if err != nil {
		return
	}

	return
}
