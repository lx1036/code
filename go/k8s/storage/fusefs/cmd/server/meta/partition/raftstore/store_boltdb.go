package raftstore

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/tiglabs/raft/proto"
	bolt "go.etcd.io/bbolt"
	"k8s.io/klog/v2"
)

const DBFilename = "db"
const BucketName = "volumes"

type BoltDBStore struct {
	*bolt.DB
}

func NewBoltDBStore(storeDir string) (*BoltDBStore, error) {
	storeDir, err := filepath.Abs(storeDir)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(storeDir); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(storeDir, 0777)
		} else {
			return nil, err
		}
	}
	dbFile := filepath.Join(storeDir, DBFilename)
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(BucketName))
		if err != nil && err != bolt.ErrBucketExists {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	store := &BoltDBStore{
		DB: db,
	}

	return store, nil
}

func (store *BoltDBStore) Close() error {
	return store.DB.Close()
}

// SeekForPrefix seeks for the place where the prefix is located in the snapshots.
func (store *BoltDBStore) SeekForPrefix(prefix []byte) (result map[string][]byte, err error) {
	result = make(map[string][]byte)
	err = store.View(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(BucketName)).ForEach(func(key, value []byte) error {
			if strings.HasPrefix(string(key), string(prefix)) {
				result[string(key)] = value
			}

			return nil
		})
	})

	return result, err
}

func (store *BoltDBStore) Put(key, value []byte) (err error) {
	return store.Batch(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))
		if bucket == nil {
			klog.Errorf(fmt.Sprintf("can't put key-value from bucket %s", BucketName))
		}

		return bucket.Put(key, value)
	})
}

func (store *BoltDBStore) BatchPut(cmdMap map[string][]byte) error {
	return store.Batch(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))
		if bucket == nil {
			klog.Errorf(fmt.Sprintf("can't batch put key-value from bucket %s", BucketName))
		}

		for key, value := range cmdMap {
			err := bucket.Put([]byte(key), value)
			if err != nil {
				klog.Errorf(fmt.Sprintf("batch put key:%s value:%s err:%v", key, value, err))
				return err
			}
		}
		return nil
	})
}

func (store *BoltDBStore) Del(key []byte) (err error) {
	err = store.Batch(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(BucketName)).Delete(key)
	})
	return err
}

func (store *BoltDBStore) Get(key []byte) (result []byte, err error) {
	err = store.View(func(tx *bolt.Tx) error {
		result = tx.Bucket([]byte(BucketName)).Get(key)
		return nil
	})

	return result, err
}

func (store *BoltDBStore) DeleteKeyAndPutIndex(key string, cmdMap map[string][]byte) error {
	return store.Batch(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))
		bucket.Delete([]byte(key))
		for k, value := range cmdMap {
			if k == key {
				continue
			}
			err := bucket.Put([]byte(k), value)
			if err != nil {
				klog.Errorf(fmt.Sprintf("batch put key:%s value:%s err:%v", k, value, err))
			}
		}

		return nil
	})
}

// RaftCmd defines the Raft commands.
// TODO: extract RaftCmd(in master/raftstore pkg) to common proto pkg
type RaftCmd struct {
	Op uint32 `json:"op"`
	K  string `json:"k"`
	V  []byte `json:"v"`
}

func (store *BoltDBStore) ApplySnapshot(iterator proto.SnapIterator) error {
	return store.Batch(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))

		var err error
		var data []byte
		for err == nil {
			if data, err = iterator.Next(); err != nil {
				break // read until end of buffer
			}
			cmd := &RaftCmd{}
			if err = json.Unmarshal(data, cmd); err != nil {
				continue // no break, skip bad data
			}
			if err = bucket.Put([]byte(cmd.K), cmd.V); err != nil {
				continue
			}
		}
		if err != nil && err != io.EOF {
			return err
		}

		return nil
	})
}
