package master

import (
	"fmt"
	"k8s-lx1036/k8s/storage/fusefs/pkg/proto"
	"k8s.io/klog/v2"
	"sync"
)

type VolStatus uint8

const (
	ReadWriteVol   VolStatus = 1
	MarkDeletedVol VolStatus = 2
	ReadOnlyVol    VolStatus = 3
)

type volValue struct {
	ID     uint64    `json:"id"`
	Name   string    `json:"name"`
	Owner  string    `json:"owner"`
	Status VolStatus `json:"status"`
}

func newVolValue(vol *Volume) (vv *volValue) {
	return &volValue{
		ID:     vol.ID,
		Name:   vol.Name,
		Owner:  vol.Owner,
		Status: vol.Status,
	}
}

// Volume represents a set of meta partitionMap and data partitionMap
type Volume struct {
	ID                 uint64
	Name               string
	Owner              string
	s3Endpoint         string
	metaPartitionCount int
	Status             VolStatus
	threshold          float32
	Capacity           uint64 // GB
	MetaPartitions     map[uint64]*MetaPartition
	mpsLock            sync.RWMutex
	mpsCache           []byte
	viewCache          []byte
	bucketdeleted      bool
	createMpMutex      sync.RWMutex
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
		bucketdeleted:      false,
	}
}

const (
	defaultMaxMetaPartitionInodeID  uint64 = 1<<63 - 1
	defaultMetaPartitionInodeIDStep uint64 = 1 << 24
)

func (vol *Volume) createMetaPartitions() (err error) {
	// initialize k meta partitionMap at a time
	var (
		start uint64
		end   uint64
	)
	for index := 0; index < vol.metaPartitionCount; index++ {
		if index != 0 {
			start = end + 1
		}

		end = start + defaultMetaPartitionInodeIDStep
		if index == vol.metaPartitionCount-1 {
			end = defaultMaxMetaPartitionInodeID
		}

		if err := vol.createMetaPartition(start, end); err != nil {
			klog.Errorf("action[initMetaPartitions] vol[%v] init meta partition err[%v]", vol.Name, err)
			break
		}
	}

	if len(vol.MetaPartitions) != vol.metaPartitionCount {
		err = fmt.Errorf("action[initMetaPartitions] vol[%v] init meta partition failed,mpCount[%v],expectCount[%v]",
			vol.Name, len(vol.MetaPartitions), vol.metaPartitionCount)
	}
	return
}

func (vol *Volume) createMetaPartition(start, end uint64) error {
	return nil
}

func (vol *Volume) doCreateMetaPartition(start, end uint64) {

}

func (vol *Volume) checkMetaPartitions(c *Cluster) {
	var tasks []*proto.AdminTask
	//vol.checkSplitMetaPartition(c)
	//maxPartitionID := vol.maxPartitionID()
	//mps := vol.cloneMetaPartitionMap()
	//for _, mp := range mps {
	//	mp.checkStatus(c.Name, true, int(vol.mpReplicaNum), maxPartitionID)
	//	mp.checkLeader()
	//	mp.checkReplicaNum(c, vol.Name, vol.mpReplicaNum)
	//	mp.checkEnd(c, maxPartitionID)
	//	mp.reportMissingReplicas(c.Name, c.leaderInfo.addr, defaultMetaPartitionTimeOutSec, defaultIntervalToAlarmMissingMetaPartition)
	//	tasks = append(tasks, mp.replicaCreationTasks(c.Name, vol.Name)...)
	//}
	c.addMetaNodeTasks(tasks)
}
