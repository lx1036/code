package raft

import (
	"fmt"
	"k8s.io/klog/v2"
	"sync"
	"sync/atomic"
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

// replicate is a long running routine that replicates log entries to a single follower.
func (r *Raft) replicate(follower *followerReplication) {
	// Start an async heartbeat routing
	stopHeartbeat := make(chan struct{})
	defer close(stopHeartbeat)
	go r.heartbeat(follower, stopHeartbeat)

	//RPC:
	shouldStop := false
	for !shouldStop {
		select {
		case <-follower.triggerCh:
			lastLogIdx, _ := r.getLastLog()
			shouldStop = r.replicateTo(follower, lastLogIdx)
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
func (r *Raft) replicateTo(replication *followerReplication, lastIndex uint64) (shouldStop bool) {
	var peer Server
	var req AppendEntriesRequest
	var resp AppendEntriesResponse

Start:
	// Prevent an excessive retry rate on errors
	if replication.failures > 0 {
		select {
		case <-time.After(backoff(failureWait, replication.failures, maxFailureScale)):
		case <-r.shutdownCh:
		}
	}

	replication.peerLock.RLock()
	peer = replication.peer
	replication.peerLock.RUnlock()

	// Setup the request
	if err := r.setupAppendEntries(replication, &req, atomic.LoadUint64(&replication.nextIndex), lastIndex); err == ErrLogNotFound {
		goto SendSnap
	} else if err != nil {
		return
	}

	// Make the RPC call
	if err := r.transport.AppendEntries(peer.ID, peer.Address, &req, &resp); err != nil {
		klog.Errorf(fmt.Sprintf("failed to appendEntries to peer:%s/%s err:%v", peer.ID, peer.Address, err))
		replication.failures++
		return
	}

	// Check for a newer term, stop running
	if resp.Term > req.Term {
		r.handleStaleTerm(replication)
		return true
	}

	// Update the last contact
	replication.setLastContact()

	// Update s based on success
	if resp.Success {
		// Update our replication state
		updateLastAppended(replication, &req)

		// Clear any failures, allow pipelining
		replication.failures = 0
		replication.allowPipeline = true
	} else {
		atomic.StoreUint64(&replication.nextIndex, max(min(replication.nextIndex-1, resp.LastLog+1), 1))
		if resp.NoRetryBackoff {
			replication.failures = 0
		} else {
			replication.failures++
		}
		klog.Warningf(fmt.Sprintf("appendEntries rejected, sending older logs to peer:%s/%s nextIndex:%d",
			peer.ID, peer.Address, atomic.LoadUint64(&replication.nextIndex)))
	}

CheckMore:
	// Poll the stop channel here in case we are looping and have been asked
	// to stop, or have stepped down as leader. Even for the best effort case
	// where we are asked to replicate to a given index and then shutdown,
	// it's better to not loop in here to send lots of entries to a straggler
	// that's leaving the cluster anyways.
	select {
	case <-replication.stopCh:
		return true
	default:
	}

	// Check if there are more logs to replicate
	if atomic.LoadUint64(&replication.nextIndex) <= lastIndex {
		goto Start
	}
	return

	// SEND_SNAP is used when we fail to get a log, usually because the follower
	// is too far behind, and we must ship a snapshot down instead
SendSnap:
	if stop, err := r.sendLatestSnapshot(replication); stop {
		return true
	} else if err != nil {
		klog.Errorf(fmt.Sprintf("failed to send snapshot to peer:%s/%s err:%v", peer.ID, peer.Address, err))
		return
	}

	// Check if there is more to replicate
	goto CheckMore
}

// setupAppendEntries is used to setup an append entries request.
func (r *Raft) setupAppendEntries(s *followerReplication, req *AppendEntriesRequest, nextIndex, lastIndex uint64) error {
	req.Term = s.currentTerm
	req.Leader = r.transport.EncodePeer(r.localID, r.localAddr)
	req.LeaderCommitIndex = r.getCommitIndex()
	if err := r.setPreviousLog(req, nextIndex); err != nil {
		return err
	}
	if err := r.setNewLogs(req, nextIndex, lastIndex); err != nil {
		return err
	}
	return nil
}

// setPreviousLog is used to setup the PrevLogEntry and PrevLogTerm for an
// AppendEntriesRequest given the next index to replicate.
func (r *Raft) setPreviousLog(req *AppendEntriesRequest, nextIndex uint64) error {
	// Guard for the first index, since there is no 0 log entry
	// Guard against the previous index being a snapshot as well
	lastSnapIdx, lastSnapTerm := r.getLastSnapshot()
	if nextIndex == 1 {
		req.PrevLogEntry = 0
		req.PrevLogTerm = 0
	} else if (nextIndex - 1) == lastSnapIdx {
		req.PrevLogEntry = lastSnapIdx
		req.PrevLogTerm = lastSnapTerm
	} else {
		var l Log
		if err := r.logs.GetLog(nextIndex-1, &l); err != nil {
			klog.Errorf(fmt.Sprintf("failed to get log index:%d err:%v", nextIndex-1, err))
			return err
		}

		// Set the previous index and term (0 if nextIndex is 1)
		req.PrevLogEntry = l.Index
		req.PrevLogTerm = l.Term
	}
	return nil
}

// setNewLogs is used to setup the logs which should be appended for a request.
func (r *Raft) setNewLogs(req *AppendEntriesRequest, nextIndex, lastIndex uint64) error {
	// Append up to MaxAppendEntries or up to the lastIndex. we need to use a
	// consistent value for maxAppendEntries in the lines below in case it ever
	// becomes reloadable.
	maxAppendEntries := r.config().MaxAppendEntries
	req.Entries = make([]*Log, 0, maxAppendEntries)
	maxIndex := min(nextIndex+uint64(maxAppendEntries)-1, lastIndex)
	for i := nextIndex; i <= maxIndex; i++ {
		oldLog := new(Log)
		if err := r.logs.GetLog(i, oldLog); err != nil {
			klog.Errorf(fmt.Sprintf("failed to get log index:%d err:%v", i, err))
			return err
		}
		req.Entries = append(req.Entries, oldLog)
	}
	return nil
}

// handleStaleTerm is used when a follower indicates that we have a stale term.
func (r *Raft) handleStaleTerm(replication *followerReplication) {
	klog.Errorf(fmt.Sprintf("peer:%s/%s has newer term, stopping replication", replication.peer.ID, replication.peer.Address))
	replication.notifyAll(false) // No longer leader
	select {
	case replication.stepDown <- struct{}{}:
	default:
	}
}

// sendLatestSnapshot is used to send the latest snapshot we have
// down to our follower.
func (r *Raft) sendLatestSnapshot(replication *followerReplication) (bool, error) {
	return true, nil
}

// updateLastAppended is used to update follower replication state after a
// successful AppendEntries RPC.
// TODO: This isn't used during InstallSnapshot, but the code there is similar.
func updateLastAppended(s *followerReplication, req *AppendEntriesRequest) {
	// Mark any inflight logs as committed
	if logs := req.Entries; len(logs) > 0 {
		last := logs[len(logs)-1]
		atomic.StoreUint64(&s.nextIndex, last.Index+1)
		s.commitment.match(s.peer.ID, last.Index)
	}

	// Notify still leader
	s.notifyAll(true)
}
