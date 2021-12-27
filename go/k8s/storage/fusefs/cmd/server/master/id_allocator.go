package master

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"

	"k8s-lx1036/k8s/storage/fusefs/cmd/server/meta/partition/raftstore"

	"k8s.io/klog/v2"
)

const (
	opSyncAddMetaNode          uint32 = 0x01
	opSyncAddVol               uint32 = 0x02
	opSyncAddMetaPartition     uint32 = 0x03
	opSyncUpdateMetaPartition  uint32 = 0x04
	opSyncDeleteMetaNode       uint32 = 0x05
	opSyncAllocMetaPartitionID uint32 = 0x06
	opAllocVolumeID            uint32 = 0x07
	opSyncPutCluster           uint32 = 0x08
	opSyncUpdateVol            uint32 = 0x09
	opSyncDeleteVol            uint32 = 0x0A
	opSyncDeleteMetaPartition  uint32 = 0x0B
	opSyncAddNodeSet           uint32 = 0x0C
	opSyncUpdateNodeSet        uint32 = 0x0D
	opSyncBatchPut             uint32 = 0x0E
	opSyncAddBucket            uint32 = 0x0F
	opSyncUpdateBucket         uint32 = 0x10
	opSyncDeleteBucket         uint32 = 0x11
	opSyncAddVolMountClient    uint32 = 0x12
	opSyncUpdateVolMountClient uint32 = 0x13
	opSyncDeleteVolMountClient uint32 = 0x14
)

const (
	maxVolumeIDKey = "#max_volume_id"
)

// IDAllocator generates and allocates ids
type IDAllocator struct {
	volumeIDLock        sync.RWMutex
	metaPartitionIDLock sync.RWMutex

	volumeID uint64

	metaPartitionID uint64
	store           *raftstore.BoltDBStore
	partition       raftstore.Partition
}

func NewIDAllocator(store *raftstore.BoltDBStore, partition raftstore.Partition) *IDAllocator {
	return &IDAllocator{
		store:     store,
		partition: partition,
	}
}

func (alloc *IDAllocator) allocateVolumeID() (id uint64, err error) {
	alloc.volumeIDLock.Lock()
	defer alloc.volumeIDLock.Unlock()

	id = atomic.LoadUint64(&alloc.volumeID) + 1
	cmd, _ := json.Marshal(&RaftCmd{
		Op: opAllocVolumeID,
		K:  maxVolumeIDKey,
		V:  []byte(strconv.Itoa(int(atomic.LoadUint64(&alloc.volumeID) + 1))),
	})
	if _, err = alloc.partition.Submit(cmd); err != nil {
		klog.Errorf(fmt.Sprintf("action[allocateVolumeID] submit cmd %s err: %v", cmd, err))
		return 0, err
	}

	atomic.StoreUint64(&alloc.volumeID, id)

	return id, nil
}

func (alloc *IDAllocator) restore() {
	alloc.restoreMaxMetaPartitionID()
	alloc.restoreMaxCommonID()
}

func (alloc *IDAllocator) restoreMaxMetaPartitionID() {
	value, err := alloc.store.Get([]byte(maxMetaPartitionIDKey))
	if err != nil {
		panic(fmt.Sprintf("Failed to restore maxPartitionID,err:%v ", err.Error()))
	}
	if len(value) == 0 {
		alloc.metaPartitionID = 0
		return
	}
	maxPartitionID, err := strconv.ParseUint(string(value), 10, 64)
	if err != nil {
		panic(fmt.Sprintf("Failed to restore maxPartitionID,err:%v ", err.Error()))
	}
	alloc.metaPartitionID = maxPartitionID
	klog.Infof("action[restoreMaxMetaPartitionID] maxMpID[%v]", alloc.metaPartitionID)
}

// The data node, meta node, and node set share the same ID allocator.
func (alloc *IDAllocator) restoreMaxCommonID() {
	value, err := alloc.store.Get([]byte(maxVolumeIDKey))
	if err != nil {
		panic(fmt.Sprintf("Failed to restore maxCommonID,err:%v ", err.Error()))
	}
	if len(value) == 0 {
		alloc.volumeID = 0
		return
	}
	maxMetaNodeID, err := strconv.ParseUint(string(value), 10, 64)
	if err != nil {
		panic(fmt.Sprintf("Failed to restore maxCommonID,err:%v ", err.Error()))
	}
	alloc.volumeID = maxMetaNodeID
	klog.Infof("action[restoreMaxCommonID] maxMnID[%v]", alloc.volumeID)
}
