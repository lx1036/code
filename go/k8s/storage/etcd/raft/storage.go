package raft

import (
	"errors"
	"fmt"
	"sync"

	pb "go.etcd.io/etcd/raft/v3/raftpb"
	"k8s.io/klog/v2"
)

// ErrCompacted is returned by Storage.Entries/Compact when a requested
// index is unavailable because it predates the last snapshot.
var ErrCompacted = errors.New("requested index is unavailable due to compaction")

// ErrSnapOutOfDate is returned by Storage.CreateSnapshot when a requested
// index is older than the existing snapshot.
var ErrSnapOutOfDate = errors.New("requested index is older than the existing snapshot")

// ErrUnavailable is returned by Storage interface when the requested log entries
// are unavailable.
var ErrUnavailable = errors.New("requested entry at index is unavailable")

// ErrSnapshotTemporarilyUnavailable is returned by the Storage interface when the required
// snapshot is temporarily unavailable.
var ErrSnapshotTemporarilyUnavailable = errors.New("snapshot is temporarily unavailable")

// Storage INFO: 使用该 Storage 接口，mvcc statemachine 可以 retrieve log entries from WAL
type Storage interface {

	// InitialState INFO: 返回 saved HardState and ConfState
	InitialState() (pb.HardState, pb.ConfState, error)

	// Entries returns a slice of log entries in the range [lo,hi).
	// MaxSize limits the total size of the log entries returned, but
	// Entries returns at least one entry if any.
	Entries(lo, hi, maxSize uint64) ([]pb.Entry, error)

	// Term returns the term of entry i, which must be in the range
	// [FirstIndex()-1, LastIndex()]. The term of the entry before
	// FirstIndex is retained for matching purposes even though the
	// rest of that entry may not be available.
	Term(i uint64) (uint64, error)

	// LastIndex returns the index of the last entry in the log.
	LastIndex() uint64

	// FirstIndex returns the index of the first log entry that is
	// possibly available via Entries (older entries have been incorporated
	// into the latest Snapshot; if storage only contains the dummy entry the
	// first log entry is not available).
	FirstIndex() uint64

	// Snapshot returns the most recent snapshot.
	// If snapshot is temporarily unavailable, it should return ErrSnapshotTemporarilyUnavailable,
	// so raft state machine could know that Storage needs some time to prepare
	// snapshot and call Snapshot later.
	Snapshot() (pb.Snapshot, error)
}

type MemoryStorage struct {
	sync.Mutex

	// TODO: 为何关注这个两个 State
	hardState pb.HardState
	snapshot  pb.Snapshot

	// ents[i] has raft log position i+snapshot.Metadata.Index
	entries []pb.Entry
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		entries: make([]pb.Entry, 1),
	}
}

func (storage *MemoryStorage) InitialState() (pb.HardState, pb.ConfState, error) {
	return storage.hardState, storage.snapshot.Metadata.ConfState, nil
}

func (storage *MemoryStorage) LastIndex() uint64 {
	storage.Lock()
	defer storage.Unlock()
	return storage.lastIndex()
}

func (storage *MemoryStorage) lastIndex() uint64 {
	return storage.entries[0].Index + uint64(len(storage.entries)) - 1 // [0].Index + len -1
}

func (storage *MemoryStorage) FirstIndex() uint64 {
	storage.Lock()
	defer storage.Unlock()
	return storage.firstIndex()
}

func (storage *MemoryStorage) firstIndex() uint64 {
	// TODO: 为何加 1
	return storage.entries[0].Index + 1
}

// Append INFO: entries[0].Index > ms.entries[0].Index, 有点类似于 log append() 追加写逻辑
func (storage *MemoryStorage) Append(entries []pb.Entry) {
	if len(entries) == 0 {
		return
	}

	storage.Lock()
	defer storage.Unlock()

	first := storage.firstIndex()
	last := entries[0].Index + uint64(len(entries)) - 1
	if last < first {
		return
	}
	// truncate compacted entries
	if first > entries[0].Index {
		entries = entries[first-entries[0].Index:]
	}

	offset := entries[0].Index - storage.entries[0].Index
	switch {
	case uint64(len(storage.entries)) > offset:
		// INFO: 这里类似于 [3-5].append([4-6])=[3, 4-6], 属于交叉情况
		storage.entries = append([]pb.Entry{}, storage.entries[:offset]...)
		storage.entries = append(storage.entries, entries...)
	case uint64(len(storage.entries)) == offset:
		storage.entries = append(storage.entries, entries...)
	default:
		klog.Fatalf(fmt.Sprintf("[Append]missing log entry [last: %d, append at: %d]", storage.lastIndex(), entries[0].Index))
	}

	return
}

func (storage *MemoryStorage) Entries(lo, hi, maxSize uint64) ([]pb.Entry, error) {
	panic("implement me")
}

func (storage *MemoryStorage) Term(i uint64) (uint64, error) {
	panic("implement me")
}

func (storage *MemoryStorage) Snapshot() (pb.Snapshot, error) {
	panic("implement me")
}
