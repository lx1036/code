// INFO: master 是一个 raft state machine

package master

import (
	"fmt"
	"net/http/httputil"
	"strconv"
	"sync"
	"time"

	"k8s-lx1036/k8s/storage/sunfs/pkg/config"
	"k8s-lx1036/k8s/storage/sunfs/pkg/raftstore"
	"k8s-lx1036/k8s/storage/sunfs/pkg/util"

	"k8s.io/klog/v2"
)

const (
	LRUCacheSize    = 3 << 30
	WriteBufferSize = 4 * util.MB
)

const (
	defaultInitMetaPartitionCount           = 3
	defaultMaxInitMetaPartitionCount        = 100
	defaultMaxMetaPartitionInodeID   uint64 = 1<<63 - 1
	defaultMetaPartitionInodeIDStep  uint64 = 1 << 24
	defaultMetaNodeReservedMem       uint64 = 1 << 30
	spaceAvailableRate                      = 0.90
	defaultNodeSetCapacity                  = 18
	retrySendSyncTaskInternal               = 3 * time.Second
)

// configuration keys
const (
	ClusterName       = "clusterName"
	ID                = "id"
	IP                = "ip"
	Port              = "port"
	LogLevel          = "logLevel"
	WalDir            = "walDir"
	StoreDir          = "storeDir"
	GroupID           = 1
	ModuleName        = "master"
	CfgRetainLogs     = "retainLogs"
	DefaultRetainLogs = 20000
	cfgTickInterval   = "tickInterval"
	cfgElectionTick   = "electionTick"
)

// Keys in the request
const (
	addrKey               = "addr"
	diskPathKey           = "disk"
	nameKey               = "name"
	idKey                 = "id"
	countKey              = "count"
	startKey              = "start"
	enableKey             = "enable"
	thresholdKey          = "threshold"
	metaPartitionCountKey = "mpCount"
	volCapacityKey        = "capacity"
	volOwnerKey           = "owner"
	replicaNumKey         = "replicaNum"
	s3EndpointKey         = "endpoint"
	accessKeyKey          = "accesskey"
	secretKeyKey          = "secretkey"
	regionKey             = "region"
	createBackendKey      = "createBackend"
	deleteBackendKey      = "deleteBackend"
	clientIdKey           = "clientId"
	clientMemoryUsedKey   = "clientMemoryUsed"
)

// Server represents the server in a cluster
type Server struct {
	id           uint64
	clusterName  string
	ip           string
	port         string
	walDir       string
	storeDir     string
	Version      string
	retainLogs   uint64
	tickInterval int
	electionTick int
	leaderInfo   *LeaderInfo
	config       *clusterConfig
	cluster      *Cluster
	store        raftstore.Store
	raftStore    raftstore.RaftStore
	fsm          *MetadataFsm
	partition    raftstore.Partition
	wg           sync.WaitGroup
	reverseProxy *httputil.ReverseProxy
	metaReady    bool
}

func (server *Server) createRaftServer() error {
	var err error
	raftCfg := &raftstore.Config{
		NodeID:            server.id,
		RaftPath:          server.walDir,
		NumOfLogsToRetain: server.retainLogs,
		HeartbeatPort:     int(server.config.heartbeatPort),
		ReplicaPort:       int(server.config.replicaPort),
		TickInterval:      server.tickInterval,
		ElectionTick:      server.electionTick,
	}
	if server.raftStore, err = raftstore.NewRaftStore(raftCfg); err != nil {
		return fmt.Errorf("NewRaftStore failed! id[%v] walPath[%v] err: %v", server.id, server.walDir, err)
	}

	klog.Infof("peers[%v],tickInterval[%v],electionTick[%v]\n", server.config.peers, server.tickInterval, server.electionTick)
	server.fsm = newMetadataFsm(server.store, server.retainLogs, server.raftStore.RaftServer())
	server.fsm.registerLeaderChangeHandler(server.handleLeaderChange)
	server.fsm.registerPeerChangeHandler(server.handlePeerChange)
	server.fsm.registerApplySnapshotHandler(server.handleApplySnapshot)
	server.fsm.restore()

	partitionCfg := &raftstore.PartitionConfig{
		ID:      GroupID,
		Peers:   server.config.peers,
		Applied: server.fsm.GetApply(),
		SM:      server.fsm,
	}
	if server.partition, err = server.raftStore.CreatePartition(partitionCfg); err != nil {
		return fmt.Errorf("CreatePartition failed err %v", err)
	}

	return nil
}

