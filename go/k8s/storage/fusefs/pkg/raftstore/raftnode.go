package raftstore

import (
	"fmt"
	"time"

	"k8s-lx1036/k8s/storage/etcd/multiraft"
	"k8s-lx1036/k8s/storage/raft/proto"
	//multiraft "github.com/tiglabs/raft"
)

// INFO: 参考 https://github.com/tiglabs/raft/blob/master/test/testserver.go

// Constants for network port definition.
const (
	DefaultHeartbeatPort     = 5901
	DefaultReplicaPort       = 5902
	DefaultNumOfLogsToRetain = 20000
	DefaultTickInterval      = 300
	DefaultElectionTick      = 3
)

// Config defines the configuration properties for the raft node.
type Config struct {
	NodeID            uint64 // Identity of raft server instance.
	RaftPath          string // Path of raft logs
	IPAddr            string // IP address
	HeartbeatPort     int
	ReplicaPort       int
	NumOfLogsToRetain uint64 // number of logs to be kept after truncation. The default value is 20000.

	// TickInterval is the interval of timer which check heartbeat and election timeout.
	// The default value is 300,unit is millisecond.
	TickInterval int

	// ElectionTick is the election timeout. If a follower does not receive any message
	// from the leader of current term during ElectionTick, it will become candidate and start an election.
	// ElectionTick must be greater than HeartbeatTick.
	// We suggest to use ElectionTick = 10 * HeartbeatTick to avoid unnecessary leader switching.
	// The default value is 1s.
	ElectionTick int
}

func (config *Config) validate() *Config {
	if config.HeartbeatPort <= 0 {
		config.HeartbeatPort = DefaultHeartbeatPort
	}
	if config.ReplicaPort <= 0 {
		config.ReplicaPort = DefaultReplicaPort
	}
	if config.NumOfLogsToRetain == 0 {
		config.NumOfLogsToRetain = DefaultNumOfLogsToRetain
	}
	if config.ElectionTick < DefaultElectionTick {
		config.ElectionTick = DefaultElectionTick
	}
	if config.TickInterval < DefaultTickInterval {
		config.TickInterval = DefaultTickInterval
	}

	return config
}

// PeerAddress defines the set of addresses that will be used by the peers.
type PeerAddress struct {
	proto.Peer
	Address       string
	HeartbeatPort int
	ReplicaPort   int
}

// PartitionConfig defines the configuration properties for the partitions.
type PartitionConfig struct {
	ID      uint64
	Applied uint64
	Leader  uint64
	Term    uint64
	Peers   []PeerAddress
	SM      PartitionFsm
	WalPath string
}

type RaftNode interface {
	CreatePartition(cfg *PartitionConfig) (Partition, error)
	Stop()
	RaftConfig() *multiraft.NodeConfig
	RaftStatus(raftID uint64) (raftStatus *multiraft.Status)
	NodeManager
	RaftNode() *multiraft.Node
}

type raftNode struct {
	nodeID   uint64
	resolver NodeResolver

	raftNodeConfig *multiraft.NodeConfig
	raftNode       *multiraft.Node

	raftPath string
}

// NewRaftNode returns a new raft store instance.
func NewRaftNode(config *Config) (RaftNode, error) {
	config = config.validate()

	resolver := NewNodeResolver()
	nodeConfig := multiraft.DefaultConfig()
	nodeConfig.NodeID = config.NodeID
	nodeConfig.LeaseCheck = true
	nodeConfig.HeartbeatAddr = fmt.Sprintf("%s:%d", config.IPAddr, config.HeartbeatPort)
	nodeConfig.ReplicateAddr = fmt.Sprintf("%s:%d", config.IPAddr, config.ReplicaPort)
	nodeConfig.Resolver = resolver
	nodeConfig.TickInterval = time.Duration(config.TickInterval) * time.Millisecond
	nodeConfig.ElectionTick = config.ElectionTick
	n, err := multiraft.NewNode(nodeConfig)
	if err != nil {
		return nil, err
	}

	store := &raftNode{
		nodeID:         config.NodeID,
		resolver:       resolver,
		raftNodeConfig: nodeConfig,
		raftNode:       n,
		raftPath:       config.RaftPath,
	}

	return store, nil
}

// CreatePartition INFO: 每一个 partition 都是一个 raft
func (node *raftNode) CreatePartition(cfg *PartitionConfig) (Partition, error) {
	memoryStorage := multiraft.NewStorage(node.raftNode)

	peers := make([]proto.Peer, 0)
	for _, peerAddress := range cfg.Peers {
		peers = append(peers, peerAddress.Peer)
		node.AddNodeWithPort(
			peerAddress.ID,
			peerAddress.Address,
			peerAddress.HeartbeatPort,
			peerAddress.ReplicaPort,
		)
	}
	raftConfig := &multiraft.RaftConfig{
		ID:           cfg.ID,
		Peers:        peers,
		Leader:       cfg.Leader,
		Term:         cfg.Term,
		Storage:      memoryStorage,
		StateMachine: cfg.SM,
		Applied:      cfg.Applied,
	}
	if err := node.raftNode.CreateRaft(raftConfig); err != nil {
		return nil, err
	}

	return newPartition(cfg, node.raftNode), nil
}

func (node *raftNode) RaftConfig() *multiraft.NodeConfig {
	return node.raftNodeConfig
}

