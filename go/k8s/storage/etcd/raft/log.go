package raft

import (
	"fmt"
	"log"

	pb "go.etcd.io/etcd/raft/v3/raftpb"
	"k8s.io/klog/v2"
)

type raftLog struct {
	// storage contains all stable entries since the last snapshot.
	storage Storage

	// unstable contains all unstable entries and snapshot.
	// they will be saved into storage.
	unstable unstable

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

func (log *raftLog) append(ents ...pb.Entry) uint64 {
	if len(ents) == 0 {
		return log.lastIndex()
	}

	if after := ents[0].Index - 1; after < log.committed {
		klog.Fatalf(fmt.Sprintf("after(%d) is out of range [committed(%d)]", after, log.committed))
	}

	log.unstable.truncateAndAppend(ents)
	return log.lastIndex()
}

func (log *raftLog) maybeCommit(maxIndex, term uint64) bool {
	//if maxIndex > log.committed && log.zeroTermOnErrCompacted(log.term(maxIndex)) == term {
	if maxIndex > log.committed {
		log.commitTo(maxIndex)
		return true
	}
	return false
}

func (log *raftLog) commitTo(tocommit uint64) {
	// never decrease commit
	if log.committed < tocommit {
		if log.lastIndex() < tocommit {
			klog.Errorf(fmt.Sprintf("tocommit(%d) is out of range [lastIndex(%d)]. Was the raft log corrupted, truncated, or lost?",
				tocommit, log.lastIndex()))
		}

		log.committed = tocommit
	}
}

// INFO:
func (log *raftLog) hasPendingSnapshot() bool {
	//return log.unstable.snapshot != nil && !IsEmptySnap(*log.unstable.snapshot)
	return false
}

func (log *raftLog) term(i uint64) (uint64, error) {
	// the valid term range is [index of dummy entry, last index]
	dummyIndex := log.firstIndex() - 1
	if i < dummyIndex || i > log.lastIndex() {
		// TODO: return an error instead?
		return 0, nil
	}

	if t, ok := log.unstable.maybeTerm(i); ok {
		return t, nil
	}

	t, err := log.storage.Term(i)
	if err == nil {
		return t, nil
	}
	if err == ErrCompacted || err == ErrUnavailable {
		return 0, err
	}
	panic(err)
}

// unstable.entries[i] has raft log position i+unstable.offset.
// Note that unstable.offset may be less than the highest log
// position in storage; this means that the next write to storage
// might need to truncate the log before persisting unstable.entries.
type unstable struct {
	// the incoming unstable snapshot, if any.
	snapshot *pb.Snapshot
	// all entries that have not yet been written to storage.
	entries []pb.Entry
	offset  uint64
}
