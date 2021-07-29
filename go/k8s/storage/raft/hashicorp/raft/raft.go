package raft

import (
	"fmt"
	"k8s.io/klog/v2"
)

// Raft implements a Raft node.
type Raft struct {
	raftState
}

func (r *Raft) run() {
	for {
		switch r.getState() {
		case Follower:
			r.runFollower()
		}
	}
}

func (r *Raft) runFollower() {
	for r.getState() == Follower {
		select {}
	}
}

func (r *Raft) setState(state RaftState) {
	oldState := r.raftState.getState()
	r.raftState.setState(state)
	if oldState != state {
		klog.Infof(fmt.Sprintf("swich raft state from %s to %s", oldState, state))
	}
}

func NewRaft(config *Config) (*Raft, error) {
	// Create Raft struct.
	r := &Raft{}

	// Initialize as a follower.
	r.setState(Follower)

	go r.run()

	return r, nil
}
