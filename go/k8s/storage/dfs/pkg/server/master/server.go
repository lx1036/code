package master

import (
	"fmt"
	"github.com/pingcap/errors"
	"net/http/httputil"
	"sync"

	"k8s-lx1036/k8s/storage/dfs/pkg/config"
	"k8s-lx1036/k8s/storage/dfs/pkg/raftstore"

	"k8s.io/klog/v2"
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

func (m *Server) createRaftServer() (err error) {
	raftCfg := &raftstore.Config{
		NodeID:            m.id,
		RaftPath:          m.walDir,
		NumOfLogsToRetain: m.retainLogs,
		HeartbeatPort:     int(m.config.heartbeatPort),
		ReplicaPort:       int(m.config.replicaPort),
		TickInterval:      m.tickInterval,
		ElectionTick:      m.electionTick,
	}
	if m.raftStore, err = raftstore.NewRaftStore(raftCfg); err != nil {
		return fmt.Errorf("NewRaftStore failed! id[%v] walPath[%v] err: %v", m.id, m.walDir, err)
	}

	klog.Infof("peers[%v],tickInterval[%v],electionTick[%v]\n", m.config.peers, m.tickInterval, m.electionTick)
	m.fsm = newMetadataFsm(m.rocksDBStore, m.retainLogs, m.raftStore.RaftServer())
	m.fsm.registerLeaderChangeHandler(m.handleLeaderChange)
	m.fsm.registerPeerChangeHandler(m.handlePeerChange)

	// register the handlers for the interfaces defined in the Raft library
	m.fsm.registerApplySnapshotHandler(m.handleApplySnapshot)
	m.fsm.restore()
	partitionCfg := &raftstore.PartitionConfig{
		ID:      GroupID,
		Peers:   m.config.peers,
		Applied: m.fsm.applied,
		SM:      m.fsm,
	}
	if m.partition, err = m.raftStore.CreatePartition(partitionCfg); err != nil {
		return errors.Trace(err, "CreatePartition failed")
	}
	return
}

func (srv *Server) Start(cfg *config.Config) error {

	srv.rocksDBStore, err = raftstore.NewRocksDBStore(srv.storeDir, LRUCacheSize, WriteBufferSize)
	if err != nil {
		klog.Errorf("NewRocksDBStore error: %v", err)
		return err
	}

	if err = srv.createRaftServer(); err != nil {
		klog.Error(err)
		return err
	}

	srv.cluster = newCluster(srv.clusterName, srv.leaderInfo, srv.fsm, srv.partition, srv.config)
	srv.cluster.retainLogs = srv.retainLogs
	srv.cluster.partition = srv.partition
	srv.cluster.idAlloc.partition = srv.partition
	srv.cluster.scheduleTask()
	srv.startHTTPService()

	srv.wg.Add(1)
	return nil
}

func (srv *Server) Shutdown() {
	srv.wg.Done()
}

func (srv *Server) Sync() {
	srv.wg.Wait()
}

// NewServer creates a new server
func NewServer() *Server {
	return &Server{}
}
