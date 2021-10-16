package raft

import (
	"errors"

	pb "go.etcd.io/etcd/raft/v3/raftpb"
)

// RawNode INFO: RawNode 是一个非线程安全的 Node，包含 raft 对象!!!
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

// Tick INFO: 这里不同角色 raft node，tick() 函数不一样：对于 Leader node, tick() 就是 tickHeartbeat；对于 Follower/PreCandidate/Candidate，tick() 就是 tickElection
func (rawNode *RawNode) Tick() {
	rawNode.raft.tick()
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

// HasReady called when RawNode user need to check if any Ready pending.
// Checking logic in this method should be consistent with Ready.containsUpdates().
func (rawNode *RawNode) HasReady() bool {
	r := rawNode.raft
	if !r.softState().equal(rawNode.prevSoftSt) {
		return true
	}
	if hardSt := r.hardState(); !IsEmptyHardState(hardSt) && !isHardStateEqual(hardSt, rawNode.prevHardSt) {
		return true
	}
	if r.raftLog.hasPendingSnapshot() {
		return true
	}
	if len(r.msgs) > 0 || len(r.raftLog.unstableEntries()) > 0 || r.raftLog.hasNextEnts() {
		return true
	}
	if len(r.readStates) != 0 {
		return true
	}

	return false
}

func (rawNode *RawNode) readyWithoutAccept() Ready {
	return rawNode.raft.newReady(rawNode.prevSoftSt, rawNode.prevHardSt)
}

// acceptReady is called when the consumer of the RawNode has decided to go
// ahead and handle a Ready. Nothing must alter the state of the RawNode between
// this call and the prior call to Ready().
func (rawNode *RawNode) acceptReady(ready Ready) {
	if ready.SoftState != nil {
		rawNode.prevSoftSt = ready.SoftState
	}
	if len(ready.ReadStates) != 0 {
		rawNode.raft.readStates = nil
	}

	rawNode.raft.msgs = nil
}

// Advance INFO: 推动用户提交的 []pb.Message 提交到 log 模块中，这里才是最终目标!!!
func (rawNode *RawNode) Advance(ready Ready) {
	if !IsEmptyHardState(ready.HardState) {
		rawNode.prevHardSt = ready.HardState
	}

	rawNode.raft.advance(ready)
}
