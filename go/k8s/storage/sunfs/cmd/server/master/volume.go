package master

import (
	"k8s-lx1036/k8s/storage/sunfs/pkg/proto"
	"sync"
)

// Vol represents a set of meta partitionMap and data partitionMap
type Volume struct {
	ID             uint64
	Name           string
	Owner          string
	s3Endpoint     string
	mpReplicaNum   uint8
	Status         uint8
	threshold      float32
	Capacity       uint64 // GB
	MetaPartitions map[uint64]*MetaPartition
	mpsLock        sync.RWMutex
	mpsCache       []byte
	viewCache      []byte
	bucketdeleted  bool
	createMpMutex  sync.RWMutex
	sync.RWMutex
}

func (vol *Volume) checkMetaPartitions(c *Cluster) {
	var tasks []*proto.AdminTask
	vol.checkSplitMetaPartition(c)
	maxPartitionID := vol.maxPartitionID()
	mps := vol.cloneMetaPartitionMap()
	for _, mp := range mps {
		mp.checkStatus(c.Name, true, int(vol.mpReplicaNum), maxPartitionID)
		mp.checkLeader()
		mp.checkReplicaNum(c, vol.Name, vol.mpReplicaNum)
		mp.checkEnd(c, maxPartitionID)
		mp.reportMissingReplicas(c.Name, c.leaderInfo.addr, defaultMetaPartitionTimeOutSec, defaultIntervalToAlarmMissingMetaPartition)
		tasks = append(tasks, mp.replicaCreationTasks(c.Name, vol.Name)...)
	}
	c.addMetaNodeTasks(tasks)
}

func newVol(id uint64, name, owner string, capacity uint64, mpReplicaNum uint8) *Volume {
	volume := &Volume{
		ID:             id,
		Name:           name,
		Owner:          owner,
		threshold:      defaultMetaPartitionMemUsageThreshold,
		Capacity:       capacity,
		MetaPartitions: make(map[uint64]*MetaPartition, 0),
		bucketdeleted:  false,
	}

	if mpReplicaNum < defaultReplicaNum {
		mpReplicaNum = defaultReplicaNum
	}
	volume.mpReplicaNum = mpReplicaNum

	return volume
}
