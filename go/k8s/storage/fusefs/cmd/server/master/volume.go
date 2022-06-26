package master

import (
	"encoding/json"
	"fmt"
	"sync"
)

// Volume represents a set of meta partitionMap and data partitionMap
type Volume struct {
	mpsLock       sync.RWMutex
	createMpMutex sync.RWMutex

	ID                 uint64                    `json:"id"`
	Name               string                    `json:"name"`
	Owner              string                    `json:"owner"`
	s3Endpoint         string                    `json:"s3Endpoint"`
	metaPartitionCount int                       `json:"metaPartitionCount"`
	Status             VolStatus                 `json:"status"`
	threshold          float32                   `json:"threshold"`
	Capacity           uint64                    `json:"capacity"` // GB
	MetaPartitions     map[uint64]*MetaPartition `json:"metaPartitions"`
	mpsCache           []byte                    `json:"mpsCache"`
	viewCache          []byte                    `json:"viewCache"`
	bucketDeleted      bool                      `json:"bucketDeleted"`

	// stats
	TotalGB   uint64 `json:"totalGB"`
	UsedGB    uint64 `json:"usedGB"`
	UsedRatio string `json:"usedRatio"`
}

func newVol(id uint64, name, owner string, capacity uint64, metaPartitionCount int) *Volume {
	return &Volume{
		ID:                 id,
		Name:               name,
		metaPartitionCount: metaPartitionCount,
		Owner:              owner,
		threshold:          defaultMetaPartitionMemUsageThreshold,
		Capacity:           capacity,
		MetaPartitions:     make(map[uint64]*MetaPartition, 0),
		bucketDeleted:      false,
	}
}

func (vol *Volume) addMetaPartition(mp *MetaPartition) {
	vol.MetaPartitions[mp.PartitionID] = mp
}

func (vol *Volume) maxPartitionID() (maxPartitionID uint64) {
	for id := range vol.MetaPartitions {
		if id > maxPartitionID {
			maxPartitionID = id
		}
	}
	return
}

func (vol *Volume) totalUsedSpace() (totalSize uint64) {
	for _, mp := range vol.MetaPartitions {
		totalSize += mp.Size
	}
	return
}

const (
	defaultMaxMetaPartitionInodeID  uint64 = 1<<63 - 1
	defaultMetaPartitionInodeIDStep uint64 = 1 << 24 // 16MB
)

func (vol *Volume) cloneMetaPartitionMap() map[uint64]*MetaPartition {
	mps := make(map[uint64]*MetaPartition, 0)
	for partitionID, mp := range vol.MetaPartitions {
		mps[partitionID] = mp
	}

	return mps
}

//key=#vol#volID,value=json.Marshal(vv)
func (cluster *Cluster) submitVol(opType uint32, vol *Volume) (err error) {
	cmd := new(RaftCmd)
	cmd.Op = opType
	cmd.K = fmt.Sprintf("%s%d", volPrefix, vol.ID)
	cmd.V, _ = json.Marshal(vol)
	return cluster.submit(cmd)
}
