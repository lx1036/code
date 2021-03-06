package master

import (
	"fmt"
	"net/http/httputil"
	"sync"

	"k8s-lx1036/k8s/storage/dfs/pkg/raftstore"
	"k8s-lx1036/k8s/storage/dfs/pkg/util"
	"k8s-lx1036/k8s/storage/dfs/pkg/util/config"

	"k8s.io/klog/v2"
)

const (
	LRUCacheSize    = 3 << 30
	WriteBufferSize = 4 * util.MB
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
	rocksDBStore *raftstore.RocksDBStore
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
	server.fsm = newMetadataFsm(server.rocksDBStore, server.retainLogs, server.raftStore.RaftServer())
	server.fsm.registerLeaderChangeHandler(server.handleLeaderChange)
	server.fsm.registerPeerChangeHandler(server.handlePeerChange)

	// register the handlers for the interfaces defined in the Raft library
	server.fsm.registerApplySnapshotHandler(server.handleApplySnapshot)
	server.fsm.restore()
	partitionCfg := &raftstore.PartitionConfig{
		ID:      GroupID,
		Peers:   server.config.peers,
		Applied: server.fsm.applied,
		SM:      server.fsm,
	}
	if server.partition, err = server.raftStore.CreatePartition(partitionCfg); err != nil {
		return fmt.Errorf("CreatePartition failed err %v", err)
	}

	return nil
}

func (server *Server) Start(cfg *config.Config) error {

	var err error
	server.rocksDBStore, err = raftstore.NewRocksDBStore(server.storeDir, LRUCacheSize, WriteBufferSize)
	if err != nil {
		klog.Errorf("NewRocksDBStore error: %v", err)
		return err
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
	server.startHTTPService()

	server.wg.Add(1)
	return nil
}

func (server *Server) Shutdown() {
	server.wg.Done()
}

func (server *Server) Sync() {
	server.wg.Wait()
}

// NewServer creates a new server
func NewServer() *Server {
	return &Server{}
}
