package raft

import (
	"fmt"
	"sync"
	"time"

	"k8s.io/klog/v2"
)

type Raft struct {
	raftState

	// lastContact is the last time we had contact from the
	// leader node. This can be used to gauge staleness.
	lastContact     time.Time
	lastContactLock sync.RWMutex
}

func NewRaft() (*Raft, error) {

	r := &Raft{}

	go r.run()

}

func (r *Raft) run() {
	for {
		// Check if we are doing a shutdown
		select {
		case <-r.shutdownCh:
			// Clear the leader to prevent forwarding
			r.setLeader("")
			return
		default:
		}

		switch r.getState() {
		case Follower:
			r.runFollower()
		case Candidate:
			//r.runCandidate()
		case Leader:
			//r.runLeader()
		}
	}
}

func (r *Raft) runFollower() {
	klog.Infof(fmt.Sprintf("%s/%s entering follower state in cluster for leader:%s", r.localID, r.localAddr, r.Leader()))

	heartbeatTimer := randomTimeout(r.config().HeartbeatTimeout)
	for r.getState() == Follower {
		select {
		case <-heartbeatTimer: // 每 [1s, 2s] 一次心跳检查是否有心跳
			// Restart the heartbeat timer
			heartbeatTimeout := r.config().HeartbeatTimeout
			heartbeatTimer = randomTimeout(heartbeatTimeout) // [1s, 2s]

			// INFO: 提高safety: 这里使用 lastContact，如果是正常的 log replicate，也会修改 lastContact
			//  本来担心网络抖动会导致几次心跳没成功，会发起 leader election，但是每 HeartbeatTimeout / 10 leader 发起一次心跳，如果
			//  10次心跳都没成功，就必然 ElectionTimeout，则可以发起选举, @see https://github.com/hashicorp/raft/blob/v1.3.3/replication.go#L389-L394
			lastContact := r.LastContact()
			if time.Now().Sub(lastContact) < heartbeatTimeout {
				continue
			}

			// TODO: Heartbeat failed! Transition to the candidate state

		}
	}

}

// setLastContact is used to set the last contact time to now
func (r *Raft) setLastContact() {
	r.lastContactLock.Lock()
	r.lastContact = time.Now()
	r.lastContactLock.Unlock()
}

// LastContact returns the time of last contact by a leader.
// This only makes sense if we are currently a follower.
func (r *Raft) LastContact() time.Time {
	r.lastContactLock.RLock()
	last := r.lastContact
	r.lastContactLock.RUnlock()
	return last
}
