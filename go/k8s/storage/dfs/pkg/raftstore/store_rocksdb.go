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

// Get returns the value based on the given key.
func (rs *RocksDBStore) Get(key interface{}) (result interface{}, err error) {
	ro := gorocksdb.NewDefaultReadOptions()
	ro.SetFillCache(false)
	defer ro.Destroy()
	return rs.db.GetBytes(ro, []byte(key.(string)))
}

// Put adds a new key-value pair to the RocksDB.
func (rs *RocksDBStore) Put(key, value interface{}, isSync bool) (result interface{}, err error) {
	wo := gorocksdb.NewDefaultWriteOptions()
	wb := gorocksdb.NewWriteBatch()
	wo.SetSync(isSync)
	defer func() {
		wo.Destroy()
		wb.Destroy()
	}()
	wb.Put([]byte(key.(string)), value.([]byte))
	if err := rs.db.Write(wo, wb); err != nil {
		return nil, err
	}
	result = value
	return result, nil
}

// Del deletes a key-value pair.
func (rs *RocksDBStore) Del(key interface{}, isSync bool) (result interface{}, err error) {
	ro := gorocksdb.NewDefaultReadOptions()
	wo := gorocksdb.NewDefaultWriteOptions()
	wb := gorocksdb.NewWriteBatch()
	wo.SetSync(isSync)
	defer func() {
		wo.Destroy()
		ro.Destroy()
		wb.Destroy()
	}()
	slice, err := rs.db.Get(ro, []byte(key.(string)))
	if err != nil {
		return
	}
	result = slice.Data()
	err = rs.db.Delete(wo, []byte(key.(string)))
	return
}

// DeleteKeyAndPutIndex deletes the key-value pair based on the given key and put other keys in the cmdMap to RocksDB.
// TODO explain
func (rs *RocksDBStore) DeleteKeyAndPutIndex(key string, cmdMap map[string][]byte, isSync bool) error {
	wo := gorocksdb.NewDefaultWriteOptions()
	wo.SetSync(isSync)
	wb := gorocksdb.NewWriteBatch()
	defer func() {
		wo.Destroy()
		wb.Destroy()
	}()
	wb.Delete([]byte(key))
	for otherKey, value := range cmdMap {
		if otherKey == key {
			continue
		}
		wb.Put([]byte(otherKey), value)
	}

	if err := rs.db.Write(wo, wb); err != nil {
		err = fmt.Errorf("action[deleteFromRocksDB],err:%v", err)
		return err
	}
	return nil
}

// BatchPut puts the key-value pairs in batch.
func (rs *RocksDBStore) BatchPut(cmdMap map[string][]byte, isSync bool) error {
	wo := gorocksdb.NewDefaultWriteOptions()
	wo.SetSync(isSync)
	wb := gorocksdb.NewWriteBatch()
	defer func() {
		wo.Destroy()
		wb.Destroy()
	}()
	for key, value := range cmdMap {
		wb.Put([]byte(key), value)
	}
	if err := rs.db.Write(wo, wb); err != nil {
		err = fmt.Errorf("action[batchPutToRocksDB],err:%v", err)
		return err
	}
	return nil
}

// ReleaseSnapshot releases the snapshot and its resources.
func (rs *RocksDBStore) ReleaseSnapshot(snapshot *gorocksdb.Snapshot) {
	rs.db.ReleaseSnapshot(snapshot)
}

// RocksDBSnapshot returns the RocksDB snapshot.
func (rs *RocksDBStore) RocksDBSnapshot() *gorocksdb.Snapshot {
	return rs.db.NewSnapshot()
}

// Iterator returns the iterator of the snapshot.
func (rs *RocksDBStore) Iterator(snapshot *gorocksdb.Snapshot) *gorocksdb.Iterator {
	ro := gorocksdb.NewDefaultReadOptions()
	ro.SetFillCache(false)
	ro.SetSnapshot(snapshot)

	return rs.db.NewIterator(ro)
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
