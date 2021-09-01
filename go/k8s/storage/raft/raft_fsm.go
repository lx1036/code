package raft

import (
	"fmt"
	"math/rand"
	"strings"

	"k8s-lx1036/k8s/storage/raft/proto"
	"k8s-lx1036/k8s/storage/raft/raftlog"

	"k8s.io/klog/v2"
)

// finite state machine
type raftFsm struct {
	id               uint64
	term             uint64
	vote             uint64
	leader           uint64
	electionElapsed  int
	heartbeatElapsed int
	// randElectionTick is a random number between[electiontimetick, 2 * electiontimetick - 1].
	// It gets reset when raft changes its state to follower or candidate.
	randElectionTick int
	// New configuration is ignored if there exists unapplied configuration.
	pendingConf bool
	state       FsmState
	sm          StateMachine
	config      *Config
	raftLog     *raftlog.RaftLog
	rand        *rand.Rand
	votes       map[uint64]bool
	acks        map[uint64]bool
	replicas    map[uint64]*Replica
	readOnly    *readOnly
	msgs        []*proto.Message
	step        stepFunc
	tick        func()
}

func NewRaftFsm(config *Config, raftConfig *RaftConfig) (*raftFsm, error) {
	raftLog, err := raftlog.NewRaftLog(raftConfig.Storage)
	if err != nil {
		return nil, err
	}
	hs, err := raftConfig.Storage.InitialState()
	if err != nil {
		return nil, err
	}

	r := &raftFsm{
		id:       raftConfig.ID,
		sm:       raftConfig.StateMachine,
		config:   config,
		leader:   NoLeader,
		raftLog:  raftLog,
		replicas: make(map[uint64]*Replica),
		readOnly: newReadOnly(raftConfig.ID, config.ReadOnlyOption),
	}
	r.rand = rand.New(rand.NewSource(int64(config.NodeID + r.id)))
	for _, p := range raftConfig.Peers {
		r.replicas[p.ID] = NewReplica(p, 0)
	}
	if !hs.IsEmpty() {
		if raftConfig.Applied > r.raftLog.LastIndex() {
			raftConfig.Applied = r.raftLog.LastIndex()
		}
		if hs.Commit > r.raftLog.LastIndex() {
			hs.Commit = r.raftLog.LastIndex()
		}
		if err := r.loadState(hs); err != nil {
			return nil, err
		}
	}

	klog.Info("newRaft[%v] [commit: %d, applied: %d, lastindex: %d]", r.id, raftLog.Committed, raftConfig.Applied, raftLog.LastIndex())

	if raftConfig.Applied > 0 {
		lasti := raftLog.LastIndex()
		if lasti == 0 {
			// If there is application data but no raft log, then restore to initial state.
			raftLog.Committed = 0
			raftConfig.Applied = 0
		} else if lasti < raftConfig.Applied {
			// If lastIndex<appliedIndex, then the log as the standard.
			raftLog.Committed = lasti
			raftConfig.Applied = lasti
		} else if raftLog.Committed < raftConfig.Applied {
			raftLog.Committed = raftConfig.Applied
		}
		raftLog.AppliedTo(raftConfig.Applied)
	}

	// recover committed
	if err := r.recoverCommit(); err != nil {
		return nil, err
	}
	if raftConfig.Leader == config.NodeID {
		if raftConfig.Term != 0 && r.term <= raftConfig.Term {
			r.term = raftConfig.Term
			r.state = stateLeader
			r.becomeLeader()
			r.bcastAppend()
		} else {
			r.becomeFollower(r.term, NoLeader)
		}
	} else {
		if raftConfig.Leader == NoLeader {
			r.becomeFollower(r.term, NoLeader)
		} else {
			r.becomeFollower(raftConfig.Term, raftConfig.Leader)
		}
	}

	peerStrs := make([]string, 0)
	for _, p := range r.peers() {
		peerStrs = append(peerStrs, fmt.Sprintf("%v", p.String()))
	}
	klog.Infof("newRaft[%v] [peers: [%s], term: %d, commit: %d, applied: %d, lastindex: %d, lastterm: %d]",
		r.id, strings.Join(peerStrs, ","), r.term, r.raftLog.Committed, r.raftLog.Applied, r.raftLog.LastIndex(), r.raftLog.LastTerm())

	go r.doRandomSeed()

	return r, nil
}
