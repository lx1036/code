package master

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"k8s-lx1036/k8s/storage/fusefs/cmd/server/meta/partition/raftstore"
	"k8s-lx1036/k8s/storage/fusefs/pkg/proto"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

//default value
const (
	defaultIntervalToCheckHeartbeat      = 10
	defaultIntervalToCheckMetaPartition  = 60
	defaultIntervalToCheckVolMountClient = 300
	noHeartBeatTimes                     = 3 // number of times that no heartbeat reported
	defaultNodeTimeOutSec                = noHeartBeatTimes * defaultIntervalToCheckHeartbeat
	defaultMetaPartitionTimeOutSec       = 10 * defaultIntervalToCheckHeartbeat
	//DefaultMetaPartitionMissSec                      = 3600
	defaultIntervalToAlarmMissingMetaPartition = 10 * 60 // interval of checking if a replica is missing
	defaultMaxMetaPartitionCountOnEachNode     = 10000
	defaultReplicaNum                          = 3
)

const (
	keySeparator         = "#"
	metaNodeAcronym      = "mn"
	metaPartitionAcronym = "mp"
	volAcronym           = "vol"
	clusterAcronym       = "c"
	nodeSetAcronym       = "s"
	bucketAcronym        = "bucket"
	clientAcronym        = "client"
	metaNodePrefix       = "#metanode#"
	volPrefix            = "#vol#"
	metaPartitionPrefix  = "#metapartition#"
	clusterPrefix        = keySeparator + clusterAcronym + keySeparator
	nodeSetPrefix        = "#s#"
	bucketPrefix         = keySeparator + bucketAcronym + keySeparator
	clientPrefix         = keySeparator + clientAcronym + keySeparator
)

type RaftCmd struct {
	Op uint32 `json:"op"`
	K  string `json:"k"`
	V  []byte `json:"v"`
}

// AddrDatabase is a map that stores the address of a given host (e.g., the leader)
var AddrDatabase = make(map[uint64]string)

type LeaderInfo struct {
	addr string //host:port
}

// Cluster stores all the cluster-level information.
type Cluster struct {
	volMutex  sync.RWMutex // volume mutex
	metaMutex sync.RWMutex // meta node mutex

	Name string

	buckets     map[string]*DeleteBucketInfo
	bucketMutex sync.RWMutex

	leaderInfo *LeaderInfo
	retainLogs uint64
	idAlloc    *IDAllocator

	topology *Topology

	metaNodes        map[string]*MetaNode
	metaNodeStatInfo *nodeStatInfo

	vols map[string]*Volume
	//volMountClients     map[string]*MountClients
	volStatInfo         sync.Map
	volMountClientMutex sync.RWMutex // volume mount client mutex

	DisableAutoAllocate bool
	fsm                 *MetadataFsm
	partition           raftstore.Partition
	peers               []raftstore.PeerAddress

	nodeSetCapacity int
}

func NewCluster(server *Server) *Cluster {
	return &Cluster{
		Name: server.clusterName,
		vols: make(map[string]*Volume),
		//volMountClients: make(map[string]*MountClients),
		buckets:         make(map[string]*DeleteBucketInfo),
		leaderInfo:      server.leaderInfo,
		idAlloc:         NewIDAllocator(server.fsm.store, server.partition),
		topology:        newTopology(),
		fsm:             server.fsm,
		partition:       server.partition,
		nodeSetCapacity: server.nodeSetCapacity,
		peers:           server.peers,
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
			if leader := cluster.getLeader(leaderID); leader != nil {
				cluster.leaderInfo.addr = fmt.Sprintf("%s:%d", leader.Address, leader.Port) // port
				klog.Infof(fmt.Sprintf("[scheduleToCheckHeartbeat]leaderID:%d, term:%d, leader is %s", leaderID, term, cluster.leaderInfo.addr))
			}
		}
	}, time.Second*defaultIntervalToCheckHeartbeat)

	go wait.UntilWithContext(context.TODO(), func(ctx context.Context) {
		if cluster.partition != nil && cluster.partition.IsRaftLeader() {
			//cluster.checkMetaNodeHeartbeat()
		}
	}, time.Second*defaultIntervalToCheckHeartbeat)
}

