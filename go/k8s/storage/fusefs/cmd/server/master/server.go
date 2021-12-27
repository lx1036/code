package master

import (
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"k8s-lx1036/k8s/storage/fusefs/cmd/server/meta/partition/raftstore"
	"k8s-lx1036/k8s/storage/fusefs/pkg/util"

	"github.com/tiglabs/raft/proto"
	"k8s.io/klog/v2"
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
	Port        int    `json:"port"`

	// raft
	HeartbeatPort int    `json:"heartbeatPort"`
	ReplicaPort   int    `json:"replicaPort"`
	WalDir        string `json:"walDir"`   // wal raft log
	StoreDir      string `json:"storeDir"` // statemachine
	Peers         string `json:"peers"`
	RetainLogs    uint64 `json:"retainLogs"`

	// s3
	Endpoint string `json:"endpoint"`

	// cluster
	nodeSetCapacity int `json:"nodeSetCapacity" default:"18"`
}

// Server represents the server in a cluster
type Server struct {
	id          uint64
	clusterName string
	ip          string
	port        int

	// raft
	peers         []raftstore.PeerAddress
	heartbeatPort int
	replicaPort   int
	walDir        string // raft log wal
	retainLogs    uint64
	storeDir      string // boltdb statemachine
	fsm           *MetadataFsm
	leaderInfo    *LeaderInfo
	boltdbStore   *raftstore.BoltDBStore
	raftStore     *raftstore.RaftStore
	partition     raftstore.Partition

	// cluster
	//config     *clusterConfig
	cluster         *Cluster
	nodeSetCapacity int
}

// NewServer creates a new server
func NewServer(config Config) *Server {
	walDir, _ := filepath.Abs(config.WalDir)
	storeDir, _ := filepath.Abs(config.StoreDir)
	server := &Server{
		id:              config.ID,
		ip:              config.IP,
		port:            config.Port,
		heartbeatPort:   config.HeartbeatPort,
		replicaPort:     config.ReplicaPort,
		walDir:          walDir,
		storeDir:        storeDir,
		retainLogs:      config.RetainLogs,
		nodeSetCapacity: config.nodeSetCapacity,
		leaderInfo:      &LeaderInfo{},
	}

	peers := strings.Split(config.Peers, ",")
	for _, peer := range peers {
		values := strings.Split(peer, ":") // 1:127.0.0.1:9500
		id, _ := strconv.ParseUint(values[0], 10, 64)
		server.peers = append(server.peers, raftstore.PeerAddress{
			Peer: proto.Peer{
				ID: id,
			},
			Address:       values[1],
			HeartbeatPort: server.heartbeatPort,
			ReplicaPort:   server.replicaPort,
		})
	}

	return server
}

func (server *Server) Start() (err error) {

	klog.Info("afffff the master raft cluster")

	// 1. create a partition raft and statemachine
	if server.raftStore, err = raftstore.NewRaftStore(&raftstore.Config{
		NodeID:            server.id,
		IPAddr:            server.ip,
		HeartbeatPort:     server.heartbeatPort,
		ReplicaPort:       server.replicaPort,
		RaftPath:          server.walDir,
		NumOfLogsToRetain: server.retainLogs,
		TickInterval:      1000, // 1s
		ElectionTick:      5,
	}); err != nil {
		return err
	}
	server.boltdbStore, err = raftstore.NewBoltDBStore(server.storeDir)
	if err != nil {
		return err
	}

	klog.Info("afffaaaff the master raft cluster")

	server.fsm = newMetadataFsm(server.boltdbStore, server.retainLogs, server.raftStore.RaftServer())
	server.fsm.registerLeaderChangeHandler(server.handleLeaderChange)
	server.fsm.registerPeerChangeHandler(server.handlePeerChange)
	server.fsm.registerApplySnapshotHandler(server.handleApplySnapshot)
	server.fsm.restore()
	if server.partition, err = server.raftStore.CreatePartition(&raftstore.PartitionConfig{
		ID:      GroupID,
		Peers:   server.peers,
		Applied: server.fsm.applied,
		SM:      server.fsm,
	}); err != nil {
		return err
	}

	// 2. cluster -> partition raft
	server.cluster = NewCluster(server)
	server.cluster.start()
	server.startHTTPService()

	return nil
}

func (server *Server) Stop() {
	_ = server.boltdbStore.Close()
}
