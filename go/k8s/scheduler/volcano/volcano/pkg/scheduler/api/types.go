package api

type NodePhase int

const (
	// Ready means the node is ready for scheduling
	Ready NodePhase = 1 << iota
	// NotReady means the node is not ready for scheduling
	NotReady
)

func (np NodePhase) String() string {
	switch np {
	case Ready:
		return "Ready"
	case NotReady:
		return "NotReady"
	}

	return "Unknown"
}

// CompareFn is the func declaration used by sort or priority queue.
type CompareFn func(interface{}, interface{}) int

// EvictableFn is the func declaration used to evict tasks.
type EvictableFn func(*TaskInfo, []*TaskInfo) ([]*TaskInfo, int)

// ValidateFn is the func declaration used to check object's status.
type ValidateFn func(interface{}) bool

// ValidateExFn is the func declaration used to validate the result.
type ValidateExFn func(interface{}) *ValidateResult

// ValidateResult is struct to which can used to determine the result
type ValidateResult struct {
	Pass    bool
	Reason  string
	Message string
}

// LessFn is the func declaration used by sort or priority queue.
type LessFn func(interface{}, interface{}) bool

// VoteFn is the func declaration used to check object's complicated status.
type VoteFn func(interface{}) int

// JobEnqueuedFn is the func declaration used to call after job enqueued.
type JobEnqueuedFn func(interface{})
