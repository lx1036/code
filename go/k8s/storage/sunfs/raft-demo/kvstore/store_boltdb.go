package main

import (
	bolt "go.etcd.io/bbolt"
	"k8s.io/klog/v2"
)

const (
	DefaultBucket = "default"
)

type Store struct {
	*bolt.DB
}

func (store *Store) Get(key []byte) ([]byte, error) {
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

func (store *Store) Put(key, value []byte) error {
	err := store.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(DefaultBucket))
		if err != nil {
			return err
		}
		return bucket.Put(key, value)
	})

	return err
}

func (store *Store) Delete(key []byte) error {
	err := store.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(DefaultBucket))
		if err != nil {
			return err
		}
		return bucket.Delete(key)
	})

	return err
}

func (store *Store) Close() {
	store.DB.Close()
}

func newStore(dbPath string) *Store {
	db, err := bolt.Open(dbPath, 0666, nil)
	if err != nil {
		klog.Fatalf("init boltdb failed: %v, path: %v", err, dbPath)
	}
	//defer db.Close()

	return &Store{DB: db}
}
