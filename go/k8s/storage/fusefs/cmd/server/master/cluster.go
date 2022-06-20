package master

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"k8s-lx1036/k8s/storage/fusefs/pkg/proto"
	"k8s-lx1036/k8s/storage/fusefs/pkg/util"

	"github.com/hashicorp/raft"
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
	nodeSetPrefix        = "#nodeset#"
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

	r      *raft.Raft
	server *Server

	buckets     map[string]*DeleteBucketInfo
	bucketMutex sync.RWMutex

	retainLogs uint64
	idAlloc    *IDAllocator
	fsm        *Fsm

	nodeSets *NodeSets

	metaNodes        map[string]*MetaNode
	metaNodeStatInfo *nodeStatInfo
	vols             map[string]*Volume
	//volStatInfo      sync.Map

	DisableAutoAllocate bool

	nodeSetCapacity int
}

func NewCluster(server *Server) *Cluster {
	return &Cluster{
		Name:   server.clusterName,
		r:      server.r,
		server: server,

		vols:             make(map[string]*Volume),
		metaNodes:        make(map[string]*MetaNode),
		buckets:          make(map[string]*DeleteBucketInfo),
		metaNodeStatInfo: new(nodeStatInfo),
		idAlloc:          NewIDAllocator(server.fsm.store, server.r),
		nodeSets:         newNodeSets(),
		fsm:              server.fsm,
		nodeSetCapacity:  server.nodeSetCapacity,
	}
}

func (cluster *Cluster) start() {
	cluster.checkMetaNodeHeartbeat()
	//cluster.scheduleToCheckMetaPartitions()
	//cluster.scheduleToUpdateStatInfo()
	//cluster.scheduleToCheckVolStatus()
	//cluster.scheduleToLoadMetaPartitions()
	//cluster.scheduleToCheckVolMountClients()
}

func (cluster *Cluster) checkMetaNodeHeartbeat() {
	go wait.UntilWithContext(context.TODO(), func(ctx context.Context) {
		if cluster.server.isRaftLeader() {
			for _, node := range cluster.metaNodes {
				node.checkHeartbeat()
				task := node.createHeartbeatTask(string(cluster.r.Leader()))
				node.Sender.AddTask(task)
			}
		}
	}, time.Second*defaultIntervalToCheckHeartbeat)
}

func (cluster *Cluster) scheduleToCheckMetaPartitions() {
	go wait.UntilWithContext(context.TODO(), func(ctx context.Context) {
		if cluster.server.isRaftLeader() {
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
		if cluster.server.isRaftLeader() {
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
		if cluster.server.isRaftLeader() {

		}
	}, time.Second*defaultIntervalToCheckMetaPartition)
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
		if cluster.server.isRaftLeader() {
			if cluster.vols != nil {
				//cluster.checkLoadMetaPartitions()
			}
		}
	}, time.Second*defaultIntervalToCheckMetaPartition)
}

/// metaNode api ///

