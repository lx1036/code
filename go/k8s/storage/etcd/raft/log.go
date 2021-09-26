package raft

import (
	"log"
)

type raftLog struct {
	// storage contains all stable entries since the last snapshot.
	storage Storage

	// unstable contains all unstable entries and snapshot.
	// they will be saved into storage.
	//unstable unstable

	// committed is the highest log position that is known to be in
	// stable storage on a quorum of nodes.
	committed uint64

	// applied is the highest log position that the application has
	// been instructed to apply to its state machine.
	// Invariant: applied <= committed
	applied uint64

	// maxNextEntsSize is the maximum number aggregate byte size of the messages
	// returned from calls to nextEnts.
	maxNextEntsSize uint64
}

// newLogWithSize returns a log using the given storage and max
// message size.
func newLogWithSize(storage Storage, maxNextEntsSize uint64) *raftLog {
	if storage == nil {
		log.Panic("storage must not be nil")
	}

	rLog := &raftLog{
		storage:         storage,
		maxNextEntsSize: maxNextEntsSize,
	}

	firstIndex := storage.FirstIndex()
	//lastIndex := storage.LastIndex()
	//rLog.unstable.offset = lastIndex + 1
	// Initialize our committed and applied pointers to the time of the last compaction.
	rLog.committed = firstIndex - 1
	rLog.applied = firstIndex - 1

	return rLog
}

func (log *raftLog) lastIndex() uint64 {
	/*if i, ok := l.unstable.maybeLastIndex(); ok {
		return i
	}*/

	return log.storage.LastIndex()
}

// INFO:
func (log *raftLog) hasPendingSnapshot() bool {
	//return log.unstable.snapshot != nil && !IsEmptySnap(*log.unstable.snapshot)
	return false
}
