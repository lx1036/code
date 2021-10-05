package raft

import (
	"errors"

	pb "go.etcd.io/etcd/raft/v3/raftpb"
)

// RawNode is a thread-unsafe Node.
type RawNode struct {
	raft       *raft
	prevSoftSt *SoftState
	prevHardSt pb.HardState
}

func NewRawNode(config *Config) (*RawNode, error) {
	r := newRaft(config)
	rn := &RawNode{
		raft: r,
	}
	rn.prevSoftSt = r.softState()
	rn.prevHardSt = r.hardState()
	return rn, nil
}

func (rawNode *RawNode) Bootstrap(peers []Peer) error {
	if len(peers) == 0 {
		return errors.New("must provide at least one peer to Bootstrap")
	}

	lastIndex := rawNode.raft.raftLog.storage.LastIndex()
	if lastIndex != 0 {
		return errors.New("can't bootstrap a nonempty Storage")
	}

	rawNode.prevHardSt = emptyState

	rawNode.raft.becomeFollower(1, None)
	ents := make([]pb.Entry, len(peers))
	for i, peer := range peers {
		cc := pb.ConfChange{Type: pb.ConfChangeAddNode, NodeID: peer.ID, Context: peer.Context}
		data, err := cc.Marshal()
		if err != nil {
			return err
		}

		ents[i] = pb.Entry{Type: pb.EntryConfChange, Term: 1, Index: uint64(i + 1), Data: data}
	}
	rawNode.raft.raftLog.append(ents...)

	rawNode.raft.raftLog.committed = uint64(len(ents))
	for _, peer := range peers {
		rawNode.raft.applyConfChange(pb.ConfChange{NodeID: peer.ID, Type: pb.ConfChangeAddNode}.AsV2())
	}

	return nil
}