func (node *raftNode) RaftStatus(raftID uint64) (raftStatus *multiraft.Status) {
	panic("not implemented")
	//return node.raftNode.Status(raftID)
}

// AddNodeWithPort add a new node with the given port.
func (node *raftNode) AddNodeWithPort(nodeID uint64, addr string, heartbeat int, replicate int) {
	if node.resolver != nil {
		node.resolver.AddNodeWithPort(nodeID, addr, heartbeat, replicate)
	}
}

func (node *raftNode) RaftNode() *multiraft.Node {
	return node.raftNode
}

func (node *raftNode) DeleteNode(nodeID uint64) {
	panic("implement me")
}

func (node *raftNode) Stop() {
	node.raftNode.Stop()
}

// PartitionFsm wraps necessary methods include both FSM implementation
// and data storage operation for raft store partition.
// It extends from raft StateMachine and Store.
type PartitionFsm interface {
	multiraft.StateMachine
}

// Partition wraps necessary methods for raft store partition operation.
// Partition is a shard for multi-raft in RaftSore. RaftStore is based on multi-raft which
// manages multiple raft replication groups at same time through a single
// raft server instance and system resource.
type Partition interface {
	// Submit submits command data to raft log.
	Submit(cmd []byte) error

	// ChaneMember submits member change event and information to raft log.
	ChangeMember(changeType proto.ConfChangeType, peer proto.Peer, context []byte) error

	// Stop removes the raft partition from raft server and shuts down this partition.
	Stop() error

	// Delete stops and deletes the partition.
	Delete() error

	// Status returns the current raft status.
	Status() (status *multiraft.Status)

	// LeaderTerm returns the current term of leader in the raft group.
	LeaderTerm() (leaderID, term uint64)

	// IsRaftLeader returns true if this node is the leader of the raft group it belongs to.
	IsRaftLeader() bool

	// AppliedIndex returns the current index of the applied raft log in the raft store partition.
	AppliedIndex() uint64

	// CommittedIndex returns the current index of the applied raft log in the raft store partition.
	CommittedIndex() uint64

	// FirstCommittedIndex returns the first committed index of raft log in the raft store partition.
	FirstCommittedIndex() uint64

	// Truncate raft log
	Truncate(index uint64)

	TryToLeader(nodeID uint64) error

	IsOfflinePeer() bool
}

// Default implementation of the Partition interface.
type partition struct {
	id   uint64
	node *multiraft.Node
	//walPath string
	config *PartitionConfig
}

func newPartition(cfg *PartitionConfig, node *multiraft.Node) Partition {
	return &partition{
		id:   cfg.ID,
		node: node,
		//walPath: walPath,
		config: cfg,
	}
}

// Submit submits command data to raft log.
func (p *partition) Submit(cmd []byte) error {
	if !p.IsRaftLeader() {
		return fmt.Errorf("raft is not leader")
	}

	p.node.Propose(p.id, cmd)
	return nil
}

// ChangeMember submits member change event and information to raft log.
func (p *partition) ChangeMember(changeType proto.ConfChangeType, peer proto.Peer, context []byte) error {
	if !p.IsRaftLeader() {
		return fmt.Errorf("raft is not leader")
	}

	return p.node.ChangeMember(p.id, changeType, peer, context)
}

// Delete stops and deletes the partition.
func (p *partition) Delete() error {
	return p.Stop()

	//return os.RemoveAll(p.walPath)
}

// Status returns the current raft status.
func (p *partition) Status() (status *multiraft.Status) {
	return p.node.Status(p.id)
}

// LeaderTerm returns the current term of leader in the raft group.
func (p *partition) LeaderTerm() (leader, term uint64) {
	return p.node.LeaderTerm(p.id)
}

// IsRaftLeader returns true if this node is the leader of the raft group it belongs to.
func (p *partition) IsRaftLeader() bool {
	return p.node != nil && p.node.IsLeader(p.id)
}

// AppliedIndex returns the current index of the applied raft log in the raft store partition.
func (p *partition) AppliedIndex() uint64 {
	//return p.node.AppliedIndex(p.id)
	panic("not implemented")
}

// CommittedIndex returns the current index of the committed raft log in the raft store partition.
func (p *partition) CommittedIndex() uint64 {
	//return p.node.CommittedIndex(p.id)
	panic("not implemented")

}

func (p *partition) FirstCommittedIndex() uint64 {
	//return p.node.FirstCommittedIndex(p.id)
	panic("not implemented")
}

// Truncate truncates the raft log
func (p *partition) Truncate(index uint64) {
	/*if p.node != nil {
		p.node.Truncate(p.id, index)
	}*/
}

func (p *partition) TryToLeader(nodeID uint64) error {
	panic("not implemented")
	/*_, err := p.node.TryToLeader(nodeID).Response()

	return err*/
}

func (p *partition) IsOfflinePeer() bool {
	status := p.Status()
	active := 0
	sumPeers := 0
	for _, peer := range status.Replicas {
		if peer.Active == true {
			active++
		}
		sumPeers++
	}

	return active >= (int(sumPeers)/2 + 1)
}

// Stop removes the raft partition from raft server and shuts down this partition.
func (p *partition) Stop() error {
	//return p.node.RemoveRaft(p.id)
	panic("not implemented")
}
