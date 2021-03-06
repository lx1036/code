package raftstore

import (
	"github.com/tiglabs/raft"
	"github.com/tiglabs/raft/proto"
)

// PartitionStatus is a type alias of raft.Status
type PartitionStatus = raft.Status

// PartitionFsm wraps necessary methods include both FSM implementation
// and data storage operation for raft store partition.
// It extends from raft StateMachine and Store.
type PartitionFsm interface {
	raft.StateMachine
	Store
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
	Status() (status *PartitionStatus)

	// LeaderTerm returns the current term of leader in the raft group. TODO what is term?
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

func (p partition) ChangeMember(changeType proto.ConfChangeType, peer proto.Peer, context []byte) (resp interface{}, err error) {
	panic("implement me")
}

func (p partition) Stop() error {
	panic("implement me")
}

func (p partition) Delete() error {
	panic("implement me")
}

func (p partition) Status() (status *PartitionStatus) {
	panic("implement me")
}

func (p partition) LeaderTerm() (leaderID, term uint64) {
	panic("implement me")
}

func (p *partition) IsRaftLeader() bool {
	return p.raft != nil && p.raft.IsLeader(p.id)
}

func (p partition) AppliedIndex() uint64 {
	panic("implement me")
}

func (p partition) CommittedIndex() uint64 {
	panic("implement me")
}

func (p partition) FirstCommittedIndex() uint64 {
	panic("implement me")
}

func (p partition) Truncate(index uint64) {
	panic("implement me")
}

func (p partition) TryToLeader(nodeID uint64) error {
	panic("implement me")
}

func (p partition) IsOfflinePeer() bool {
	panic("implement me")
}

func newPartition(cfg *PartitionConfig, raft *raft.RaftServer, walPath string) Partition {
	return &partition{
		id:      cfg.ID,
		raft:    raft,
		walPath: walPath,
		config:  cfg,
	}
}
