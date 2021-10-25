package master

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/tiglabs/raft"
	"github.com/tiglabs/raft/proto"
	"github.com/tiglabs/raft/storage"
	bolt "go.etcd.io/bbolt"
	"k8s.io/klog/v2"
)

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
	Status() (status *raft.Status)

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
	SM      raft.StateMachine
	WalPath string
}

// Default implementation of the Partition interface.
type partition struct {
	id     uint64
	raft   *raft.RaftServer
	config *PartitionConfig
}

func newPartition(cfg *PartitionConfig, raft *raft.RaftServer) Partition {
	return &partition{
		id:     cfg.ID,
		raft:   raft,
		config: cfg,
	}
}

// Submit submits command data to raft log.
func (p *partition) Submit(cmd []byte) error {
	if !p.IsRaftLeader() {
		return fmt.Errorf("raft is not leader")
	}

	p.raft.Submit(p.id, cmd)
	return nil
}

// ChangeMember submits member change event and information to raft log.
func (p *partition) ChangeMember(changeType proto.ConfChangeType, peer proto.Peer, context []byte) error {
	if !p.IsRaftLeader() {
		return fmt.Errorf("raft is not leader")
	}

	p.raft.ChangeMember(p.id, changeType, peer, context)
	return nil
}

// Delete stops and deletes the partition.
func (p *partition) Delete() error {
	return p.Stop()

	//return os.RemoveAll(p.walPath)
}

// Status returns the current raft status.
func (p *partition) Status() (status *raft.Status) {
	return p.raft.Status(p.id)
}

// LeaderTerm returns the current term of leader in the raft group.
func (p *partition) LeaderTerm() (leader, term uint64) {
	return p.raft.LeaderTerm(p.id)
}

// IsRaftLeader returns true if this node is the leader of the raft group it belongs to.
func (p *partition) IsRaftLeader() bool {
	return p.raft != nil && p.raft.IsLeader(p.id)
}

// AppliedIndex returns the current index of the applied raft log in the raft store partition.
func (p *partition) AppliedIndex() uint64 {
	//return p.raft.AppliedIndex(p.id)
	panic("not implemented")
}

// CommittedIndex returns the current index of the committed raft log in the raft store partition.
func (p *partition) CommittedIndex() uint64 {
	//return p.raft.CommittedIndex(p.id)
	panic("not implemented")

}

func (p *partition) FirstCommittedIndex() uint64 {
	//return p.raft.FirstCommittedIndex(p.id)
	panic("not implemented")
}

// Truncate truncates the raft log
func (p *partition) Truncate(index uint64) {
	/*if p.raft != nil {
		p.raft.Truncate(p.id, index)
	}*/
}

