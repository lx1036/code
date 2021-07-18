package raftstore

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/tiglabs/raft"
	"github.com/tiglabs/raft/logger"
	"github.com/tiglabs/raft/proto"
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

// CreatePartition creates a new partition in the raft store.
func (r *raftStore) CreatePartition(cfg *PartitionConfig) (p Partition, err error) {
	// Init WaL Storage for this partition.
	// Variables:
	// wc: WaL Configuration.
	// wp: WaL Path.
	// ws: WaL Storage.

	var walPath string
	if cfg.WalPath == "" {
		walPath = path.Join(r.raftPath, strconv.FormatUint(cfg.ID, 10))
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
		r.AddNodeWithPort(
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
	}
	if err = r.raftServer.CreateRaft(rc); err != nil {
		return
	}
	p = newPartition(cfg, r.raftServer, walPath)

	return

}

func (r raftStore) Stop() {
	panic("implement me")
}

func (r raftStore) RaftConfig() *raft.Config {
	return r.raftConfig
}

func (r raftStore) RaftStatus(raftID uint64) (raftStatus *raft.Status) {
	return r.raftServer.Status(raftID)
}

func (r raftStore) AddNodeWithPort(nodeID uint64, addr string, heartbeat int, replicate int) {
	panic("implement me")
}

func (r raftStore) DeleteNode(nodeID uint64) {
	panic("implement me")
}

func (r raftStore) RaftServer() *raft.RaftServer {
	return r.raftServer
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
		fmt.Println("Fatal: failed to start the baud storage daemon - ", err)
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
