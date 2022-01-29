package raft

import (
	"errors"
	pb "k8s-lx1036/k8s/storage/raft/hashicorp/raft/rpc"
	"sync"
)

// MemoryStore implements the LogStore and StableStore interface.
// It should NOT EVER be used for production. It is used only for
// unit tests. Use the MDBStore implementation instead.
type MemoryStore struct {
	sync.RWMutex

	logs  map[uint64]*pb.Log
	kv    map[string][]byte
	kvInt map[string]uint64

	lowIndex  uint64
	highIndex uint64
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		logs:  make(map[uint64]*pb.Log),
		kv:    make(map[string][]byte),
		kvInt: make(map[string]uint64),
	}
}

// FirstIndex implements the LogStore interface.
func (store *MemoryStore) FirstIndex() (uint64, error) {
	store.RLock()
	defer store.RUnlock()
	return store.lowIndex, nil
}

// LastIndex implements the LogStore interface.
func (store *MemoryStore) LastIndex() (uint64, error) {
	store.RLock()
	defer store.RUnlock()
	return store.highIndex, nil
}

// GetLog implements the LogStore interface.
func (store *MemoryStore) GetLog(index uint64, log *pb.Log) error {
	store.RLock()
	defer store.RUnlock()
	l, ok := store.logs[index]
	if !ok {
		return ErrLogNotFound
	}

	log = &pb.Log{
		Index:      l.Index,
		Term:       l.Term,
		Type:       l.Type,
		Data:       l.Data,
		Extensions: l.Extensions,
		AppendedAt: l.AppendedAt,
	}
	return nil
}

// StoreLog implements the LogStore interface.
func (store *MemoryStore) StoreLog(log *pb.Log) error {
	return store.StoreLogs([]*pb.Log{log})
}

// StoreLogs implements the LogStore interface.
func (store *MemoryStore) StoreLogs(logs []*pb.Log) error {
	store.Lock()
	defer store.Unlock()
	for _, l := range logs {
		store.logs[l.Index] = l
		if store.lowIndex == 0 {
			store.lowIndex = l.Index
		}
		if l.Index > store.highIndex {
			store.highIndex = l.Index
		}
	}
	return nil
}

// DeleteRange implements the LogStore interface.
func (store *MemoryStore) DeleteRange(min, max uint64) error {
	store.Lock()
	defer store.Unlock()
	for j := min; j <= max; j++ {
		delete(store.logs, j)
	}
	if min <= store.lowIndex {
		store.lowIndex = max + 1
	}
	if max >= store.highIndex {
		store.highIndex = min - 1
	}
	if store.lowIndex > store.highIndex {
		store.lowIndex = 0
		store.highIndex = 0
	}
	return nil
}

// Set implements the StableStore interface.
func (store *MemoryStore) Set(key []byte, val []byte) error {
	store.Lock()
	defer store.Unlock()
	store.kv[string(key)] = val
	return nil
}

// Get implements the StableStore interface.
func (store *MemoryStore) Get(key []byte) ([]byte, error) {
	store.RLock()
	defer store.RUnlock()
	val := store.kv[string(key)]
	if val == nil {
		return nil, errors.New("not found")
	}
	return val, nil
}

// SetUint64 implements the StableStore interface.
func (store *MemoryStore) SetUint64(key []byte, val uint64) error {
	store.Lock()
	defer store.Unlock()
	store.kvInt[string(key)] = val
	return nil
}

// GetUint64 implements the StableStore interface.
func (store *MemoryStore) GetUint64(key []byte) (uint64, error) {
	store.RLock()
	defer store.RUnlock()
	return store.kvInt[string(key)], nil
}
