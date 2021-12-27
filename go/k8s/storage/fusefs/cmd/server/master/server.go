package master

import (
	"time"

	"k8s-lx1036/k8s/storage/fusefs/cmd/server/meta/partition/raftstore"
	"k8s-lx1036/k8s/storage/fusefs/pkg/util"
)

const (
	LRUCacheSize    = 3 << 30
	WriteBufferSize = 4 * util.MB
)

const (
	defaultInitMetaPartitionCount    = 3
	defaultMaxInitMetaPartitionCount = 100

	defaultMetaNodeReservedMem uint64 = 1 << 30
	spaceAvailableRate                = 0.90
	defaultNodeSetCapacity            = 18
	retrySendSyncTaskInternal         = 3 * time.Second
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
	replicaNumKey         = "replicaNum"
	s3EndpointKey         = "endpoint"

	secretKeyKey        = "secretkey"
	regionKey           = "region"
	createBackendKey    = "createBackend"
	deleteBackendKey    = "deleteBackend"
	clientIdKey         = "clientId"
	clientMemoryUsedKey = "clientMemoryUsed"
)

type Config struct {
	ID          uint64 `json:"id"`
	ClusterName string `json:"clusterName"`
	IP          string `json:"ip"`
	Port        string `json:"port"`

	// raft
	HeartbeatPort int    `json:"heartbeatPort"`
	ReplicaPort   int    `json:"replicaPort"`
	WalDir        string `json:"walDir"`   // wal raft log
	StoreDir      string `json:"storeDir"` // statemachine
	Peers         string `json:"peers"`
	RetainLogs    uint64 `json:"retainLogs"`

	// s3
	Endpoint string `json:"endpoint"`
}

// Server represents the server in a cluster
type Server struct {
	id          uint64
	clusterName string
	ip          string
	port        string

	// raft
	heartbeatPort int
	replicaPort   int
	walDir        string // raft log wal
	retainLogs    uint64
	storeDir      string // boltdb statemachine
	fsm           *MetadataFsm

	leaderInfo *LeaderInfo
	config     *clusterConfig
	cluster    *Cluster

	boltdbStore *raftstore.BoltDBStore
	raftStore   *raftstore.RaftStore
	partition   raftstore.Partition
}

// NewServer creates a new server
func NewServer(config Config) *Server {
	return &Server{
		id:            config.ID,
		ip:            config.IP,
		heartbeatPort: config.HeartbeatPort,
		replicaPort:   config.ReplicaPort,
		walDir:        config.WalDir,
		storeDir:      config.StoreDir,
		retainLogs:    config.RetainLogs,

		leaderInfo: &LeaderInfo{},
	}
}

func (server *Server) Start() (err error) {
	// 1. create a partition raft and statemachine
	if server.raftStore, err = raftstore.NewRaftStore(&raftstore.Config{
		NodeID:            server.id,
		IPAddr:            server.ip,
		HeartbeatPort:     server.heartbeatPort,
		ReplicaPort:       server.replicaPort,
		RaftPath:          server.walDir,
		NumOfLogsToRetain: server.retainLogs,
		TickInterval:      1000,  // 1s
		ElectionTick:      10000, // 10s
	}); err != nil {
		return err
	}
	server.boltdbStore, err = raftstore.NewBoltDBStore(server.storeDir)
	if err != nil {
		return err
	}
	server.fsm = newMetadataFsm(server.boltdbStore, server.retainLogs, server.raftStore.RaftServer())
	server.fsm.registerLeaderChangeHandler(server.handleLeaderChange)
	server.fsm.registerPeerChangeHandler(server.handlePeerChange)
	server.fsm.registerApplySnapshotHandler(server.handleApplySnapshot)
	server.fsm.restore()
	if server.partition, err = server.raftStore.CreatePartition(&raftstore.PartitionConfig{
		ID:      GroupID,
		Peers:   server.config.peers,
		Applied: server.fsm.applied,
		SM:      server.fsm,
	}); err != nil {
		return err
	}

	// 2. cluster -> partition raft
	server.cluster = NewCluster(server.clusterName, server.leaderInfo, server.fsm, server.partition, server.config)
	server.cluster.start()
	server.startHTTPService()

	return nil
}

func (server *Server) Stop() {
	_ = server.boltdbStore.Close()
}
