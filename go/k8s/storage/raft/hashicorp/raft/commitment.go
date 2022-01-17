package raft

import (
	"sort"
	"sync"
)

// Commitment is used to advance the leader's commit index. The leader and
// replication goroutines report in newly written entries with Match(), and
// this notifies on commitCh when the commit index has advanced.
type commitment struct {
	// protects matchIndexes and commitIndex
	sync.Mutex
	// notified when commitIndex increases
	commitCh chan struct{}
	// voter ID to log index: the server stores up through this log entry
	matchIndexes map[ServerID]uint64
	// a quorum stores up through this log entry. monotonically increases.
	commitIndex uint64
	// the first index of this leader's term: this needs to be replicated to a
	// majority of the cluster before this leader may mark anything committed
	// (per Raft's commitment rule)
	startIndex uint64
}

// newCommitment returns a commitment struct that notifies the provided
// channel when log entries have been committed. A new commitment struct is
// created each time this server becomes leader for a particular term.
// 'configuration' is the servers in the cluster.
// 'startIndex' is the first index created in this term (see
// its description above).
func newCommitment(commitCh chan struct{}, configuration Configuration, startIndex uint64) *commitment {
	matchIndexes := make(map[ServerID]uint64)
	for _, server := range configuration.Servers {
		if server.Suffrage == Voter {
			matchIndexes[server.ID] = 0
		}
	}
	return &commitment{
		commitCh:     commitCh,
		matchIndexes: matchIndexes,
		commitIndex:  0,
		startIndex:   startIndex,
	}
}

// Match is called once a server completes writing entries to disk: either the
// leader has written the new entry or a follower has replied to an
// AppendEntries RPC. The given server's disk agrees with this server's log up
// through the given index.
func (c *commitment) match(id ServerID, matchIndex uint64) {
	c.Lock()
	defer c.Unlock()

	if prev, ok := c.matchIndexes[id]; ok && matchIndex > prev {
		c.matchIndexes[id] = matchIndex
		c.recalculate()
	}
}

// Internal helper to calculate new commitIndex from matchIndexes.
// Must be called with lock held.
func (c *commitment) recalculate() {
	if len(c.matchIndexes) == 0 {
		return
	}

	matched := make([]uint64, 0, len(c.matchIndexes))
	for _, idx := range c.matchIndexes {
		matched = append(matched, idx)
	}
	sort.Sort(uint64Slice(matched))
	quorumMatchIndex := matched[(len(matched)-1)/2] // TODO: 查找论文计算 commit index
	if quorumMatchIndex > c.commitIndex && quorumMatchIndex >= c.startIndex {
		c.commitIndex = quorumMatchIndex
		select {
		case c.commitCh <- struct{}{}: // @see runLeader()::<-r.leaderState.commitCh
		default:
		}
	}
}

type uint64Slice []uint64

func (p uint64Slice) Len() int           { return len(p) }
func (p uint64Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p uint64Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
