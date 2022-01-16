package raft

import (
	"fmt"

	pb "go.etcd.io/etcd/raft/v3/raftpb"
	"k8s.io/klog/v2"
)

type raftLog struct {
	// INFO: 这个 storage 是用的 raft MemoryStorage，存在内存里，这个很重要!!!
	//  @see EtcdServer 模块中的 RaftNode.storage，这个是 wal 持久化文件
	// storage contains all stable entries since the last snapshot.
	storage Storage

	// INFO: 表示还未 commit 的 log entry，committed log entry 会保存到 raftLog.storage，也就是 stable
	unstable unstable

	// INFO: position 记录已经提交的 log entry
	committed uint64

	// INFO: position 记录 state machine 状态机已经 applied应用的 log entry，并且 applied <= committed
	applied uint64

	// maxNextEntsSize is the maximum number aggregate byte size of the messages
	// returned from calls to nextEnts.
	maxNextEntsSize uint64 // 1MB
}

func newLog(storage Storage) *raftLog {
	return newRaftLogWithSize(storage, noLimit)
}

func newRaftLogWithSize(storage Storage, maxNextEntsSize uint64) *raftLog {
	if storage == nil {
		klog.Fatalf("storage must not be nil")
	}

	log := &raftLog{
		storage:         storage,
		maxNextEntsSize: maxNextEntsSize,
	}

	firstIndex := storage.FirstIndex()
	lastIndex := storage.LastIndex()
	log.unstable.offset = lastIndex + 1
	// Initialize our committed and applied pointers to the time of the last compaction.
	log.committed = firstIndex - 1
	log.applied = firstIndex - 1

	return log
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
	if maxIndex > log.committed && log.zeroTermOnErrCompacted(log.term(maxIndex)) == term {
		log.commitTo(maxIndex)
		return true
	}
	return false
}

func (log *raftLog) commitTo(toCommittedIndex uint64) {
	// never decrease commit
	if log.committed < toCommittedIndex {
		if log.lastIndex() < toCommittedIndex {
			klog.Errorf(fmt.Sprintf("tocommit(%d) is out of range [lastIndex(%d)]. Was the raft log corrupted, truncated, or lost?",
				toCommittedIndex, log.lastIndex()))
		}

		log.committed = toCommittedIndex
	}
}

func (log *raftLog) appliedTo(index uint64) {
	if index == 0 {
		return
	}
	if log.committed < index || index < log.applied {
		klog.Fatalf(fmt.Sprintf("[raft log appliedTo]applied(%d) is out of range [prevApplied(%d), committed(%d)]", index, log.applied, log.committed))
	}

	log.applied = index
}

func (log *raftLog) stableTo(index, term uint64) {
	log.unstable.stableTo(index, term)
}

func (log *raftLog) stableSnapTo(i uint64) {
	log.unstable.stableSnapTo(i)
}

// hasPendingSnapshot returns if there is pending snapshot waiting for applying.
func (log *raftLog) hasPendingSnapshot() bool {
	return log.unstable.snapshot != nil && !IsEmptySnap(*log.unstable.snapshot)
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

func (log *raftLog) matchTerm(index, term uint64) bool {
	t, err := log.term(index)
	if err != nil {
		return false
	}

	return t == term
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

// 如果是 ErrCompacted, 返回 term=0
func (log *raftLog) zeroTermOnErrCompacted(t uint64, err error) uint64 {
	if err == nil {
		return t
	}
	if err == ErrCompacted {
		return 0
	}

	klog.Fatalf(fmt.Sprintf("unexpected error (%v)", err))
	return 0
}

func (log *raftLog) restore(snapshot pb.Snapshot) {
	klog.Infof(fmt.Sprintf("log [%s] starts to restore snapshot [index: %d, term: %d]", log, snapshot.Metadata.Index, snapshot.Metadata.Term))
	log.committed = snapshot.Metadata.Index
	log.unstable.restore(snapshot)
}

func (log *raftLog) String() string {
	return fmt.Sprintf("committed=%d, applied=%d, unstable.offset=%d, len(unstable.Entries)=%d", log.committed, log.applied, log.unstable.offset, len(log.unstable.entries))
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

func (u *unstable) stableTo(i, t uint64) {
	term, ok := u.maybeTerm(i)
	if !ok {
		return
	}

	// if i < offset, term is matched with the snapshot
	// only update the unstable entries if term is matched with
	// an unstable entry.
	if term == t && i >= u.offset {
		u.entries = u.entries[i+1-u.offset:]
		u.offset = i + 1
		u.shrinkEntriesArray()
	}
}

func (u *unstable) stableSnapTo(i uint64) {
	if u.snapshot != nil && u.snapshot.Metadata.Index == i {
		u.snapshot = nil
	}
}

// INFO: 其实就是压缩下 u.entries 空间避免浪费，减少内存，u.entries 值内容不动
//  Avoid holding unneeded memory in unstable log's entries array
func (u *unstable) shrinkEntriesArray() {
	const lenMultiple = 2
	if len(u.entries) == 0 {
		u.entries = nil
	} else if len(u.entries)*lenMultiple < cap(u.entries) {
		newEntries := make([]pb.Entry, len(u.entries))
		copy(newEntries, u.entries)
		u.entries = newEntries
	}
}

// INFO: unstable log 从 snapshot.Index + 1 开始
func (u *unstable) maybeFirstIndex() (uint64, bool) {
	if u.snapshot != nil {
		return u.snapshot.Metadata.Index + 1, true
	}
	return 0, false
}

// INFO: 返回 index 对应的 term
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

func (u *unstable) restore(snapshot pb.Snapshot) {
	u.offset = snapshot.Metadata.Index + 1
	u.entries = nil
	u.snapshot = &snapshot
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
