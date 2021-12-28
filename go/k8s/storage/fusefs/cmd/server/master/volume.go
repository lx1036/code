package master

import (
	"fmt"
	"sync"

	"k8s-lx1036/k8s/storage/fusefs/pkg/proto"
	"k8s.io/klog/v2"
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

// 1. tcp every host for create meta partition
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

func (vol *Volume) cloneMetaPartitionMap() map[uint64]*MetaPartition {
	mps := make(map[uint64]*MetaPartition, 0)
	for partitionID, mp := range vol.MetaPartitions {
		mps[partitionID] = mp
	}

	return mps
}
