package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/utils/types"

	bolt "go.etcd.io/bbolt"
)

var (
	ErrNotFound = fmt.Errorf("not found")
)

type Item struct {
	Pod          *types.PodInfo
	deletionTime *time.Time
}

type MemoryStorage struct {
	sync.RWMutex

	store map[string]*Item
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		store: make(map[string]*Item),
	}
}

func (s *MemoryStorage) Put(key string, value *Item) error {
	s.Lock()
	defer s.Unlock()

	s.store[key] = value
	return nil
}

func (s *MemoryStorage) Get(key string) (*Item, error) {
	s.RLock()
	defer s.RUnlock()

	value, ok := s.store[key]
	if !ok {
		return nil, ErrNotFound
	}
	return value, nil
}

func (s *MemoryStorage) List() ([]*Item, error) {
	s.RLock()
	defer s.RUnlock()

	var items []*Item
	for _, item := range s.store {
		items = append(items, item)
	}

	return items, nil
}

func (s *MemoryStorage) Delete(key string) error {
	s.Lock()
	defer s.Unlock()

	delete(s.store, key)
	return nil
}

type DiskStorage struct {
	name   string
	db     *bolt.DB
	memory *MemoryStorage
}

func NewDiskStorage(name string, path string) (Storage, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}

	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	s := &DiskStorage{
		db:     db,
		name:   name,
		memory: NewMemoryStorage(),
	}
	if err = s.load(); err != nil {
		return nil, err
	}

	return s, nil
}

// load from disk into memory
func (s *DiskStorage) load() error {
	err := s.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(s.name))
		return err
	})
	if err != nil {
		return err
	}

	return s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(s.name))
		return bucket.ForEach(func(k, v []byte) error {
			var item Item
			if err = json.Unmarshal(v, &item); err != nil {
				return err
			}
			return s.memory.Put(string(k), &item)
		})
	})
}

func (s *DiskStorage) Put(key string, value *Item) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	err = s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(s.name))
		return bucket.Put([]byte(key), data)
	})
	if err != nil {
		return err
	}

	return s.memory.Put(key, value)
}

func (s *DiskStorage) Get(key string) (*Item, error) {
	return s.memory.Get(key)
}

func (s *DiskStorage) List() ([]*Item, error) {
	return s.memory.List()
}

func (s *DiskStorage) Delete(key string) error {
	err := s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(s.name))
		return bucket.Delete([]byte(key))
	})
	if err != nil {
		return err
	}

	return s.memory.Delete(key)
}
