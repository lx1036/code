package raftstore

import (
	"github.com/tiglabs/raft"
	"github.com/tiglabs/raft/proto"
	"os"
)

// PartitionFsm wraps necessary methods include both FSM implementation
// and data storage operation for raft store partition.
// It extends from raft StateMachine and Store.
type PartitionFsm interface {
	raft.StateMachine
}

// Partition wraps necessary methods for raft store partition operation.
// Partition is a shard for multi-raft in RaftSore. RaftStore is based on multi-raft which
// manages multiple raft replication groups at same time through a single
// raft server instance and system resource.
type Partition interface {
	// Submit submits command data to raft log.
	Submit(cmd []byte) (resp interface{}, err error)

	// ChaneMember submits member change event and information to raft log.
	ChangeMember(changeType proto.ConfChangeType, peer proto.Peer, context []byte) (resp interface{}, err error)

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

// Default implementation of the Partition interface.
type partition struct {
	id      uint64
	raft    *raft.RaftServer
	walPath string
	config  *PartitionConfig
}

// Submit submits command data to raft log.
func (p *partition) Submit(cmd []byte) (resp interface{}, err error) {
	if !p.IsRaftLeader() {
		return nil, raft.ErrNotLeader
	}

	return p.raft.Submit(p.id, cmd).Response()
}

// ChaneMember submits member change event and information to raft log.
func (p *partition) ChangeMember(changeType proto.ConfChangeType, peer proto.Peer, context []byte) (interface{}, error) {
	if !p.IsRaftLeader() {
		return nil, raft.ErrNotLeader
	}

	return p.raft.ChangeMember(p.id, changeType, peer, context).Response()
}

// Stop removes the raft partition from raft server and shuts down this partition.
func (p *partition) Stop() error {
	return p.raft.RemoveRaft(p.id)
}

// Delete stops and deletes the partition.
func (p *partition) Delete() error {
	if err := p.Stop(); err != nil {
		return err
	}

	return os.RemoveAll(p.walPath)
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
	return p.raft.AppliedIndex(p.id)
}

// CommittedIndex returns the current index of the committed raft log in the raft store partition.
func (p *partition) CommittedIndex() uint64 {
	return p.raft.CommittedIndex(p.id)
}

func (p *partition) FirstCommittedIndex() uint64 {
	return p.raft.FirstCommittedIndex(p.id)
}

// Truncate truncates the raft log
func (p *partition) Truncate(index uint64) {
	if p.raft != nil {
		p.raft.Truncate(p.id, index)
	}
}

func (p *partition) TryToLeader(nodeID uint64) error {
	_, err := p.raft.TryToLeader(nodeID).Response()

	return err
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

func newPartition(cfg *PartitionConfig, raft *raft.RaftServer, walPath string) Partition {
	return &partition{
		id:      cfg.ID,
		raft:    raft,
		walPath: walPath,
		config:  cfg,
	}
}
