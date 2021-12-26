package raftstore

import (
	bolt "go.etcd.io/bbolt"
	"k8s.io/klog/v2"
)

const (
	DefaultBucket = "default"
)

type BoltdbStore struct {
	*bolt.DB
}

func (store *BoltdbStore) Get(key []byte) ([]byte, error) {
	var value []byte
	err := store.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(DefaultBucket))
		if err != nil {
			return err
		}
		value = bucket.Get(key)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (store *BoltdbStore) Put(key, value []byte) error {
	err := store.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(DefaultBucket))
		if err != nil {
			return err
		}
		return bucket.Put(key, value)
	})

	return err
}

func (store *BoltdbStore) Delete(key []byte) error {
	err := store.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(DefaultBucket))
		if err != nil {
			return err
		}
		return bucket.Delete(key)
	})

	return err
}

func (store *BoltdbStore) Close() error {
	return store.DB.Close()
}

// NewBoltdbStore "./raft/my.db"
func NewBoltdbStore(dbPath string) *BoltdbStore {
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		klog.Fatalf("init boltdb failed: %v, path: %v", err, dbPath)
	}

	return &BoltdbStore{DB: db}
}