func (server *Server) checkConfig(cfg *config.Config) (err error) {
	server.clusterName = cfg.GetString(ClusterName)
	server.ip = cfg.GetString(IP)
	server.port = cfg.GetString(Port)
	server.walDir = cfg.GetString(WalDir)
	server.storeDir = cfg.GetString(StoreDir)
	peerAddrs := cfg.GetString(cfgPeers)
	if server.ip == "" || server.port == "" || server.walDir == "" || server.storeDir == "" || server.clusterName == "" || peerAddrs == "" {
		return fmt.Errorf("bad configuration file,err: one of (ip,port,walDir,storeDir,clusterName) is null")
	}
	if server.id, err = strconv.ParseUint(cfg.GetString(ID), 10, 64); err != nil {
		return fmt.Errorf("bad configuration file,err:%v", err)
	}

	server.config.s3Endpoint = cfg.GetString(s3EndpointKey)
	if server.config.s3Endpoint == "" {
		return fmt.Errorf("bad configuration file,err:%v", "endpoint is null")
	}
	server.config.region = cfg.GetString(regionKey)
	if server.config.region == "" {
		server.config.region = defaultRegion
	}

	server.config.heartbeatPort = cfg.GetInt64(heartbeatPortKey)
	server.config.replicaPort = cfg.GetInt64(replicaPortKey)
	if server.config.heartbeatPort <= 1024 {
		server.config.heartbeatPort = raftstore.DefaultHeartbeatPort
	}
	if server.config.replicaPort <= 1024 {
		server.config.replicaPort = raftstore.DefaultReplicaPort
	}
	fmt.Printf("heartbeatPort[%v],replicaPort[%v]\n", server.config.heartbeatPort, server.config.replicaPort)
	if err = server.config.parsePeers(peerAddrs); err != nil {
		return
	}
	capacity := cfg.GetString(nodeSetCapacity)
	if capacity != "" {
		if server.config.nodeSetCapacity, err = strconv.Atoi(capacity); err != nil {
			return fmt.Errorf("bad configuration file,err:%v", err.Error())
		}
	}
	if server.config.nodeSetCapacity < 3 {
		server.config.nodeSetCapacity = defaultNodeSetCapacity
	}

	metaNodeReservedMemory := cfg.GetString(cfgMetaNodeReservedMem)
	if metaNodeReservedMemory != "" {
		if server.config.metaNodeReservedMem, err = strconv.ParseUint(metaNodeReservedMemory, 10, 64); err != nil {
			return fmt.Errorf("bad configuration file,err:%v", err.Error())
		}
	}
	if server.config.metaNodeReservedMem < 32*1024*1024 {
		server.config.metaNodeReservedMem = defaultMetaNodeReservedMem
	}

	retainLogs := cfg.GetString(CfgRetainLogs)
	if retainLogs != "" {
		if server.retainLogs, err = strconv.ParseUint(retainLogs, 10, 64); err != nil {
			return fmt.Errorf("bad configuration file,err:%v", err.Error())
		}
	}
	if server.retainLogs <= 0 {
		server.retainLogs = DefaultRetainLogs
	}

	server.tickInterval = int(cfg.GetFloat(cfgTickInterval))
	server.electionTick = int(cfg.GetFloat(cfgElectionTick))
	if server.tickInterval <= 300 {
		server.tickInterval = 500
	}
	if server.electionTick <= 3 {
		server.electionTick = 5
	}
	return
}

func (server *Server) Start(cfg *config.Config) error {
	server.config = &clusterConfig{
		NodeTimeOutSec:      defaultNodeTimeOutSec,
		MetaNodeThreshold:   defaultMetaPartitionMemUsageThreshold,
		metaNodeReservedMem: defaultMetaNodeReservedMem,
	}
	server.leaderInfo = &LeaderInfo{}
	server.reverseProxy = server.newReverseProxy()
	if err := server.checkConfig(cfg); err != nil {
		klog.Error(err)
		return err
	}

	var err error
	server.store, err = raftstore.NewMemoryStore()
	if err != nil {
		return fmt.Errorf("NewRocksDBStore error: %v", err)
	}

	if err = server.createRaftServer(); err != nil {
		klog.Error(err)
		return err
	}

	server.cluster = NewCluster(server.clusterName, server.leaderInfo, server.fsm, server.partition, server.config)
	server.cluster.retainLogs = server.retainLogs
	server.cluster.partition = server.partition
	server.cluster.idAlloc.partition = server.partition
	server.cluster.scheduleTask()

	// INFO: server http service
	server.startHTTPService()

	server.wg.Add(1)
	return nil
}

func (server *Server) Shutdown() {
	server.wg.Done()
}

func (server *Server) Wait() {
	server.wg.Wait()
}

// NewServer creates a new server
func NewServer() *Server {
	return &Server{}
}
