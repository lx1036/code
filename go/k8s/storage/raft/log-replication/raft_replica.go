package log_replication

import (
	"time"

	"k8s-lx1036/k8s/storage/sunfs/pkg/raft/proto"
)

// replication represents a follower’s progress of replicate in the view of the leader.
// Leader maintains progresses of all followers, and sends entries to the follower based on its progress.
type Replica struct {
	inflight
	peer                                proto.Peer
	state                               replicaState
	paused, active, pending             bool
	match, next, committed, pendingSnap uint64

	lastActive time.Time
}

func NewReplica(peer proto.Peer, maxInflight int) *Replica {
	repl := &Replica{
		peer:       peer,
		state:      replicaStateProbe,
		lastActive: time.Now(),
	}
	if maxInflight > 0 {
		repl.inflight.size = maxInflight
		repl.inflight.buffer = make([]uint64, maxInflight)
	}

	return repl
}

// inflight is the replication sliding window,avoid overflowing that sending buffer.
type inflight struct {
	start  int
	count  int
	size   int
	buffer []uint64
}
