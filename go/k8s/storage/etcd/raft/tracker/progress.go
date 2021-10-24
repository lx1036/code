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
