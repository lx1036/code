package master

import (
	"fmt"
	"sync"

	"k8s-lx1036/k8s/storage/fusefs/pkg/proto"
	"k8s.io/klog/v2"
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

func (vol *Volume) addMetaPartition(mp *MetaPartition) {
	vol.MetaPartitions[mp.PartitionID] = mp
}

const (
	defaultMaxMetaPartitionInodeID  uint64 = 1<<63 - 1
	defaultMetaPartitionInodeIDStep uint64 = 1 << 24 // 16MB
)

// 1. tcp every hosts for create meta partition
// 2. submit meta partition cmd to raft
func (vol *Volume) createMetaPartitions(cluster *Cluster) (err error) {
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

		if mp, err := vol.tcpCreateMetaPartition(cluster, start, end); err != nil {
			klog.Errorf("[createMetaPartitions]vol[%v] create meta partition err[%v]", vol.Name, err)
			break
		} else {
			if err = cluster.submitMetaPartition(opSyncAddMetaPartition, mp); err != nil {
				klog.Errorf("[createMetaPartitions]vol[%v] submit meta partition to raft err[%v]", vol.Name, err)
				break
			}

			vol.addMetaPartition(mp)
		}
	}

	if len(vol.MetaPartitions) != vol.metaPartitionCount {
		err = fmt.Errorf("action[initMetaPartitions] vol[%v] init meta partition failed,mpCount[%v],expectCount[%v]",
			vol.Name, len(vol.MetaPartitions), vol.metaPartitionCount)
	}
	return
}

func (vol *Volume) tcpCreateMetaPartition(cluster *Cluster, start, end uint64) (mp *MetaPartition, err error) {
	var (
		hosts       []string
		partitionID uint64
		peers       []proto.Peer
		wg          sync.WaitGroup
	)
	if hosts, peers, err = cluster.chooseTargetMetaHosts(nil, nil, vol.metaPartitionCount); err != nil {
		return nil, err
	}
	if partitionID, err = cluster.idAlloc.allocateMetaPartitionID(); err != nil {
		return nil, err
	}
	mp = newMetaPartition(partitionID, start, end, vol.metaPartitionCount, vol.Name, vol.ID, false)
	mp.setHosts(hosts)
	mp.setPeers(peers)

	for _, host := range hosts {
		wg.Add(1)
		go func(host string) {
			defer wg.Done()
			if err = cluster.syncCreateMetaPartitionToMetaNode(host, mp); err != nil {
				klog.Error(err)
			}
		}(host)
	}
	wg.Wait()

	mp.Status = proto.ReadWrite
	return mp, nil
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
