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
	"k8s-lx1036/k8s/storage/fusefs/pkg/util"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

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

func (m *RaftCmd) setOpType() {
	keyArr := strings.Split(m.K, keySeparator)
	if len(keyArr) < 2 {
		klog.Warningf("action[setOpType] invalid length[%v]", keyArr)
		return
	}
	switch keyArr[1] {
	case metaNodeAcronym:
		m.Op = opSyncAddMetaNode
	case metaPartitionAcronym:
		m.Op = opSyncAddMetaPartition
	case volAcronym:
		m.Op = opSyncAddVol
	case clusterAcronym:
		m.Op = opSyncPutCluster
	case nodeSetAcronym:
		m.Op = opSyncAddNodeSet
	case maxMetaPartitionIDKey:
		m.Op = opSyncAllocMetaPartitionID
	case maxVolumeIDKey:
		m.Op = opAllocVolumeID
	case bucketAcronym:
		m.Op = opSyncAddBucket
	default:
		klog.Warningf("action[setOpType] unknown opCode[%v]", keyArr[1])
	}
}

// AddrDatabase is a map that stores the address of a given host (e.g., the leader)
var AddrDatabase = make(map[uint64]string)

type LeaderInfo struct {
	addr string //host:port
}

type nodeStatInfo struct {
	TotalGB     uint64
	UsedGB      uint64
	IncreasedGB int64
	UsedRatio   string
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
	vols             map[string]*Volume
	//volStatInfo      sync.Map
	//volMountClients     map[string]*MountClients
	volMountClientMutex sync.RWMutex // volume mount client mutex

	DisableAutoAllocate bool
	fsm                 *MetadataFsm
	partition           raftstore.Partition
	peers               []raftstore.PeerAddress

	nodeSetCapacity int
}

func NewCluster(server *Server) *Cluster {
	return &Cluster{
		Name:      server.clusterName,
		vols:      make(map[string]*Volume),
		metaNodes: make(map[string]*MetaNode),
		//volMountClients: make(map[string]*MountClients),
		buckets:          make(map[string]*DeleteBucketInfo),
		metaNodeStatInfo: new(nodeStatInfo),
		leaderInfo:       server.leaderInfo,
		idAlloc:          NewIDAllocator(server.fsm.store, server.partition),
		topology:         newTopology(),
		fsm:              server.fsm,
		partition:        server.partition,
		nodeSetCapacity:  server.nodeSetCapacity,
		peers:            server.peers,
	}
}

