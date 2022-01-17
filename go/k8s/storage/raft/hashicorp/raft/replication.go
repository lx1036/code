package raft

import (
	"fmt"
	"k8s.io/klog/v2"
	"sync"
	"time"
)

const (
	maxFailureScale = 12
	failureWait     = 10 * time.Millisecond
)

// followerReplication is in charge of sending snapshots and log entries from
// this leader during this particular term to a remote follower.
type followerReplication struct {
	// currentTerm and nextIndex must be kept at the top of the struct so
	// they're 64 bit aligned which is a requirement for atomic ops on 32 bit
	// platforms.

	// currentTerm is the term of this leader, to be included in AppendEntries
	// requests.
	currentTerm uint64

	// nextIndex is the index of the next log entry to send to the follower,
	// which may fall past the end of the log.
	nextIndex uint64

	// peer contains the network address and ID of the remote follower.
	peer Server
	// peerLock protects 'peer'
	peerLock sync.RWMutex

	// commitment tracks the entries acknowledged by followers so that the
	// leader's commit index can advance. It is updated on successful
	// AppendEntries responses.
	commitment *commitment

	// stopCh is notified/closed when this leader steps down or the follower is
	// removed from the cluster. In the follower removed case, it carries a log
	// index; replication should be attempted with a best effort up through that
	// index, before exiting.
	stopCh chan uint64

	// triggerCh is notified every time new entries are appended to the log.
	triggerCh chan struct{}
	// triggerDeferErrorCh is used to provide a backchannel. By sending a
	// deferErr, the sender can be notifed when the replication is done.
	triggerDeferErrorCh chan *deferError

	// lastContact is updated to the current time whenever any response is
	// received from the follower (successful or not). This is used to check
	// whether the leader should step down (Raft.checkLeaderLease()).
	lastContact time.Time
	// lastContactLock protects 'lastContact'.
	lastContactLock sync.RWMutex

	// failures counts the number of failed RPCs since the last success, which is
	// used to apply backoff.
	failures uint64

	// notifyCh is notified to send out a heartbeat, which is used to check that
	// this server is still leader.
	notifyCh chan struct{}
	// notify is a map of futures to be resolved upon receipt of an
	// acknowledgement, then cleared from this map.
	notify map[*verifyFuture]struct{}
	// notifyLock protects 'notify'.
	notifyLock sync.Mutex

	// stepDown is used to indicate to the leader that we
	// should step down based on information from a follower.
	stepDown chan struct{}

	// allowPipeline is used to determine when to pipeline the AppendEntries RPCs.
	// It is private to this replication goroutine.
	allowPipeline bool
}

// setLastContact sets the last contact to the current time.
func (s *followerReplication) setLastContact() {
	s.lastContactLock.Lock()
	s.lastContact = time.Now()
	s.lastContactLock.Unlock()
}

// LastContact returns the time of last contact.
func (s *followerReplication) LastContact() time.Time {
	s.lastContactLock.RLock()
	lastContact := s.lastContact
	s.lastContactLock.RUnlock()
	return lastContact
}

// notifyAll is used to notify all the waiting verify futures
// if the follower believes we are still the leader.
func (s *followerReplication) notifyAll(leader bool) {
	// Clear the waiting notifies minimizing lock time
	s.notifyLock.Lock()
	n := s.notify
	s.notify = make(map[*verifyFuture]struct{})
	s.notifyLock.Unlock()

	// Submit our votes
	for v := range n {
		v.vote(leader)
	}
}

// replicate is a long running routine that replicates log entries to a single
// follower.
func (r *Raft) replicate(s *followerReplication) {
	// Start an async heartbeat routing
	stopHeartbeat := make(chan struct{})
	defer close(stopHeartbeat)
	go r.heartbeat(s, stopHeartbeat)

	//RPC:
	shouldStop := false
	for !shouldStop {
		select {
		case <-s.triggerCh:
			lastLogIdx, _ := r.getLastLog()
			shouldStop = r.replicateTo(s, lastLogIdx)
		}
	}

}

// heartbeat is used to periodically invoke AppendEntries on a peer
// to ensure they don't time out. This is done async of replicate(),
// since that routine could potentially be blocked on disk IO.
func (r *Raft) heartbeat(replication *followerReplication, stopCh chan struct{}) {
	var failures uint64
	req := AppendEntriesRequest{
		Term:   replication.currentTerm,
		Leader: r.transport.EncodePeer(r.localID, r.localAddr),
	}
	var resp AppendEntriesResponse
	for {
		// Wait for the next heartbeat interval or forced notify
		select {
		case <-replication.notifyCh:
		case <-randomTimeout(r.config().HeartbeatTimeout / 10): // [100ms, 200ms]
		case <-stopCh:
			return
		}

		replication.peerLock.RLock()
		peer := replication.peer
		replication.peerLock.RUnlock()

		if err := r.transport.AppendEntries(peer.ID, peer.Address, &req, &resp); err != nil {
			klog.Errorf(fmt.Sprintf("failed to heartbeat from %s/%s to %s/%s err:%v",
				r.localID, r.localAddr, peer.ID, peer.Address, err))
			r.observe(FailedHeartbeatObservation{PeerID: peer.ID, LastContact: replication.LastContact()})

			// backoff
			failures++
			select {
			case <-time.After(backoff(failureWait, failures, maxFailureScale)):
			case <-stopCh:
				return
			}
		} else {
			if failures > 0 {
				r.observe(ResumedHeartbeatObservation{PeerID: peer.ID})
			}

			replication.setLastContact()
			failures = 0
			replication.notifyAll(resp.Success)
		}
	}
}

// replicateTo is a helper to replicate(), used to replicate the logs up to a
// given last index.
// If the follower log is behind, we take care to bring them up to date.
func (r *Raft) replicateTo(s *followerReplication, lastIndex uint64) (shouldStop bool) {
	return false
}
