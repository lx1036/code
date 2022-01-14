package raft

import (
	"sync"
	"sync/atomic"
)

type RaftState uint32

const (
	// Follower is the initial state of a Raft node.
	Follower RaftState = iota

	// Candidate is one of the valid states of a Raft node.
	Candidate

	// Leader is one of the valid states of a Raft node.
	Leader

	// Shutdown is the terminal state of a Raft node.
	Shutdown
)

func (raftState RaftState) String() string {
	switch raftState {
	case Follower:
		return "Follower"
	case Candidate:
		return "Candidate"
	case Leader:
		return "Leader"
	case Shutdown:
		return "Shutdown"
	default:
		return "Unknown"
	}
}

type raftState struct {
	// currentTerm commitIndex, lastApplied, must be kept at the top of
	// the struct so they're 64 bit aligned which is a requirement for
	// atomic ops on 32 bit platforms.

	// The current state
	state RaftState

	// The current term, cache of StableStore
	currentTerm uint64

	// protects 4 next fields
	lastLock sync.Mutex
	// Cache the latest snapshot index/term
	lastSnapshotIndex uint64
	lastSnapshotTerm  uint64
	// Cache the latest log from LogStore
	lastLogIndex uint64
	lastLogTerm  uint64
}

func (r *raftState) getState() RaftState {
	return r.state
}

func (r *raftState) setState(state RaftState) {
	r.state = state
}

func (r *raftState) setCurrentTerm(term uint64) {
	atomic.StoreUint64(&r.currentTerm, term)
}

func (r *raftState) setLastLog(index, term uint64) {
	r.lastLock.Lock()
	r.lastLogIndex = index
	r.lastLogTerm = term
	r.lastLock.Unlock()
}

func (r *raftState) getLastSnapshot() (index, term uint64) {
	r.lastLock.Lock()
	index = r.lastSnapshotIndex
	term = r.lastSnapshotTerm
	r.lastLock.Unlock()
	return
}
