package raft

import (
	"sync/atomic"
	"time"
)

// LeaderObservation is used for the data when leadership changes.
type LeaderObservation struct {
	Leader ServerAddress
}

// PeerObservation is sent to observers when peers change.
type PeerObservation struct {
	Removed bool
	Peer    Server
}

// FailedHeartbeatObservation is sent when a node fails to heartbeat with the leader
type FailedHeartbeatObservation struct {
	PeerID      ServerID
	LastContact time.Time
}

// ResumedHeartbeatObservation is sent when a node resumes to heartbeat with the leader following failures
type ResumedHeartbeatObservation struct {
	PeerID ServerID
}

// observe sends an observation to every observer.
func (r *Raft) observe(o interface{}) {

}

// Observation is sent along the given channel to observers when an event occurs.
type Observation struct {
	// Raft holds the Raft instance generating the observation.
	Raft *Raft
	// Data holds observation-specific data. Possible types are
	// *RequestVoteRequest
	// RaftState
	// PeerObservation
	// LeaderObservation
	Data interface{}
}

// FilterFn is a function that can be registered in order to filter observations.
// The function reports whether the observation should be included - if
// it returns false, the observation will be filtered out.
type FilterFn func(o *Observation) bool

// Observer describes what to do with a given observation.
type Observer struct {
	// numObserved and numDropped are performance counters for this observer.
	// 64 bit types must be 64 bit aligned to use with atomic operations on
	// 32 bit platforms, so keep them at the top of the struct.
	numObserved uint64
	numDropped  uint64

	// channel receives observations.
	channel chan Observation

	// blocking, if true, will cause Raft to block when sending an observation
	// to this observer. This should generally be set to false.
	blocking bool

	// filter will be called to determine if an observation should be sent to
	// the channel.
	filter FilterFn

	// id is the ID of this observer in the Raft map.
	id uint64
}

// nextObserverId is used to provide a unique ID for each observer to aid in
// deregistration.
var nextObserverID uint64

// NewObserver creates a new observer that can be registered
// to make observations on a Raft instance. Observations
// will be sent on the given channel if they satisfy the
// given filter.
//
// If blocking is true, the observer will block when it can't
// send on the channel, otherwise it may discard events.
func NewObserver(channel chan Observation, blocking bool, filter FilterFn) *Observer {
	return &Observer{
		channel:  channel,
		blocking: blocking,
		filter:   filter,
		id:       atomic.AddUint64(&nextObserverID, 1),
	}
}
