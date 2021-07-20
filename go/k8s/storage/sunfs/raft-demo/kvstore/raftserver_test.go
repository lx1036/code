package main

import (
	"errors"
	"fmt"
	"testing"

	"github.com/tiglabs/raft"
	"github.com/tiglabs/raft/proto"
	"github.com/tiglabs/raft/storage/wal"
)

const (
	NodeID          uint64 = 1      // ID of current node
	electionTickSec int    = 5      // Seconds of election tick
	heartbeatPort   string = "2019" // port used to process heartbeat
	replicatePort   string = "2020" // port used to process replication
	raftGroupID     uint64 = 1
)

// Errors
var (
	ErrNodeNotExists     = errors.New("node not exists")
	ErrUnknownSocketType = errors.New("unknown socket type")
)

// Implementation of interface 'github.com/tiglabs/raft/Resolver'.
type resolver struct {
	nodeMap map[uint64]string // node ID -> node address
}

func (r *resolver) NodeAddress(nodeID uint64, stype raft.SocketType) (addr string, err error) {
	var address string
	var found bool
	if address, found = r.nodeMap[nodeID]; !found {
		return "", ErrNodeNotExists
	}
	switch stype {
	case raft.HeartBeat:
		return fmt.Sprintf("%s:%s", address, heartbeatPort), nil
	case raft.Replicate:
		return fmt.Sprintf("%s:%s", address, replicatePort), nil
	default:
		return "", ErrUnknownSocketType
	}
}

func newResolver() *resolver {
	return &resolver{
		nodeMap: make(map[uint64]string),
	}
}

// Implementation of interface 'github.com/tiglabs/raft/StateMachine'
type stateMachine struct {
}

// invoked when some node need a snapshot to recover
func (s *stateMachine) Apply(command []byte, index uint64) (interface{}, error) {
	panic("implement me")
}

// invoked when some node need a snapshot to recover
func (s *stateMachine) ApplyMemberChange(confChange *proto.ConfChange, index uint64) (interface{}, error) {
	panic("implement me")
}

// invoked when some node need a snapshot to recover
func (s *stateMachine) Snapshot() (proto.Snapshot, error) {
	panic("implement me")
}

// invoked when current node applying snapshot
func (s *stateMachine) ApplySnapshot(peers []proto.Peer, iter proto.SnapIterator) error {
	panic("implement me")
}

// invoked when leader of raft group changed.
func (s *stateMachine) HandleFatalEvent(err *raft.FatalError) {
	panic("implement me")
}

// invoked when leader of raft group changed.
func (s *stateMachine) HandleLeaderChange(leader uint64) {
	panic("implement me")
}

type Node struct {
	NodeID uint64
}

var (
	nodes = []Node{
		{NodeID: 1},
	}
)

// INFO: https://github.com/tiglabs/raft/issues/21#issuecomment-593958599
func TestRaftServer(test *testing.T) {
	var err error

	// configuration for raft instance
	var config = raft.DefaultConfig()
	config.NodeID = NodeID                     // setup node ID
	config.ElectionTick = electionTickSec      // setup election tick
	config.LeaseCheck = true                   // use the lease mechanism
	config.HeartbeatAddr = ":" + heartbeatPort // setup heartbeat port
	config.ReplicateAddr = ":" + replicatePort // setup replicate port
	config.Resolver = newResolver()            // setup address resolver

	// init and start raft server instance
	var raftServer *raft.RaftServer
	if raftServer, err = raft.NewRaftServer(config); err != nil {
		panic(err) // start fail
	}

	// setup and create a raft group instance from raft server instance
	var raftConfig = &raft.RaftConfig{
		ID:           raftGroupID,
		Term:         0,                     // term of this raft group which need be managed by yourself
		Peers:        make([]proto.Peer, 0), // peers of this raft group members
		StateMachine: &stateMachine{},       // setup implementation instance for event handle
	}

	for _, node := range nodes {
		raftConfig.Peers = append(raftConfig.Peers, proto.Peer{
			Type:   proto.PeerNormal,
			ID:     node.NodeID,
			PeerID: node.NodeID,
		})
	}

	// setup WAL storage for raft this raft group instance
	var path = fmt.Sprintf("raft/%d", raftGroupID)
	if raftConfig.Storage, err = wal.NewStorage(path, &wal.Config{}); err != nil {
		panic(err) // init storage fail
	}
	if err = raftServer.CreateRaft(raftConfig); err != nil {
		panic(err) // start raft group instance fail
	}

	// submit a log to raft group 1 if current node is leader.
	if raftServer.IsLeader(raftGroupID) {
		// submit
		var future = raftServer.Submit(raftGroupID, []byte("hello raft"))
		// wait for result
		var resp interface{}
		if resp, err = future.Response(); err != nil {
			panic(err) // responded an error
		}
		fmt.Printf("response: %v\n", resp)
	}

	// get committed index ID of raft group '1'
	var committedIndex = raftServer.CommittedIndex(raftGroupID)
	fmt.Printf("committed index: %v\n", committedIndex)

	// stop raft group '1'
	if err = raftServer.RemoveRaft(raftGroupID); err != nil {
		panic(err) // stop raft group instance fail
	}

	// stop raft server instance
	raftServer.Stop()
}
