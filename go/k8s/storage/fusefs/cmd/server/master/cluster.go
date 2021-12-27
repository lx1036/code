package master

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"k8s-lx1036/k8s/storage/fusefs/cmd/server/meta/partition/raftstore"
	"k8s-lx1036/k8s/storage/fusefs/pkg/proto"

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
	metaNodePrefix        = keySeparator + metaNodeAcronym + keySeparator
	volPrefix             = "#vol#"
	metaPartitionPrefix   = keySeparator + metaPartitionAcronym + keySeparator
	clusterPrefix         = keySeparator + clusterAcronym + keySeparator
	nodeSetPrefix         = keySeparator + nodeSetAcronym + keySeparator
	bucketPrefix          = keySeparator + bucketAcronym + keySeparator
	clientPrefix          = keySeparator + clientAcronym + keySeparator
)

// AddrDatabase is a map that stores the address of a given host (e.g., the leader)
var AddrDatabase = make(map[uint64]string)

type LeaderInfo struct {
	addr string //host:port
}

// Cluster stores all the cluster-level information.
type Cluster struct {
	volMutex sync.RWMutex // volume mutex

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
	volMountClientMutex sync.RWMutex // volume mount client mutex

	DisableAutoAllocate bool
	fsm                 *MetadataFsm
	partition           raftstore.Partition
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

func (cluster *Cluster) start() {
	cluster.scheduleToCheckHeartbeat()
	//cluster.scheduleToCheckMetaPartitions()
	//cluster.scheduleToUpdateStatInfo()
	//cluster.scheduleToCheckVolStatus()
	//cluster.scheduleToLoadMetaPartitions()
	//cluster.scheduleToCheckVolMountClients()
}

func (cluster *Cluster) scheduleToCheckHeartbeat() {
	go wait.UntilWithContext(context.TODO(), func(ctx context.Context) {
		if cluster.partition != nil && cluster.partition.IsRaftLeader() {
			leaderID, term := cluster.partition.LeaderTerm()
			klog.Infof(fmt.Sprintf("[scheduleToCheckHeartbeat]leaderID:%d, term:%d", leaderID, term))
			cluster.leaderInfo.addr = AddrDatabase[leaderID]
		}
	}, time.Second*defaultIntervalToCheckHeartbeat)

	go wait.UntilWithContext(context.TODO(), func(ctx context.Context) {
		if cluster.partition != nil && cluster.partition.IsRaftLeader() {
			//cluster.checkMetaNodeHeartbeat()
		}
	}, time.Second*defaultIntervalToCheckHeartbeat)
}

// 1. submit `createVol` to raft
// 2. init 3 meta partition
func (cluster *Cluster) createVol(name, owner string, capacity uint64) (*Volume, error) {
	cluster.volMutex.Lock()
	defer cluster.volMutex.Unlock()

	if _, ok := cluster.vols[name]; ok {
		return nil, fmt.Errorf(fmt.Sprintf("volume %s is already existed", name))
	}

	// submit `createVol` to raft
	id, err := cluster.idAlloc.allocateVolumeID()
	if err != nil {
		klog.Errorf(fmt.Sprintf("allocate partition id for vol %s err: %v", name, err))
		return nil, err
	}
	vol := newVol(id, name, owner, capacity, defaultReplicaNum)
	if err = cluster.submitAddVol(vol); err != nil {
		klog.Errorf(fmt.Sprintf("submit add vol to raft err:%v", err))
		return nil, err
	}

	// submit `create meta partition` to raft
	err = vol.createMetaPartitions()
	if err != nil {
		klog.Errorf(fmt.Sprintf("create meta partitions for vol %s err: %v", name, err))
		return nil, err
	}

	cluster.vols[vol.Name] = vol

	// TODO: s3 create bucket
	return vol, nil
}

//key=#vol#volID,value=json.Marshal(vv)
func (cluster *Cluster) submitAddVol(vol *Volume) (err error) {
	return cluster.submitVol(opSyncAddVol, vol)
}

func (cluster *Cluster) submitUpdateVol(vol *Volume) (err error) {
	return cluster.submitVol(opSyncUpdateVol, vol)
}

func (cluster *Cluster) submitDeleteVol(vol *Volume) (err error) {
	return cluster.submitVol(opSyncDeleteVol, vol)
}

func (cluster *Cluster) submitVol(opType uint32, vol *Volume) (err error) {
	cmd := new(RaftCmd)
	cmd.Op = opType
	cmd.K = fmt.Sprintf("%s%d", volPrefix, vol.ID)
	cmd.V, _ = json.Marshal(newVolValue(vol))
	return cluster.submit(cmd)
}

func (cluster *Cluster) submit(cmd *RaftCmd) error {
	data, _ := json.Marshal(cmd)
	_, err := cluster.partition.Submit(data)
	return err
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
		id, err := cluster.idAlloc.allocateVolumeID()
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

	id, err := cluster.idAlloc.allocateVolumeID()
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

func (cluster *Cluster) scheduleToUpdateStatInfo() {
	go wait.UntilWithContext(context.TODO(), func(ctx context.Context) {
		if cluster.partition != nil && cluster.partition.IsRaftLeader() {
			cluster.updateStatInfo()
		}
	}, time.Second*defaultIntervalToCheckHeartbeat)
}

func (cluster *Cluster) scheduleToCheckVolStatus() {
	go wait.UntilWithContext(context.TODO(), func(ctx context.Context) {
		//check vols after switching leader two minutes
		if cluster.partition != nil && cluster.partition.IsRaftLeader() {
			cluster.checkDeleteBucket()
			//for _, vol := range cluster.vols {
			//	//vol.checkStatus(cluster)
			//}
		}
	}, time.Second*defaultIntervalToCheckMetaPartition)
}

func (cluster *Cluster) checkDeleteBucket() {
	for _, bucket := range cluster.buckets {
		deleted, err := cluster.deleteListObjects(bucket.AccessKey, bucket.SecretKey, bucket.Endpoint,
			bucket.Region, bucket.BucketName)
		if err != nil {
			klog.Errorf("action [checkDeleteBucket] deleteListObjects in bucket[%v] error: %v",
				bucket.BucketName, err)
			continue
		}

		if deleted {

		}
	}
}

func (cluster *Cluster) scheduleToCheckVolMountClients() {
	go wait.UntilWithContext(context.TODO(), func(ctx context.Context) {
		if cluster.partition != nil && cluster.partition.IsRaftLeader() {
			//cluster.checkVolMountClients()
		}
	}, time.Second*defaultIntervalToCheckVolMountClient)
}

// Return all the volumes except the ones that have been marked to be deleted.
func (cluster *Cluster) allVols() map[string]*Volume {
	cluster.volMutex.RLock()
	defer cluster.volMutex.RUnlock()

	vols := make(map[string]*Volume, 0)
	for name, vol := range cluster.vols {
		if vol.Status != MarkDeletedVol {
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

func (cluster *Cluster) getVolume(volName string) (*Volume, error) {
	cluster.volMutex.RLock()
	defer cluster.volMutex.RUnlock()
	vol, ok := cluster.vols[volName]
	if !ok {
		return nil, fmt.Errorf("vol %s not exists", volName)
	}

	return vol, nil
}