func (cluster *Cluster) start() {
	cluster.scheduleToCheckHeartbeat()
	cluster.scheduleToCheckMetaPartitions()
	cluster.scheduleToUpdateStatInfo()
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
			cluster.checkMetaNodeHeartbeat()
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

func (cluster *Cluster) checkMetaNodeHeartbeat() {
	for _, node := range cluster.metaNodes {
		node.checkHeartbeat()
		task := node.createHeartbeatTask(cluster.leaderInfo.addr)
		node.Sender.AddTask(task)
	}
}

func (cluster *Cluster) scheduleToCheckMetaPartitions() {
	go wait.UntilWithContext(context.TODO(), func(ctx context.Context) {
		if cluster.partition != nil && cluster.partition.IsRaftLeader() {
			cluster.checkMetaPartitions()
		}
	}, time.Second*defaultIntervalToCheckMetaPartition)
}

func (cluster *Cluster) checkMetaPartitions() {
	/*for _, vol := range cluster.allVols() {
		//vol.checkMetaPartitions(cluster)
		//vol.checkSplitMetaPartition(cluster)
		maxPartitionID := vol.maxPartitionID()
		mps := vol.cloneMetaPartitionMap()
		for _, mp := range mps {
			//mp.checkStatus(cluster.Name, true, int(vol.metaPartitionCount), maxPartitionID)
			//mp.checkLeader()
			//mp.checkReplicaNum(cluster, vol.Name, vol.metaPartitionCount)
			//mp.checkEnd(cluster, maxPartitionID)
			//mp.reportMissingReplicas(cluster.Name, cluster.leaderInfo.addr, defaultMetaPartitionTimeOutSec)
			task := mp.replicaCreationTasks(cluster.Name, vol.Name)
			cluster.metaNodes[task.OperatorAddr].Sender.AddTask(task)
		}
	}*/
}

func (cluster *Cluster) scheduleToUpdateStatInfo() {
	go wait.UntilWithContext(context.TODO(), func(ctx context.Context) {
		if cluster.partition != nil && cluster.partition.IsRaftLeader() {
			var (
				total uint64
				used  uint64
			)
			for _, metaNode := range cluster.metaNodes {
				total += metaNode.Total
				used += metaNode.Used
			}
			useRate := float64(used) / float64(total)
			newUsed := used / util.GB
			cluster.metaNodeStatInfo.TotalGB = total / util.GB
			cluster.metaNodeStatInfo.IncreasedGB = int64(newUsed) - int64(cluster.metaNodeStatInfo.UsedGB)
			cluster.metaNodeStatInfo.UsedGB = newUsed
			cluster.metaNodeStatInfo.UsedRatio = strconv.FormatFloat(useRate, 'f', 3, 32) // 保留3位小数并四舍五入

			for name, volume := range cluster.allVols() {
				used, total = volume.totalUsedSpace(), volume.Capacity*util.GB
				if total <= 0 {
					continue
				}
				useRate = float64(used) / float64(total)
				cluster.vols[name].TotalGB = total / util.GB
				cluster.vols[name].UsedGB = used / util.GB
				cluster.vols[name].UsedRatio = strconv.FormatFloat(useRate, 'f', 3, 32)
			}
		}
	}, time.Second*defaultIntervalToCheckHeartbeat)
}

func (cluster *Cluster) scheduleToCheckVolStatus() {
	go wait.UntilWithContext(context.TODO(), func(ctx context.Context) {
		//check vols after switching leader two minutes
		if cluster.partition != nil && cluster.partition.IsRaftLeader() {

		}
	}, time.Second*defaultIntervalToCheckMetaPartition)
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

func (cluster *Cluster) getVolume(volName string) (*Volume, error) {
	cluster.volMutex.RLock()
	defer cluster.volMutex.RUnlock()
	vol, ok := cluster.vols[volName]
	if !ok {
		return nil, fmt.Errorf("vol %s not exists", volName)
	}

	return vol, nil
}

func (cluster *Cluster) scheduleToLoadMetaPartitions() {
	go wait.UntilWithContext(context.TODO(), func(ctx context.Context) {
		//check vols after switching leader two minutes
		if cluster.partition != nil && cluster.partition.IsRaftLeader() {
			if cluster.vols != nil {
				//cluster.checkLoadMetaPartitions()
			}
		}
	}, time.Second*defaultIntervalToCheckMetaPartition)
}

////////////////////////////HTTP///////////////////////////////////

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
	if err = cluster.submitVol(opSyncAddVol, vol); err != nil {
		klog.Errorf(fmt.Sprintf("submit add vol to raft err:%v", err))
		return nil, err
	}

	// (1)tcp every hosts for create meta partition
	// (2)submit `create 3 meta partitions` to raft
	err = vol.createMetaPartitions(cluster)
	if err != nil {
		cluster.submitVol(opSyncDeleteVol, vol)
		klog.Errorf(fmt.Sprintf("create meta partitions for vol %s err: %v", name, err))
		return nil, err
	}

	cluster.vols[vol.Name] = vol

	// TODO: s3 create bucket
	return vol, nil
}

func (cluster *Cluster) updateVol(name, owner string, capacity uint64) (*Volume, error) {
	if vol, err := cluster.getVolume(name); err != nil {
		return nil, err
	} else {
		vol.Capacity = capacity // expand capacity
		vol.Owner = owner
		if err = cluster.submitVol(opSyncUpdateVol, vol); err != nil {
			return nil, err
		}

		return vol, nil
	}
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

	metaNode, err := cluster.getMetaNode(host)
	if err != nil {
		return err
	}

	_, err = metaNode.Sender.syncSendAdminTask(task)
	return err
}

func (cluster *Cluster) getMetaNode(addr string) (*MetaNode, error) {
	value, ok := cluster.metaNodes[addr]
	if !ok {
		return nil, fmt.Errorf("meta node %s is not found", addr)
	}

	return value, nil
}

func (cluster *Cluster) getAllMetaPartitionIDByMetaNode(addr string) (partitionIDs []uint64) {
	partitionIDs = make([]uint64, 0)
	for _, volume := range cluster.vols {
		for _, partition := range volume.MetaPartitions {
			for _, host := range partition.Hosts {
				if host == addr {
					partitionIDs = append(partitionIDs, partition.PartitionID)
				}
			}
		}
	}

	return partitionIDs
}

////////////////////////////Submit cmd to Raft///////////////////////////////////

type VolStatus uint8

const (
	ReadWriteVol   VolStatus = 1
	MarkDeletedVol VolStatus = 2
	ReadOnlyVol    VolStatus = 3
)

const (
	opSyncAddMetaNode         uint32 = 0x01
	opSyncAddVol              uint32 = 0x02
	opSyncAddMetaPartition    uint32 = 0x03
	opSyncUpdateMetaPartition uint32 = 0x04
	opSyncDeleteMetaNode      uint32 = 0x05
	opSyncPutCluster          uint32 = 0x08
	opSyncUpdateVol           uint32 = 0x09
	opSyncDeleteVol           uint32 = 0x0A
	opSyncDeleteMetaPartition uint32 = 0x0B
	opSyncAddNodeSet          uint32 = 0x0C
	opSyncUpdateNodeSet       uint32 = 0x0D
)

//key=#vol#volID,value=json.Marshal(vv)
func (cluster *Cluster) submitVol(opType uint32, vol *Volume) (err error) {
	cmd := new(RaftCmd)
	cmd.Op = opType
	cmd.K = fmt.Sprintf("%s%d", volPrefix, vol.ID)
	data := struct {
		ID     uint64    `json:"id"`
		Name   string    `json:"name"`
		Owner  string    `json:"owner"`
		Status VolStatus `json:"status"`
	}{
		ID:     vol.ID,
		Name:   vol.Name,
		Owner:  vol.Owner,
		Status: vol.Status,
	}
	cmd.V, _ = json.Marshal(&data)
	return cluster.submit(cmd)
}

// key=#s#
func (cluster *Cluster) submitNodeSet(opType uint32, nset *nodeSet) error {
	cmd := new(RaftCmd)
	cmd.Op = opType
	cmd.K = fmt.Sprintf("%s%s", nodeSetPrefix, strconv.FormatUint(nset.ID, 10))
	data := struct {
		ID          uint64 `json:"id"`
		Capacity    int    `json:"capacity"`
		MetaNodeLen int    `json:"metaNodeLen"`
	}{
		ID:          nset.ID,
		Capacity:    nset.Capacity,
		MetaNodeLen: nset.metaNodeLen,
	}
	cmd.V, _ = json.Marshal(&data)
	return cluster.submit(cmd)
}

// key=#metanode#
func (cluster *Cluster) submitMetaNode(opType uint32, metaNode *MetaNode) error {
	cmd := new(RaftCmd)
	cmd.Op = opType
	cmd.K = fmt.Sprintf("%s%s", metaNodePrefix, strconv.FormatUint(metaNode.ID, 10))
	data := struct {
		ID        uint64 `json:"id"`
		NodeSetID uint64 `json:"nodeSetID"`
		Addr      string `json:"addr"`
	}{
		ID:        metaNode.ID,
		NodeSetID: metaNode.NodeSetID,
		Addr:      metaNode.Addr,
	}
	cmd.V, _ = json.Marshal(&data)
	return cluster.submit(cmd)
}

// #metapartition#{volID}#{partitionID}
func (cluster *Cluster) submitMetaPartition(opType uint32, mp *MetaPartition) error {
	cmd := new(RaftCmd)
	cmd.Op = opType
	cmd.K = fmt.Sprintf("%s%s#%s", metaPartitionPrefix, strconv.FormatUint(mp.volID, 10),
		strconv.FormatUint(mp.PartitionID, 10))
	data := &struct {
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
	}{
		PartitionID: mp.PartitionID,
		Start:       mp.Start,
		End:         mp.End,
		VolID:       mp.volID,
		ReplicaNum:  mp.ReplicaNum,
		Status:      mp.Status,
		VolName:     mp.volName,
		Hosts:       strings.Join(mp.Hosts, "_"),
		Peers:       mp.Peers,
	}
	cmd.V, _ = json.Marshal(&data)
	return cluster.submit(cmd)
}

func (cluster *Cluster) submit(cmd *RaftCmd) error {
	data, _ := json.Marshal(cmd)
	_, err := cluster.partition.Submit(data)
	return err
}
