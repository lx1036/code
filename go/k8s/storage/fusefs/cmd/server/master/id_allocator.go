package master

import (
	"encoding/json"
	"fmt"
	boltdb "k8s-lx1036/k8s/storage/raft/hashicorp/bolt-store"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hashicorp/raft"
	"k8s.io/klog/v2"
)

const (
	opSyncAllocMetaPartitionID uint32 = 0x06
	opAllocVolumeID            uint32 = 0x07
	opAllocMetaNodeID          uint32 = 0x08
	opSyncBatchPut             uint32 = 0x0E
	opSyncAddBucket            uint32 = 0x0F
	opSyncUpdateBucket         uint32 = 0x10
	opSyncDeleteBucket         uint32 = 0x11
	opSyncAddVolMountClient    uint32 = 0x12
	opSyncUpdateVolMountClient uint32 = 0x13
	opSyncDeleteVolMountClient uint32 = 0x14
)

const (
	maxMetaNodeIDKey      = "#max_meta_node_id"
	maxVolumeIDKey        = "#max_volume_id"
	maxMetaPartitionIDKey = "#max_mp_id"
)

// IDAllocator generates and allocates ids
type IDAllocator struct {
	metaNodeLock        sync.RWMutex
	volumeIDLock        sync.RWMutex
	metaPartitionIDLock sync.RWMutex

	metaNodeID      uint64
	volumeID        uint64
	metaPartitionID uint64

	store *boltdb.BoltStore
	r     *raft.Raft
}

func NewIDAllocator(store *boltdb.BoltStore, r *raft.Raft) *IDAllocator {
	alloc := &IDAllocator{
		store: store,
		r:     r,
	}

	alloc.metaNodeID = alloc.restore(maxMetaNodeIDKey)
	alloc.volumeID = alloc.restore(maxVolumeIDKey)
	alloc.metaPartitionID = alloc.restore(maxMetaPartitionIDKey)

	return alloc
}

func (alloc *IDAllocator) allocateMetaNodeID() (id uint64, err error) {
	alloc.metaNodeLock.Lock()
	defer alloc.metaNodeLock.Unlock()

	id = atomic.LoadUint64(&alloc.metaNodeID) + 1
	cmd, _ := json.Marshal(&RaftCmd{
		Op: opAllocMetaNodeID,
		K:  maxMetaNodeIDKey,
		V:  []byte(strconv.FormatUint(id, 10)),
	})

	if err := alloc.r.Apply(cmd, time.Second).Error(); err != nil {
		klog.Errorf(fmt.Sprintf("[allocateVolumeID] submit cmd %s err: %v", cmd, err))
		return 0, err
	}

	atomic.StoreUint64(&alloc.metaNodeID, id)
	if err := alloc.store.Set([]byte(maxMetaNodeIDKey), []byte(strconv.Itoa(int(id)))); err != nil {
		return 0, err
	}

	return id, nil
}

func (alloc *IDAllocator) allocateVolumeID() (id uint64, err error) {
	alloc.volumeIDLock.Lock()
	defer alloc.volumeIDLock.Unlock()

	id = atomic.LoadUint64(&alloc.volumeID) + 1
	cmd, _ := json.Marshal(&RaftCmd{
		Op: opAllocVolumeID,
		K:  maxVolumeIDKey,
		V:  []byte(strconv.FormatUint(id, 10)),
	})

	if err := alloc.r.Apply(cmd, time.Second).Error(); err != nil {
		klog.Errorf(fmt.Sprintf("[allocateVolumeID] submit cmd %s err: %v", cmd, err))
		return 0, err
	}

	atomic.StoreUint64(&alloc.volumeID, id)
	if err := alloc.store.Set([]byte(maxVolumeIDKey), []byte(strconv.Itoa(int(id)))); err != nil {
		return 0, err
	}

	return id, nil
}

func (alloc *IDAllocator) allocateMetaPartitionID() (partitionID uint64, err error) {
	alloc.metaPartitionIDLock.Lock()
	defer alloc.metaPartitionIDLock.Unlock()

	partitionID = atomic.LoadUint64(&alloc.metaPartitionID) + 1
	cmd, _ := json.Marshal(&RaftCmd{
		Op: opSyncAllocMetaPartitionID,
		K:  maxMetaPartitionIDKey,
		V:  []byte(strconv.FormatUint(partitionID, 10)),
	})
	if err := alloc.r.Apply(cmd, time.Second).Error(); err != nil {
		klog.Errorf(fmt.Sprintf("[allocateMetaPartitionID] submit cmd %s err: %v", cmd, err))
		return 0, err
	}

	atomic.StoreUint64(&alloc.metaPartitionID, partitionID)
	if err := alloc.store.Set([]byte(maxMetaPartitionIDKey), []byte(strconv.Itoa(int(partitionID)))); err != nil {
		return 0, err
	}

	return partitionID, nil
}

func (alloc *IDAllocator) restore(key string) uint64 {
	value, err := alloc.store.Get([]byte(maxMetaNodeIDKey))
	if len(value) == 0 || err != nil {
		return 0
	}

	id, _ := strconv.ParseUint(string(value), 10, 64)
	return id
}
