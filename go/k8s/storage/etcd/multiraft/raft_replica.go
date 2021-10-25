package multiraft

import (
	"fmt"
	"k8s.io/klog/v2"
	"time"

	"k8s-lx1036/k8s/storage/raft/proto"
)

// inflight is the replication sliding window,avoid overflowing that sending buffer.
type inflight struct {
	start  int
	count  int
	size   int
	buffer []uint64
}

func (in *inflight) add(index uint64) {
	if in.full() {
		klog.Fatalf(fmt.Sprintf("inflight.add cannot add into a full inflights."))
	}

	next := in.start + in.count
	if next >= in.size {
		next = next - in.size
	}
	in.buffer[next] = index
	in.count = in.count + 1
}

func (in *inflight) full() bool {
	return in.count == in.size
}

// Replica replication represents a follower’s progress of replicate in the view of the leader.
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

func (r *Replica) maybeUpdate(index, commit uint64) bool {
	updated := false
	if r.committed < commit {
		r.committed = commit
	}
	if r.match < index {
		r.match = index
		updated = true
		r.resume()
	}
	next := index + 1
	if r.next < next {
		r.next = next
	}

	return updated
}

func (r *Replica) resume() {
	r.paused = false
}

func (r *Replica) update(index uint64) {
	r.next = index + 1
}

func (r *Replica) pause() {
	r.paused = true
}
