package raft

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
	// The current state
	state RaftState
}

func (r *raftState) getState() RaftState {
	return r.state
}

func (r *raftState) setState(state RaftState) {
	r.state = state
}
