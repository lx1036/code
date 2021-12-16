package raftstore

import (
	"fmt"
	syslog "log"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/tiglabs/raft"
	"github.com/tiglabs/raft/logger"
	"github.com/tiglabs/raft/storage/wal"
	raftlog "github.com/tiglabs/raft/util/log"
)

// RaftStore defines the interface for the raft store.
type RaftStore interface {
	CreatePartition(cfg *PartitionConfig) (Partition, error)
	Stop()
	RaftConfig() *raft.Config
	RaftStatus(raftID uint64) (raftStatus *raft.Status)
	NodeManager
	RaftServer() *raft.RaftServer
}

type raftStore struct {
	nodeID     uint64
	resolver   NodeResolver
	raftConfig *raft.Config
	raftServer *raft.RaftServer
	raftPath   string
}

// RaftConfig returns the raft configuration.
func (s *raftStore) RaftConfig() *raft.Config {
	return s.raftConfig
}

func (s *raftStore) RaftStatus(raftID uint64) (raftStatus *raft.Status) {
	return s.raftServer.Status(raftID)
}

// AddNodeWithPort add a new node with the given port.
func (s *raftStore) AddNodeWithPort(nodeID uint64, addr string, heartbeat int, replicate int) {
	s.resolver.AddNodeWithPort(nodeID, addr, heartbeat, replicate)
}

// DeleteNode deletes the node with the given ID in the raft store.
func (s *raftStore) DeleteNode(nodeID uint64) {
	s.resolver.DeleteNode(nodeID)
}

// Stop stops the raft store server.
func (s *raftStore) Stop() {
	if s.raftServer != nil {
		s.raftServer.Stop()
	}
}

func newRaftLogger(dir string) {

	raftLogPath := path.Join(dir, "logs")
	_, err := os.Stat(raftLogPath)
	if err != nil {
		if pathErr, ok := err.(*os.PathError); ok {
			if os.IsNotExist(pathErr) {
				os.MkdirAll(raftLogPath, 0755)
			}
		}
	}

	raftLog, err := raftlog.NewLog(raftLogPath, "raft", "debug")
	if err != nil {
		syslog.Println("Fatal: failed to start the baud storage daemon - ", err)
		return
	}
	logger.SetLogger(raftLog)
	return
}

// NewRaftStore returns a new raft store instance.
func NewRaftStore(cfg *Config) (mr RaftStore, err error) {
	resolver := NewNodeResolver()

	newRaftLogger(cfg.RaftPath)

	rc := raft.DefaultConfig()
	rc.NodeID = cfg.NodeID
	rc.LeaseCheck = true
	if cfg.HeartbeatPort <= 0 {
		cfg.HeartbeatPort = DefaultHeartbeatPort
	}
	if cfg.ReplicaPort <= 0 {
		cfg.ReplicaPort = DefaultReplicaPort
	}
	if cfg.NumOfLogsToRetain == 0 {
		cfg.NumOfLogsToRetain = DefaultNumOfLogsToRetain
	}
	if cfg.ElectionTick < DefaultElectionTick {
		cfg.ElectionTick = DefaultElectionTick
	}
	if cfg.TickInterval < DefaultTickInterval {
		cfg.TickInterval = DefaultTickInterval
	}
	// if cfg's RecvBufSize bigger than the default 2048,
	// use the bigger one.
	if cfg.RecvBufSize > rc.ReqBufferSize {
		rc.ReqBufferSize = cfg.RecvBufSize
	}
	rc.HeartbeatAddr = fmt.Sprintf("%s:%d", cfg.IPAddr, cfg.HeartbeatPort)
	rc.ReplicateAddr = fmt.Sprintf("%s:%d", cfg.IPAddr, cfg.ReplicaPort)
	rc.Resolver = resolver
	rc.RetainLogs = cfg.NumOfLogsToRetain
	rc.TickInterval = time.Duration(cfg.TickInterval) * time.Millisecond
	rc.ElectionTick = cfg.ElectionTick
	rs, err := raft.NewRaftServer(rc)
	if err != nil {
		return
	}
	mr = &raftStore{
		nodeID:     cfg.NodeID,
		resolver:   resolver,
		raftConfig: rc,
		raftServer: rs,
		raftPath:   cfg.RaftPath,
	}
	return
}

func (s *raftStore) RaftServer() *raft.RaftServer {
	return s.raftServer
}

// CreatePartition creates a new partition in the raft store.
func (s *raftStore) CreatePartition(cfg *PartitionConfig) (p Partition, err error) {
	// Init WaL Storage for this partition.
	// Variables:
	// wc: WaL Configuration.
	// wp: WaL Path.
	// ws: WaL Storage.
	var walPath string
	if cfg.WalPath == "" {
		walPath = path.Join(s.raftPath, strconv.FormatUint(cfg.ID, 10))
	} else {
		walPath = path.Join(cfg.WalPath, "wal_"+strconv.FormatUint(cfg.ID, 10))
	}

	wc := &wal.Config{}
	ws, err := wal.NewStorage(walPath, wc)
	if err != nil {
		return
	}
	peers := make([]proto.Peer, 0)
	for _, peerAddress := range cfg.Peers {
		peers = append(peers, peerAddress.Peer)
		s.AddNodeWithPort(
			peerAddress.ID,
			peerAddress.Address,
			peerAddress.HeartbeatPort,
			peerAddress.ReplicaPort,
		)
	}
	rc := &raft.RaftConfig{
		ID:           cfg.ID,
		Peers:        peers,
		Leader:       cfg.Leader,
		Term:         cfg.Term,
		Storage:      ws,
		StateMachine: cfg.SM,
		Applied:      cfg.Applied,
		Monitor:      newMonitor(),
	}
	if err = s.raftServer.CreateRaft(rc); err != nil {
		return
	}
	p = newPartition(cfg, s.raftServer, walPath)
	return
}
