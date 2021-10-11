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

func newLog(storage Storage) *raftLog {
	return newLogWithSize(storage, noLimit)
}

// newLogWithSize returns a log using the given storage and max
// message size.
func newLogWithSize(storage Storage, maxNextEntsSize uint64) *raftLog {
	if storage == nil {
		log.Panic("storage must not be nil")
	}

	rlog := &raftLog{
		storage:         storage,
		maxNextEntsSize: maxNextEntsSize,
	}

	firstIndex := storage.FirstIndex()
	lastIndex := storage.LastIndex()
	rlog.unstable.offset = lastIndex + 1
	// Initialize our committed and applied pointers to the time of the last compaction.
	rlog.committed = firstIndex - 1
	rlog.applied = firstIndex - 1

	return rlog
}

func (log *raftLog) lastIndex() uint64 {
	if i, ok := log.unstable.maybeLastIndex(); ok {
		return i
	}

	return log.storage.LastIndex()
}

func (log *raftLog) entries(i, maxsize uint64) ([]pb.Entry, error) {
	if i > log.lastIndex() {
		return nil, nil
	}
	return log.slice(i, log.lastIndex()+1, maxsize)
}

// slice returns a slice of log entries from lo through hi-1, inclusive.
func (log *raftLog) slice(lo, hi, maxSize uint64) ([]pb.Entry, error) {
	err := log.mustCheckOutOfBounds(lo, hi)
	if err != nil {
		return nil, err
	}
	if lo == hi {
		return nil, nil
	}
	var ents []pb.Entry
	if lo < log.unstable.offset {
		storedEnts, err := log.storage.Entries(lo, min(hi, log.unstable.offset), maxSize)
		if err == ErrCompacted {
			return nil, err
		} else if err == ErrUnavailable {
			klog.Fatalf("entries[%d:%d) is unavailable from storage", lo, min(hi, log.unstable.offset))
		} else if err != nil {
			panic(err) // TODO(bdarnell)
		}

		// check if ents has reached the size limitation
		if uint64(len(storedEnts)) < min(hi, log.unstable.offset)-lo {
			return storedEnts, nil
		}

		ents = storedEnts
	}
	if hi > log.unstable.offset {
		unstable := log.unstable.slice(max(lo, log.unstable.offset), hi)
		if len(ents) > 0 {
			combined := make([]pb.Entry, len(ents)+len(unstable))
			n := copy(combined, ents)
			copy(combined[n:], unstable)
			ents = combined
		} else {
			ents = unstable
		}
	}
	return limitSize(ents, maxSize), nil
}

// l.firstIndex <= lo <= hi <= l.firstIndex + len(l.entries)
func (log *raftLog) mustCheckOutOfBounds(lo, hi uint64) error {
	if lo > hi {
		klog.Fatalf("invalid slice %d > %d", lo, hi)
	}
	firstIndex := log.firstIndex()
	if lo < firstIndex {
		return ErrCompacted
	}

	length := log.lastIndex() + 1 - firstIndex
	if hi > firstIndex+length {
		klog.Fatalf("slice[%d,%d) out of bound [%d,%d]", lo, hi, firstIndex, log.lastIndex())
	}
	return nil
}

// INFO: 追加写到 storage 中，并返回 lastIndex
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

func (log *raftLog) lastTerm() uint64 {
	t, err := log.term(log.lastIndex())
	if err != nil {
		klog.Fatalf(fmt.Sprintf("unexpected error when getting the last term (%v)", err))
	}

	return t
}

func (log *raftLog) firstIndex() uint64 {
	if i, ok := log.unstable.maybeFirstIndex(); ok {
		return i
	}

	return log.storage.FirstIndex()
}

func (log *raftLog) unstableEntries() []pb.Entry {
	if len(log.unstable.entries) == 0 {
		return nil
	}

	return log.unstable.entries
}

// nextEnts returns all the available entries for execution.
// If applied is smaller than the index of snapshot, it returns all committed
// entries after the index of snapshot.
func (log *raftLog) nextEnts() []pb.Entry {
	off := max(log.applied+1, log.firstIndex())
	if log.committed+1 > off {
		entries, err := log.slice(off, log.committed+1, log.maxNextEntsSize)
		if err != nil {
			klog.Fatalf(fmt.Sprintf("[]unexpected error when getting unapplied entries: %v", err))
		}

		return entries
	}

	return nil
}

// hasNextEnts returns if there is any available entries for execution. This
// is a fast check without heavy raftLog.slice() in raftLog.nextEnts().
func (log *raftLog) hasNextEnts() bool {
	off := max(log.applied+1, log.firstIndex())
	return log.committed+1 > off
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

// maybeFirstIndex returns the index of the first possible entry in entries
// if it has a snapshot.
func (u *unstable) maybeFirstIndex() (uint64, bool) {
	if u.snapshot != nil {
		return u.snapshot.Metadata.Index + 1, true
	}
	return 0, false
}

// maybeTerm returns the term of the entry at index i, if there
// is any.
func (u *unstable) maybeTerm(i uint64) (uint64, bool) {
	if i < u.offset {
		if u.snapshot != nil && u.snapshot.Metadata.Index == i {
			return u.snapshot.Metadata.Term, true
		}
		return 0, false
	}

	last, ok := u.maybeLastIndex()
	if !ok {
		return 0, false
	}
	if i > last {
		return 0, false
	}

	return u.entries[i-u.offset].Term, true
}

// maybeLastIndex returns the last index if it has at least one
// unstable entry or snapshot.
func (u *unstable) maybeLastIndex() (uint64, bool) {
	if l := len(u.entries); l != 0 {
		return u.offset + uint64(l) - 1, true
	}
	if u.snapshot != nil {
		return u.snapshot.Metadata.Index, true
	}
	return 0, false
}

func (u *unstable) truncateAndAppend(ents []pb.Entry) {
	after := ents[0].Index
	switch {
	case after == u.offset+uint64(len(u.entries)):
		// after is the next index in the u.entries
		// directly append
		u.entries = append(u.entries, ents...)
	case after <= u.offset:
		klog.Infof("replace the unstable entries from index %d", after)
		// The log is being truncated to before our current offset
		// portion, so set the offset and replace the entries
		u.offset = after
		u.entries = ents
	default:
		// truncate to after and copy to u.entries
		// then append
		klog.Infof("truncate the unstable entries before index %d", after)
		u.entries = append([]pb.Entry{}, u.slice(u.offset, after)...)
		u.entries = append(u.entries, ents...)
	}
}

func (u *unstable) slice(lo uint64, hi uint64) []pb.Entry {
	u.mustCheckOutOfBounds(lo, hi)
	return u.entries[lo-u.offset : hi-u.offset]
}

// u.offset <= lo <= hi <= u.offset+len(u.entries)
func (u *unstable) mustCheckOutOfBounds(lo, hi uint64) {
	if lo > hi {
		klog.Fatalf("invalid unstable.slice %d > %d", lo, hi)
	}
	upper := u.offset + uint64(len(u.entries))
	if lo < u.offset || hi > upper {
		klog.Fatalf("unstable.slice[%d,%d) out of bound [%d,%d]", lo, hi, u.offset, upper)
	}
}

func min(a, b uint64) uint64 {
	if a > b {
		return b
	}
	return a
}

func max(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

func limitSize(ents []pb.Entry, maxSize uint64) []pb.Entry {
	if len(ents) == 0 {
		return ents
	}
	size := ents[0].Size()
	var limit int
	for limit = 1; limit < len(ents); limit++ {
		size += ents[limit].Size()
		if uint64(size) > maxSize {
			break
		}
	}
	return ents[:limit]
}
