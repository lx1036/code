package master

import (
	"context"
	"fmt"
	"sync"
	"time"

	"k8s-lx1036/k8s/storage/gofs/pkg/raftstore"
	"k8s-lx1036/k8s/storage/gofs/pkg/util/proto"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

type MountClients struct {
	clientInfoMap map[uint64]*proto.ClientInfo
}

func newMountClients() (mountClients *MountClients) {
	mountClients = &MountClients{
		clientInfoMap: make(map[uint64]*proto.ClientInfo, 0),
	}
	return
}

//default value
const (
	defaultIntervalToCheckHeartbeat      = 60
	defaultIntervalToCheckMetaPartition  = 60
	defaultIntervalToCheckVolMountClient = 300
	noHeartBeatTimes                     = 3 // number of times that no heartbeat reported
	defaultNodeTimeOutSec                = noHeartBeatTimes * defaultIntervalToCheckHeartbeat
	defaultMetaPartitionTimeOutSec       = 10 * defaultIntervalToCheckHeartbeat
	//DefaultMetaPartitionMissSec                      = 3600
	defaultIntervalToAlarmMissingMetaPartition         = 10 * 60 // interval of checking if a replica is missing
	defaultMetaPartitionMemUsageThreshold      float32 = 0.75    // memory usage threshold on a meta partition
	defaultMaxMetaPartitionCountOnEachNode             = 10000
	defaultReplicaNum                                  = 3
	defaultRegion                                      = "us-east-2"
)

const (
	opSyncAddMetaNode          uint32 = 0x01
	opSyncAddVol               uint32 = 0x02
	opSyncAddMetaPartition     uint32 = 0x03
	opSyncUpdateMetaPartition  uint32 = 0x04
	opSyncDeleteMetaNode       uint32 = 0x05
	opSyncAllocMetaPartitionID uint32 = 0x06
	opSyncAllocCommonID        uint32 = 0x07
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
	keySeparator          = "#"
	metaNodeAcronym       = "mn"
	metaPartitionAcronym  = "mp"
	volAcronym            = "vol"
	clusterAcronym        = "c"
	nodeSetAcronym        = "s"
	bucketAcronym         = "bucket"
	clientAcronym         = "client"
	maxMetaPartitionIDKey = keySeparator + "max_mp_id"
	maxCommonIDKey        = keySeparator + "max_common_id"
	metaNodePrefix        = keySeparator + metaNodeAcronym + keySeparator
	volPrefix             = keySeparator + volAcronym + keySeparator
	metaPartitionPrefix   = keySeparator + metaPartitionAcronym + keySeparator
	clusterPrefix         = keySeparator + clusterAcronym + keySeparator
	nodeSetPrefix         = keySeparator + nodeSetAcronym + keySeparator
	bucketPrefix          = keySeparator + bucketAcronym + keySeparator
	clientPrefix          = keySeparator + clientAcronym + keySeparator
)

const (
	normal     uint8 = 0
	markDelete uint8 = 1
)

// AddrDatabase is a map that stores the address of a given host (e.g., the leader)
var AddrDatabase = make(map[uint64]string)

// Cluster stores all the cluster-level information.
type Cluster struct {
	Name string
	cfg  *clusterConfig

	buckets     map[string]*DeleteBucketInfo
	bucketMutex sync.RWMutex

	leaderInfo *LeaderInfo
	retainLogs uint64
	idAlloc    *IDAllocator

	topology *Topology

	metaNodes        map[string]*MetaNode
	metaNodeStatInfo *nodeStatInfo
	metaMutex        sync.RWMutex // meta node mutex

	vols                map[string]*Volume
	volMountClients     map[string]*MountClients
	volStatInfo         sync.Map
	volMutex            sync.RWMutex // volume mutex
	volMountClientMutex sync.RWMutex // volume mount client mutex
	createVolMutex      sync.RWMutex // create volume mutex

	DisableAutoAllocate bool
	fsm                 *raftstore.FilesystemStateMachine
	partition           raftstore.Partition
}

func (cluster *Cluster) scheduleToCheckHeartbeat() {
	go wait.UntilWithContext(context.TODO(), func(ctx context.Context) {
		if cluster.partition != nil && cluster.partition.IsRaftLeader() {
			leaderID, _ := cluster.partition.LeaderTerm()
			cluster.leaderInfo.addr = AddrDatabase[leaderID]
		}
	}, time.Second*defaultIntervalToCheckHeartbeat)

	go wait.UntilWithContext(context.TODO(), func(ctx context.Context) {
		if cluster.partition != nil && cluster.partition.IsRaftLeader() {
			cluster.checkMetaNodeHeartbeat()
		}
	}, time.Second*defaultIntervalToCheckHeartbeat)
}

func (cluster *Cluster) masterAddr() (addr string) {
	return cluster.leaderInfo.addr
}

func (cluster *Cluster) addMetaNode(nodeAddr string) (uint64, error) {
	cluster.metaMutex.Lock()
	defer cluster.metaMutex.Unlock()

	if metaNode, ok := cluster.metaNodes[nodeAddr]; ok {
		return metaNode.ID, nil
	}

	var metaNode *MetaNode
	var err error
	metaNode = newMetaNode(nodeAddr, cluster.Name)
	node := cluster.topology.getAvailNodeSetForMetaNode()
	if node == nil {
		// create node set
		id, err := cluster.idAlloc.allocateCommonID()
		if err != nil {
			klog.Error(err)
			return 0, err
		}
		node = newNodeSet(id, cluster.cfg.nodeSetCapacity)
		if err = cluster.putNodeSetInfo(opSyncAddNodeSet, node); err != nil {
			klog.Error(err)
			return 0, nil
		}

		cluster.topology.putNodeSet(node)
	}

	id, err := cluster.idAlloc.allocateCommonID()
	if err != nil {
		klog.Error(err)
		return 0, err
	}

	metaNode.ID = id
	metaNode.NodeSetID = node.ID
	if err = cluster.putMetaNodeInfo(opSyncAddMetaNode, metaNode); err != nil {
		klog.Error(err)
		return 0, err
	}

	node.increaseMetaNodeLen()
	if err = cluster.putNodeSetInfo(opSyncUpdateNodeSet, node); err != nil {
		node.decreaseMetaNodeLen()

		klog.Error(err)
		return 0, nil
	}

	// store metaNode
	cluster.metaNodes[nodeAddr] = metaNode

	return metaNode.ID, nil
}
func (cluster *Cluster) checkMetaNodeHeartbeat() {
	tasks := make([]*proto.AdminTask, 0)
	for _, node := range cluster.metaNodes {
		node.checkHeartbeat()
		task := node.createHeartbeatTask(cluster.masterAddr())
		tasks = append(tasks, task)
	}

	cluster.addMetaNodeTasks(tasks)
}

func (cluster *Cluster) scheduleToCheckMetaPartitions() {
	go wait.UntilWithContext(context.TODO(), func(ctx context.Context) {
		if cluster.partition != nil && cluster.partition.IsRaftLeader() {
			cluster.checkMetaPartitions()
		}
	}, time.Second*defaultIntervalToCheckMetaPartition)
}

// Return all the volumes except the ones that have been marked to be deleted.
func (cluster *Cluster) allVols() map[string]*Volume {
	vols := make(map[string]*Volume, 0)
	cluster.volMutex.RLock()
	defer cluster.volMutex.RUnlock()
	for name, vol := range cluster.vols {
		if vol.Status == normal {
			vols[name] = vol
		}
	}

	return vols
}

func (cluster *Cluster) checkMetaPartitions() {
	defer func() {
		if r := recover(); r != nil {
			klog.Warningf("checkMetaPartitions occurred panic,err[%v]", r)
		}
	}()

	volumes := cluster.allVols()
	for _, vol := range volumes {
		vol.checkMetaPartitions(cluster)
	}
}

func (cluster *Cluster) scheduleTask() {
	cluster.scheduleToCheckHeartbeat()
	cluster.scheduleToCheckMetaPartitions()
	//cluster.scheduleToUpdateStatInfo()
	//cluster.scheduleToCheckVolStatus()
	//cluster.scheduleToLoadMetaPartitions()
	//cluster.scheduleToCheckVolMountClients()
}

func (cluster *Cluster) getVolume(volName string) (*Volume, error) {
	cluster.volMutex.RLock()
	defer cluster.volMutex.RUnlock()
	vol, ok := cluster.vols[volName]
	if !ok {
		return nil, fmt.Errorf("vol %s not exists", volName)
	}

	return vol, nil
}

func NewCluster(name string, leaderInfo *LeaderInfo, fsm *MetadataFsm, partition raftstore.Partition, cfg *clusterConfig) *Cluster {
	return &Cluster{
		Name:            name,
		vols:            make(map[string]*Volume),
		volMountClients: make(map[string]*MountClients),
		buckets:         make(map[string]*DeleteBucketInfo),
		leaderInfo:      leaderInfo,
		cfg:             cfg,
		idAlloc:         NewIDAllocator(fsm.store, partition),
		topology:        newTopology(),
		fsm:             fsm,
		partition:       partition,
	}
}