func (cluster *Cluster) addMetaNode(nodeAddr string) (uint64, error) {
	cluster.metaMutex.Lock()
	defer cluster.metaMutex.Unlock()

	if metaNode, ok := cluster.metaNodes[nodeAddr]; ok {
		return metaNode.ID, nil
	}

	var metaNode *MetaNode
	metaNode = newMetaNode(nodeAddr, cluster.Name)
	node := cluster.nodeSets.getAvailNodeSetForMetaNode()
	if node == nil {
		// create node set
		id, err := cluster.idAlloc.allocateMetaNodeID()
		if err != nil {
			return 0, err
		}
		node = newNodeSet(id, cluster.nodeSetCapacity)
		if err = cluster.submitNodeSet(opSyncAddNodeSet, node); err != nil {
			return 0, err
		}

		cluster.nodeSets.putNodeSet(node)
	}

	id, err := cluster.idAlloc.allocateMetaNodeID()
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

/// volume api ///

// 1. submit `createVol` to raft
// 2. init 3 meta partition
func (cluster *Cluster) createVol(name, owner, accessKey, secretKey, endpoint string, capacity uint64, createBackend bool) (*Volume, error) {
	cluster.volMutex.Lock()
	defer cluster.volMutex.Unlock()

	var err error

	if _, ok := cluster.vols[name]; ok {
		return nil, fmt.Errorf(fmt.Sprintf("volume %s is already existed", name))
	}

	if createBackend {
		if err = cluster.CreateBucket(accessKey, secretKey, endpoint, name); err != nil {
			return nil, fmt.Errorf(fmt.Sprintf("[createVol]create s3 bucket err:%v", err))
		}

		defer func() {
			if err != nil {
				_ = cluster.DeleteBucket(accessKey, secretKey, endpoint, name)
			}
		}()
	}

	// submit `createVol` to raft
	id, err := cluster.idAlloc.allocateVolumeID()
	if err != nil {
		klog.Errorf(fmt.Sprintf("allocate partition id for vol %s err: %v", name, err))
		return nil, err
	}

	// inode 三个范围: 0 ~ 1<<24, 1<<24+1 ~ 1<<25, 1<<25+1 ~ 1<<63-1，并创建三个对应的 metaPartition
	var (
		start uint64
		end   uint64
	)
	vol := newVol(id, name, owner, capacity, defaultReplicaNum)
	for index := 0; index < vol.metaPartitionCount; index++ {
		if index != 0 {
			start = end + 1
		}

		end = start + defaultMetaPartitionInodeIDStep
		if index == vol.metaPartitionCount-1 {
			end = defaultMaxMetaPartitionInodeID
		}
		// (1)tcp every host for create meta partition
		mp, err := cluster.tcpCreateMetaPartition(vol, start, end)
		if err != nil {
			klog.Errorf("[createMetaPartitions]vol[%v] create meta partition err[%v]", vol.Name, err)
			break
		}

		// (2)submit `create 3 meta partitions` to raft
		if err = cluster.submitMetaPartition(opSyncAddMetaPartition, mp); err != nil {
			klog.Errorf("[createMetaPartitions]vol[%v] submit meta partition to raft err[%v]", vol.Name, err)
			break
		}

		vol.addMetaPartition(mp)
	}
	if len(vol.MetaPartitions) != vol.metaPartitionCount {
		return nil, fmt.Errorf("[createVol]vol %s init meta partition failed,mpCount[%v],expectCount[%v]",
			vol.Name, len(vol.MetaPartitions), vol.metaPartitionCount)
	}

	if err = cluster.submitVol(opSyncAddVol, vol); err != nil {
		cluster.submitVol(opSyncDeleteVol, vol)
		klog.Errorf(fmt.Sprintf("submit add vol to raft err:%v", err))
		return nil, err
	}

	cluster.vols[vol.Name] = vol

	// TODO: s3 create bucket
	return vol, nil
}

// INFO: 从5个 meta node 里根据使用率最小选择3个，并在每个meta node里异步创建对应的 MetaPartition
func (cluster *Cluster) tcpCreateMetaPartition(vol *Volume, start, end uint64) (mp *MetaPartition, err error) {
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

	mp = &MetaPartition{
		PartitionID: partitionID,
		Start:       start,
		End:         end,
		Replicas:    make([]*MetaReplica, 0),
		ReplicaNum:  vol.metaPartitionCount,
		Status:      Unavailable,
		volID:       vol.ID,
		volName:     vol.Name,
		Hosts:       hosts,
		Peers:       peers,
		MissNodes:   make(map[string]int64, 0),
	}

	// tcp call metaNode for create partition
	for _, host := range hosts {
		wg.Add(1)
		go func(host string) {
			defer wg.Done()

			req := &proto.CreateMetaPartitionRequest{
				Start:       start,
				End:         end,
				PartitionID: partitionID,
				Members:     peers,
				VolName:     vol.Name,
			}
			task := proto.NewAdminTask(proto.OpCreateMetaPartition, host, req)
			task.ID = fmt.Sprintf("%v_pid[%v]", task.ID, partitionID)
			task.PartitionID = partitionID
			metaNode, err := cluster.getMetaNode(host)
			if err != nil {
				klog.Error(err)
			}
			_, err = metaNode.Sender.syncSendAdminTask(task)
			if err != nil {
				klog.Error(err)
			}
		}(host)
	}
	wg.Wait()

	mp.Status = ReadWrite
	return mp, nil
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

// Choose the target hosts from the available node sets and meta nodes.
func (cluster *Cluster) chooseTargetMetaHosts(excludeNodeSet *nodeSet, excludeHosts []string, replicaNum int) (hosts []string,
	peers []proto.Peer, err error) {
	var (
		masterAddr []string
		slaveAddrs []string
		masterPeer []proto.Peer
		slavePeers []proto.Peer
		nodeSets   *nodeSet
	)
	if nodeSets, err = cluster.nodeSets.allocNodeSetForMetaNode(excludeNodeSet, uint8(replicaNum)); err != nil {
		return nil, nil, err
	}
	if masterAddr, masterPeer, err = nodeSets.getAvailMetaNodeHosts(excludeHosts, 1); err != nil {
		return nil, nil, err
	}
	peers = append(peers, masterPeer...)
	hosts = append(hosts, masterAddr[0])
	otherReplica := replicaNum - 1
	if otherReplica == 0 {
		return
	}
	excludeHosts = append(excludeHosts, hosts...)
	if slaveAddrs, slavePeers, err = nodeSets.getAvailMetaNodeHosts(excludeHosts, otherReplica); err != nil {
		return nil, nil, err
	}
	hosts = append(hosts, slaveAddrs...)
	peers = append(peers, slavePeers...)
	if len(hosts) != replicaNum {
		return nil, nil, fmt.Errorf("no enough meta nodes for creating a meta partition")
	}

	return
}

////////////////////////////Submit cmd to Raft///////////////////////////////////
// apply log to fsm
func (cluster *Cluster) submit(cmd *RaftCmd) error {
	data, _ := json.Marshal(cmd)
	return cluster.r.Apply(data, time.Second).Error()
}