func (cluster *Cluster) getLeader(leaderID uint64) *raftstore.PeerAddress {
	for _, peer := range cluster.peers {
		if peer.ID == leaderID {
			return &peer
		}
	}

	return nil
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

	// (1)tcp every hosts for create meta partition
	// (2)submit `create 3 meta partitions` to raft
	err = vol.createMetaPartitions(cluster)
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

func (cluster *Cluster) addMetaNode(nodeAddr string) (uint64, error) {
	cluster.metaMutex.Lock()
	defer cluster.metaMutex.Unlock()

	if metaNode, ok := cluster.metaNodes[nodeAddr]; ok {
		return metaNode.ID, nil
	}

	var metaNode *MetaNode
	metaNode = newMetaNode(nodeAddr, cluster.Name)
	node := cluster.topology.getAvailNodeSetForMetaNode()
	if node == nil {
		// create node set
		id, err := cluster.idAlloc.allocateVolumeID()
		if err != nil {
			return 0, err
		}
		node = newNodeSet(id, cluster.nodeSetCapacity)
		if err = cluster.submitNodeSet(opSyncAddNodeSet, node); err != nil {
			return 0, err
		}

		cluster.topology.putNodeSet(node)
	}

	id, err := cluster.idAlloc.allocateVolumeID()
	if err != nil {
		return 0, err
	}
	metaNode.ID = id
	metaNode.NodeSetID = node.ID
	if err = cluster.submitMetaNode(opSyncAddMetaNode, metaNode); err != nil {
		return 0, err
	}

	node.increaseMetaNodeLen()
	if err = cluster.submitNodeSet(opSyncUpdateNodeSet, node); err != nil {
		node.decreaseMetaNodeLen()
		return 0, err
	}

	// store metaNode
	cluster.metaNodes[nodeAddr] = metaNode
	klog.Infof(fmt.Sprintf("add meta node %s succefully", nodeAddr))
	return metaNode.ID, nil
}

type nodeSetValue struct {
	ID          uint64 `json:"id"`
	Capacity    int    `json:"capacity"`
	MetaNodeLen int    `json:"metaNodeLen"`
}

func newNodeSetValue(nset *nodeSet) *nodeSetValue {
	return &nodeSetValue{
		ID:          nset.ID,
		Capacity:    nset.Capacity,
		MetaNodeLen: nset.metaNodeLen,
	}
}

// key=#s#
func (cluster *Cluster) submitNodeSet(opType uint32, nset *nodeSet) error {
	cmd := new(RaftCmd)
	cmd.Op = opType
	cmd.K = fmt.Sprintf("%s%s", nodeSetPrefix, strconv.FormatUint(nset.ID, 10))
	cmd.V, _ = json.Marshal(newNodeSetValue(nset))
	return cluster.submit(cmd)
}

type metaNodeValue struct {
	ID        uint64
	NodeSetID uint64
	Addr      string
}

func newMetaNodeValue(metaNode *MetaNode) *metaNodeValue {
	return &metaNodeValue{
		ID:        metaNode.ID,
		NodeSetID: metaNode.NodeSetID,
		Addr:      metaNode.Addr,
	}
}

// key=#metanode#
func (cluster *Cluster) submitMetaNode(opType uint32, metaNode *MetaNode) error {
	cmd := new(RaftCmd)
	cmd.Op = opType
	cmd.K = fmt.Sprintf("%s%s", metaNodePrefix, strconv.FormatUint(metaNode.ID, 10))
	cmd.V, _ = json.Marshal(newMetaNodeValue(metaNode))
	return cluster.submit(cmd)
}

// Choose the target hosts from the available node sets and meta nodes.
func (cluster *Cluster) chooseTargetMetaHosts(excludeNodeSet *nodeSet, excludeHosts []string, replicaNum int) (hosts []string,
	peers []proto.Peer, err error) {
	var (
		masterAddr []string
		slaveAddrs []string
		masterPeer []proto.Peer
		slavePeers []proto.Peer
		ns         *nodeSet
	)
	if ns, err = cluster.topology.allocNodeSetForMetaNode(excludeNodeSet, uint8(replicaNum)); err != nil {
		return nil, nil, err
	}
	if masterAddr, masterPeer, err = ns.getAvailMetaNodeHosts(excludeHosts, 1); err != nil {
		return nil, nil, err
	}
	peers = append(peers, masterPeer...)
	hosts = append(hosts, masterAddr[0])
	otherReplica := replicaNum - 1
	if otherReplica == 0 {
		return
	}
	excludeHosts = append(excludeHosts, hosts...)
	if slaveAddrs, slavePeers, err = ns.getAvailMetaNodeHosts(excludeHosts, otherReplica); err != nil {
		return nil, nil, err
	}
	hosts = append(hosts, slaveAddrs...)
	peers = append(peers, slavePeers...)
	if len(hosts) != replicaNum {
		return nil, nil, fmt.Errorf("no enough meta nodes for creating a meta partition")
	}

	return
}

// tcp call metaNode for create partition
func (cluster *Cluster) syncCreateMetaPartitionToMetaNode(host string, mp *MetaPartition) (err error) {
	req := &proto.CreateMetaPartitionRequest{
		Start:       mp.Start,
		End:         mp.End,
		PartitionID: mp.PartitionID,
		Members:     mp.Peers,
		VolName:     mp.volName,
	}
	task := proto.NewAdminTask(proto.OpCreateMetaPartition, host, req)
	task.ID = fmt.Sprintf("%v_pid[%v]", task.ID, mp.PartitionID)
	task.PartitionID = mp.PartitionID

	metaNode, err := cluster.metaNode(host)
	if err != nil {
		return err
	}

	_, err = metaNode.Sender.syncSendAdminTask(task)
	return err
}

func (cluster *Cluster) metaNode(addr string) (*MetaNode, error) {
	value, ok := cluster.metaNodes[addr]
	if !ok {
		return nil, fmt.Errorf("meta node %s is not found", addr)
	}

	return value, nil
}

type metaPartitionValue struct {
	PartitionID   uint64
	Start         uint64
	End           uint64
	VolID         uint64
	ReplicaNum    int
	Status        int8
	VolName       string
	Hosts         string
	Peers         []proto.Peer
	OfflinePeerID uint64
	IsMarkDeleted bool
}

func newMetaPartitionValue(mp *MetaPartition) *metaPartitionValue {
	return &metaPartitionValue{
		PartitionID: mp.PartitionID,
		Start:       mp.Start,
		End:         mp.End,
		VolID:       mp.volID,
		ReplicaNum:  mp.ReplicaNum,
		Status:      mp.Status,
		VolName:     mp.volName,
		Hosts:       strings.Join(mp.Hosts, "_"),
		Peers:       mp.Peers,
		//OfflinePeerID: mp.OfflinePeerID,
		//IsMarkDeleted: mp.IsMarkDeleted,
	}
}

// #metapartition#{volID}#{partitionID}
func (cluster *Cluster) submitMetaPartition(opType uint32, mp *MetaPartition) error {
	cmd := new(RaftCmd)
	cmd.Op = opType
	cmd.K = fmt.Sprintf("%s%s#%s", metaPartitionPrefix, strconv.FormatUint(mp.volID, 10), strconv.FormatUint(mp.PartitionID, 10))
	cmd.V, _ = json.Marshal(newMetaPartitionValue(mp))
	return cluster.submit(cmd)
}