func (p *partition) TryToLeader(nodeID uint64) error {
	panic("not implemented")
	/*_, err := p.raft.TryToLeader(nodeID).Response()

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
	//return p.raft.RemoveRaft(p.id)
	panic("not implemented")
}

var (
	ErrNoSuchNode        = errors.New("no such node")
	ErrIllegalAddress    = errors.New("illegal address")
	ErrUnknownSocketType = errors.New("unknown socket type")
)

// NodeManager defines the necessary methods for node address management.
type NodeManager interface {
	// add node address with specified port.
	AddNodeWithPort(nodeID uint64, addr string, heartbeat int, replicate int)

	// delete node address information
	DeleteNode(nodeID uint64)
}

// NodeResolver defines the methods for node address resolving and management.
// It is extended from SocketResolver and NodeManager.
type NodeResolver interface {
	raft.SocketResolver
	NodeManager
}

// This private struct defines the necessary properties for node address info.
type nodeAddress struct {
	Heartbeat string
	Replicate string
}

// Default thread-safe implementation of the NodeResolver interface.
type nodeResolver struct {
	nodeMap sync.Map
}

func (resolver *nodeResolver) NodeAddress(nodeID uint64, socketType raft.SocketType) (addr string, err error) {
	val, ok := resolver.nodeMap.Load(nodeID)
	if !ok {
		err = ErrNoSuchNode
		return
	}

	address, ok := val.(*nodeAddress)
	if !ok {
		err = ErrIllegalAddress
		return
	}
	switch socketType {
	case raft.HeartBeat:
		addr = address.Heartbeat
	case raft.Replicate:
		addr = address.Replicate
	default:
		err = ErrUnknownSocketType
	}

	return
}

func (resolver *nodeResolver) AddNodeWithPort(nodeID uint64, addr string, heartbeat int, replicate int) {
	if heartbeat == 0 {
		heartbeat = DefaultHeartbeatPort
	}
	if replicate == 0 {
		replicate = DefaultReplicaPort
	}
	if len(strings.TrimSpace(addr)) != 0 {
		resolver.nodeMap.Store(nodeID, &nodeAddress{
			Heartbeat: fmt.Sprintf("%s:%d", addr, heartbeat),
			Replicate: fmt.Sprintf("%s:%d", addr, replicate),
		})
	}
}

func (resolver *nodeResolver) DeleteNode(nodeID uint64) {
	resolver.nodeMap.Delete(nodeID)
}

// NewNodeResolver returns a new NodeResolver instance for node address management and resolving.
func NewNodeResolver() NodeResolver {
	return &nodeResolver{}
}

type RaftNode interface {
	CreatePartition(cfg *PartitionConfig) (Partition, error)
	Stop()
	RaftConfig() *raft.Config
	RaftStatus(raftID uint64) (raftStatus *raft.Status)
	NodeManager
	RaftNode() *raft.RaftServer
}

type raftNode struct {
	nodeID   uint64
	resolver NodeResolver

	raftNodeConfig *raft.Config
	raftNode       *raft.RaftServer

	raftPath string
}

// NewRaftNode returns a new raft store instance.
func NewRaftNode(config *Config) (RaftNode, error) {
	config = config.validate()

	resolver := NewNodeResolver()
	nodeConfig := raft.DefaultConfig()
	nodeConfig.NodeID = config.NodeID
	nodeConfig.LeaseCheck = true
	nodeConfig.HeartbeatAddr = fmt.Sprintf("%s:%d", config.IPAddr, config.HeartbeatPort)
	nodeConfig.ReplicateAddr = fmt.Sprintf("%s:%d", config.IPAddr, config.ReplicaPort)
	nodeConfig.Resolver = resolver
	nodeConfig.TickInterval = time.Duration(config.TickInterval) * time.Millisecond
	nodeConfig.ElectionTick = config.ElectionTick
	node, err := raft.NewRaftServer(nodeConfig)
	if err != nil {
		return nil, err
	}

	store := &raftNode{
		nodeID:         config.NodeID,
		resolver:       resolver,
		raftNodeConfig: nodeConfig,
		raftNode:       node,
		raftPath:       config.RaftPath,
	}

	return store, nil
}

// CreatePartition INFO: 每一个 partition 都是一个 raft
func (node *raftNode) CreatePartition(cfg *PartitionConfig) (Partition, error) {
	memoryStorage := storage.NewMemoryStorage(node.raftNode, 1, 8192)

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
	raftConfig := &raft.RaftConfig{
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

func (node *raftNode) RaftConfig() *raft.Config {
	return node.raftNodeConfig
}

func (node *raftNode) RaftStatus(raftID uint64) (raftStatus *raft.Status) {
	panic("not implemented")
	//return node.raftNode.Status(raftID)
}

// AddNodeWithPort add a new node with the given port.
func (node *raftNode) AddNodeWithPort(nodeID uint64, addr string, heartbeat int, replicate int) {
	if node.resolver != nil {
		node.resolver.AddNodeWithPort(nodeID, addr, heartbeat, replicate)
	}
}

func (node *raftNode) RaftNode() *raft.RaftServer {
	return node.raftNode
}

func (node *raftNode) DeleteNode(nodeID uint64) {
	panic("implement me")
}

func (node *raftNode) Stop() {
	node.raftNode.Stop()
}

const (
	DefaultBucket = "default"
)

type BoltdbStore struct {
	*bolt.DB
}

func (store *BoltdbStore) Get(key []byte) ([]byte, error) {
	var value []byte
	err := store.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(DefaultBucket))
		if err != nil {
			return err
		}
		value = bucket.Get(key)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (store *BoltdbStore) Put(key, value []byte) error {
	err := store.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(DefaultBucket))
		if err != nil {
			return err
		}
		return bucket.Put(key, value)
	})

	return err
}

func (store *BoltdbStore) Delete(key []byte) error {
	err := store.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(DefaultBucket))
		if err != nil {
			return err
		}
		return bucket.Delete(key)
	})

	return err
}

func (store *BoltdbStore) Close() error {
	return store.DB.Close()
}

// NewBoltdbStore "./raft/my.db"
func NewBoltdbStore(dbPath string) *BoltdbStore {
	db, err := bolt.Open(dbPath, 0666, nil)
	if err != nil {
		klog.Fatalf("init boltdb failed: %v, path: %v", err, dbPath)
	}

	return &BoltdbStore{DB: db}
}
