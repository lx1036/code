package master

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"

	"k8s-lx1036/k8s/storage/fusefs/pkg/raftstore"

	"k8s.io/klog/v2"
)

// IDAllocator generates and allocates ids
type IDAllocator struct {
	metaPartitionID uint64
	commonID        uint64
	store           *raftstore.RocksDBStore
	partition       raftstore.Partition
	mpIDLock        sync.RWMutex
	metaNodeIDLock  sync.RWMutex
}

func (alloc *IDAllocator) allocateCommonID() (id uint64, err error) {
	alloc.metaNodeIDLock.Lock()
	defer alloc.metaNodeIDLock.Unlock()

	var cmd []byte
	metadata := new(RaftCmd)
	metadata.Op = opSyncAllocCommonID
	metadata.K = maxCommonIDKey
	id = atomic.LoadUint64(&alloc.commonID) + 1
	value := strconv.FormatUint(uint64(id), 10)
	metadata.V = []byte(value)
	cmd, err = json.Marshal(metadata)
	if err != nil {
		klog.Error(err)
		return 0, err
	}

	// 向 raft log 提交cmd
	if _, err = alloc.partition.Submit(cmd); err != nil {
		klog.Errorf("action[allocateCommonID] Submit cmd err: %v", err)
		return 0, err
	}

	alloc.setCommonID(id)

	return id, nil
}

func (alloc *IDAllocator) setCommonID(id uint64) {
	atomic.StoreUint64(&alloc.commonID, id)
}

func (alloc *IDAllocator) restore() {
	alloc.restoreMaxMetaPartitionID()
	alloc.restoreMaxCommonID()
}

func (alloc *IDAllocator) restoreMaxMetaPartitionID() {
	value, err := alloc.store.Get(maxMetaPartitionIDKey)
	if err != nil {
		panic(fmt.Sprintf("Failed to restore maxPartitionID,err:%v ", err.Error()))
	}
	bytes := value.([]byte)
	if len(bytes) == 0 {
		alloc.metaPartitionID = 0
		return
	}
	maxPartitionID, err := strconv.ParseUint(string(bytes), 10, 64)
	if err != nil {
		panic(fmt.Sprintf("Failed to restore maxPartitionID,err:%v ", err.Error()))
	}
	alloc.metaPartitionID = maxPartitionID
	klog.Infof("action[restoreMaxMetaPartitionID] maxMpID[%v]", alloc.metaPartitionID)
}

// The data node, meta node, and node set share the same ID allocator.
func (alloc *IDAllocator) restoreMaxCommonID() {
	value, err := alloc.store.Get(maxCommonIDKey)
	if err != nil {
		panic(fmt.Sprintf("Failed to restore maxCommonID,err:%v ", err.Error()))
	}
	bytes := value.([]byte)
	if len(bytes) == 0 {
		alloc.commonID = 0
		return
	}
	maxMetaNodeID, err := strconv.ParseUint(string(bytes), 10, 64)
	if err != nil {
		panic(fmt.Sprintf("Failed to restore maxCommonID,err:%v ", err.Error()))
	}
	alloc.commonID = maxMetaNodeID
	klog.Infof("action[restoreMaxCommonID] maxMnID[%v]", alloc.commonID)
}

func NewIDAllocator(store *raftstore.RocksDBStore, partition raftstore.Partition) *IDAllocator {
	return &IDAllocator{
		store:     store,
		partition: partition,
	}
}
