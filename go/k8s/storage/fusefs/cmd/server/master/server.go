package master

import (
	"fmt"
	boltdb "k8s-lx1036/k8s/storage/raft/hashicorp/bolt-store"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"k8s-lx1036/k8s/storage/fusefs/cmd/server/meta/partition/raftstore"
	"k8s-lx1036/k8s/storage/fusefs/pkg/util"

	"github.com/hashicorp/raft"
	"k8s.io/klog/v2"
)

const (
	LRUCacheSize    = 3 << 30
	WriteBufferSize = 4 * util.MB
)

const (
	defaultInitMetaPartitionCount    = 3
	defaultMaxInitMetaPartitionCount = 100

	spaceAvailableRate        = 0.90
	defaultNodeSetCapacity    = 18
	retrySendSyncTaskInternal = 3 * time.Second
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
	HttpPort    int    `json:"httpPort"`

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
	localIP     string
	httpPort    int
	s3Endpoint  string

	cluster         *Cluster
	nodeSetCapacity int

	// raft
	r        *raft.Raft
	peers    []raft.Server
	isLeader bool
	walDir   string // raft log wal
	//leader   raft.ServerAddress

	retainLogs  uint64
	storeDir    string // boltdb statemachine
	fsm         *Fsm
	boltdbStore *raftstore.BoltDBStore
	raftStore   *raftstore.RaftStore
	partition   raftstore.Partition
}

// NewServer creates a new server
func NewServer(config Config) *Server {
	walDir, _ := filepath.Abs(config.WalDir)
	storeDir, _ := filepath.Abs(config.StoreDir)
	server := &Server{
		id:              config.ID,
		httpPort:        config.HttpPort,
		walDir:          walDir,
		storeDir:        storeDir,
		retainLogs:      config.RetainLogs,
		nodeSetCapacity: config.nodeSetCapacity,
		s3Endpoint:      config.Endpoint,
	}

	var localRaftAddr string
	peers := strings.Split(config.Peers, ",")
	for _, peer := range peers {
		values := strings.Split(peer, "/") // 1/127.0.0.1:9500
		id, _ := strconv.ParseUint(values[0], 10, 64)
		addr := values[1] // 127.0.0.1:9500
		values = strings.Split(values[1], ":")
		if id == server.id {
			server.localIP = values[0] // 127.0.0.1
		}
		server.peers = append(server.peers, raft.Server{
			Suffrage: raft.Voter,
			ID:       raft.ServerID(strconv.FormatUint(id, 10)),
			Address:  raft.ServerAddress(addr),
		})
	}

	if len(localRaftAddr) == 0 {
		klog.Fatal(fmt.Sprintf("local raft addr is empty"))
	}
	addr, err := net.ResolveTCPAddr("tcp", localRaftAddr)
	if err != nil {
		klog.Fatal(err)
	}
	transport, err := raft.NewTCPTransport(localRaftAddr, addr, 2, 5*time.Second, os.Stderr)
	if err != nil {
		klog.Fatal(err)
	}
	raftDir := fmt.Sprintf("%s/raft_%d", walDir, server.id) // raft log 存储在目录下 ./tmp/master1/wal/raft_1/raft-log.db
	store, err := boltdb.NewBoltStore(filepath.Join(raftDir, "raft-log.db"))
	if err != nil {
		klog.Fatal(err)
	}
	// ./tmp/master1/wal/raft_1/snapshots/{snapshotID}/{meta.json,state.bin}
	snapshots, err := raft.NewFileSnapshotStore(raftDir, 2, os.Stderr)
	if err != nil {
		klog.Fatal(err)
	}
	c := raft.DefaultConfig()
	c.LocalID = raft.ServerID(strconv.FormatUint(server.id, 10))
	fsm := newFsm(store) // INFO: 这是 fsm boltdb 文件，一般尽量和 raft-log boltdb 文件，是两个文件。这里仅仅测试，使用同一个就行。
	r, err := raft.NewRaft(c, fsm, store, store, snapshots, transport)
	if err != nil {
		klog.Fatal(err)
	}
	server.r = r

	return server
}

func (server *Server) Start() (err error) {
	// 1. create a partition raft
	server.r.BootstrapCluster(raft.Configuration{
		Servers: server.peers,
	})
	go server.watchLeaderCh()
	klog.Infof(fmt.Sprintf("raft is started"))

	// 2. cluster -> partition raft
	server.cluster = NewCluster(server)
	server.cluster.start()

	// 3. start http server
	server.startHTTPService()

	return nil
}

func (server *Server) watchLeaderCh() {
	for leader := range server.r.LeaderCh() {
		server.isLeader = leader
	}
}

func (server *Server) isRaftLeader() bool {
	return server.isLeader
}

func (server *Server) Stop() {
	server.raftStore.Stop()
	_ = server.boltdbStore.Close()
}
