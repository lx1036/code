package raft

// LeaderObservation is used for the data when leadership changes.
type LeaderObservation struct {
	Leader ServerAddress
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
