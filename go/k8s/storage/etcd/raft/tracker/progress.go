package tracker

// StateType is the state of a tracked follower.
type StateType uint64

const (
	// StateProbe indicates a follower whose last index isn't known. Such a
	// follower is "probed" (i.e. an append sent periodically) to narrow down
	// its last index. In the ideal (and common) case, only one round of probing
	// is necessary as the follower will react with a hint. Followers that are
	// probed over extended periods of time are often offline.
	StateProbe StateType = iota
	// StateReplicate is the state steady in which a follower eagerly receives
	// log entries to append to its log.
	StateReplicate
	// StateSnapshot indicates a follower that needs log entries not available
	// from the leader's Raft log. Such a follower needs a full snapshot to
	// return to StateReplicate.
	StateSnapshot
)

var stateTypeStrings = []string{
	"StateProbe",
	"StateReplicate",
	"StateSnapshot",
}

func (stateType StateType) String() string {
	return stateTypeStrings[stateType]
}

// Progress represents a follower’s progress in the view of the leader. Leader
// maintains progresses of all followers, and sends entries to the follower
// based on its progress.
type Progress struct {
	Match, Next uint64

	State StateType

	PendingSnapshot uint64

	RecentActive bool

	ProbeSent bool

	Inflights *Inflights

	IsLearner bool
}

// MaybeUpdate is called when an MsgAppResp arrives from the follower, with the
// index acked by it. The method returns false if the given n index comes from
// an outdated message. Otherwise it updates the progress and returns true.
func (pr *Progress) MaybeUpdate(n uint64) bool {
	var updated bool
	if pr.Match < n {
		pr.Match = n
		updated = true
		pr.ProbeAcked()
	}
	pr.Next = max(pr.Next, n+1)
	return updated
}

// ProbeAcked is called when this peer has accepted an append. It resets
// ProbeSent to signal that additional append messages should be sent without
// further delay.
func (pr *Progress) ProbeAcked() {
	pr.ProbeSent = false
}

// IsPaused returns whether sending log entries to this node has been throttled.
func (pr *Progress) IsPaused() bool {
	switch pr.State {
	case StateProbe:
		return pr.ProbeSent
	case StateReplicate:
		return pr.Inflights.Full()
	case StateSnapshot:
		return true
	default:
		panic("unexpected state")
	}
}

// OptimisticUpdate signals that appends all the way up to and including index n
// are in-flight. As a result, Next is increased to n+1.
func (pr *Progress) OptimisticUpdate(n uint64) {
	pr.Next = n + 1
}

// BecomeReplicate transitions into StateReplicate, resetting Next to Match+1.
func (pr *Progress) BecomeReplicate() {
	pr.ResetState(StateReplicate)
	pr.Next = pr.Match + 1
}

// ResetState moves the Progress into the specified State, resetting ProbeSent,
// PendingSnapshot, and Inflights.
func (pr *Progress) ResetState(state StateType) {
	pr.ProbeSent = false
	pr.PendingSnapshot = 0
	pr.State = state
	pr.Inflights.reset()
}

type ProgressMap map[uint64]*Progress

// Inflights limits the number of MsgApp (represented by the largest index contained within) sent to followers but not yet acknowledged by them.
type Inflights struct {
	// the starting index in the buffer
	start int
	// number of inflights in the buffer
	count int

	// the size of the buffer
	size int

	// buffer contains the index of the last entry
	// inside one message.
	buffer []uint64
}

func NewInflights(size int) *Inflights {
	return &Inflights{
		size: size,
	}
}

// Full returns true if no more messages can be sent at the moment.
func (in *Inflights) Full() bool {
	return in.count == in.size
}

// Add notifies the Inflights that a new message with the given index is being
// dispatched. Full() must be called prior to Add() to verify that there is room
// for one more message, and consecutive calls to add Add() must provide a
// monotonic sequence of indexes.
func (in *Inflights) Add(inflight uint64) {
	if in.Full() {
		panic("cannot add into a Full inflights")
	}
	next := in.start + in.count
	size := in.size
	if next >= size {
		next -= size
	}
	if next >= len(in.buffer) {
		in.grow()
	}
	in.buffer[next] = inflight
	in.count++
}

// INFO: 扩容 buffer cap
func (in *Inflights) grow() {
	newSize := len(in.buffer) * 2
	if newSize == 0 {
		newSize = 1
	} else if newSize > in.size {
		newSize = in.size
	}
	newBuffer := make([]uint64, newSize)
	copy(newBuffer, in.buffer)
	in.buffer = newBuffer
}

// reset frees all inflights
func (in *Inflights) reset() {
	in.count = 0
	in.start = 0
}

func max(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}
