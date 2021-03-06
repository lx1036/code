package storage

import (
	"errors"
	"fmt"
	"runtime"

	"github.com/tiglabs/raft"
	"github.com/tiglabs/raft/proto"
	"github.com/tiglabs/raft/util"
	"k8s.io/klog/v2"
)

type fsm interface {
	AppliedIndex(id uint64) uint64
}

// This storage is circular storage in memory and truncate when over capacity,
// but keep it a high capacity.
type MemoryStorage struct {
	fsm fsm
	id  uint64
	// the threshold of truncate
	capacity uint64
	// the index of last truncate
	truncIndex uint64
	truncTerm  uint64
	// the starting offset in the ents
	start uint64
	// the actual log in the ents
	count uint64
	// the total size of the ents
	size uint64
	// ents[i] has raft log position i+snapshot.Metadata.Index
	ents      []*proto.Entry
	hardState proto.HardState
}

func (ms *MemoryStorage) InitialState() (proto.HardState, error) {
	return ms.hardState, nil
}

func (ms *MemoryStorage) FirstIndex() (uint64, error) {
	return ms.truncIndex + 1, nil
}

func (ms *MemoryStorage) LastIndex() (uint64, error) {
	return ms.lastIndex(), nil
}

func (ms *MemoryStorage) lastIndex() uint64 {
	return ms.truncIndex + ms.count
}

func (ms *MemoryStorage) Term(index uint64) (term uint64, isCompact bool, err error) {
	switch {
	case index < ms.truncIndex:
		return 0, true, nil
	case index == ms.truncIndex:
		return ms.truncTerm, false, nil
	default:
		return ms.ents[ms.locatePosition(index)].Term, false, nil
	}
}

func (ms *MemoryStorage) locatePosition(index uint64) uint64 {
	position := ms.start + (index - ms.truncIndex - 1)
	if position >= ms.size {
		position = position - ms.size
	}
	return position
}

func (ms *MemoryStorage) Entries(lo, hi uint64, maxSize uint64) (entries []*proto.Entry, isCompact bool, err error) {
	if lo <= ms.truncIndex {
		return nil, true, nil
	}
	if hi > ms.lastIndex()+1 {
		return nil, false, fmt.Errorf("[MemoryStorage->Entries]entries's hi(%d) is out of bound lastindex(%d)", hi, ms.lastIndex())
	}
	// only contains dummy entries.
	if ms.count == 0 {
		return nil, false, errors.New("requested entry at index is unavailable")
	}

	count := hi - lo
	if count <= 0 {
		return []*proto.Entry{}, false, nil
	}
	retEnts := make([]*proto.Entry, count)
	pos := ms.locatePosition(lo)
	retEnts[0] = ms.ents[pos]
	size := ms.ents[pos].Size()
	limit := uint64(1)
	for ; limit < count; limit++ {
		pos = pos + 1
		if pos >= ms.size {
			pos = pos - ms.size
		}
		size = size + ms.ents[pos].Size()
		if uint64(size) > maxSize {
			break
		}
		retEnts[limit] = ms.ents[pos]
	}
	return retEnts[:limit], false, nil
}

func (ms *MemoryStorage) StoreEntries(entries []*proto.Entry) error {
	if len(entries) == 0 {
		return nil
	}

	appIndex := uint64(0)
	if ms.fsm != nil {
		appIndex = ms.fsm.AppliedIndex(ms.id)
	}
	first := appIndex + 1
	last := entries[0].Index + uint64(len(entries)) - 1
	if last < first {
		// shortcut if there is no new entry.
		return nil
	}
	if first > entries[0].Index {
		// truncate compacted entries
		entries = entries[first-entries[0].Index:]
	}
	offset := entries[0].Index - ms.truncIndex - 1
	if ms.count < offset {
		klog.Errorf("missing log entry [last: %d, append at: %d]", ms.lastIndex(), entries[0].Index)
		return nil
	}

	// resize and truncate compacted ents
	entriesSize := uint64(len(entries))
	maxSize := offset + entriesSize
	minSize := maxSize - (appIndex - ms.truncIndex)
	switch {
	case minSize > ms.capacity:
		// truncate compacted ents
		if ms.truncIndex < appIndex {
			ms.truncateTo(appIndex)
		}
		// grow ents
		if minSize > ms.size {
			ms.resize(ms.capacity+minSize, minSize)
		}

	default:
		// truncate compacted ents
		if maxSize > ms.capacity {
			cmpIdx := util.Min(appIndex, maxSize-ms.capacity+ms.truncIndex)
			if ms.truncIndex < cmpIdx {
				ms.truncateTo(cmpIdx)
			}
		}
		// short ents
		if ms.size > ms.capacity {
			ms.resize(ms.capacity, maxSize)
		}
	}

	// append new entries
	start := ms.locatePosition(entries[0].Index)
	next := start + entriesSize
	if next <= ms.size {
		copy(ms.ents[start:], entries)
		if ms.start <= start {
			ms.count = next - ms.start
		} else {
			ms.count = (ms.size - ms.start) + (next - 0)
		}
	} else {
		count := ms.size - start
		copy(ms.ents[start:], entries[0:count])
		copy(ms.ents[0:], entries[count:])
		ms.count = (ms.size - ms.start) + (entriesSize - count)
	}

	return nil
}

func (ms *MemoryStorage) StoreHardState(st proto.HardState) error {
	ms.hardState = st
	return nil
}

func (ms *MemoryStorage) ApplySnapshot(meta proto.SnapshotMeta) error {
	ms.truncIndex = meta.Index
	ms.truncTerm = meta.Term
	ms.start = 0
	ms.count = 0
	ms.size = ms.capacity
	ms.ents = make([]*proto.Entry, ms.capacity)
	return nil
}

func (ms *MemoryStorage) Truncate(index uint64) error {
	if index == 0 || index <= ms.truncIndex {
		return errors.New("requested index is unavailable due to compaction")
	}
	if index > ms.lastIndex() {
		return fmt.Errorf("compact %d is out of bound lastindex(%d)", index, ms.lastIndex())
	}
	ms.truncateTo(index)
	return nil
}

func (ms *MemoryStorage) truncateTo(index uint64) {
	ms.truncTerm = ms.ents[ms.locatePosition(index)].Term
	ms.start = ms.locatePosition(index + 1)
	ms.count = ms.count - (index - ms.truncIndex)
	ms.truncIndex = index
}

func (ms *MemoryStorage) resize(capacity, needSize uint64) {
	ents := make([]*proto.Entry, capacity)
	count := util.Min(util.Min(capacity, ms.count), needSize)
	next := ms.start + count
	if next <= ms.size {
		copy(ents, ms.ents[ms.start:next])
	} else {
		next = next - ms.size
		copy(ents, ms.ents[ms.start:])
		copy(ents[ms.size-ms.start:], ms.ents[0:next])
	}

	ms.start = 0
	ms.count = count
	ms.size = capacity
	ms.ents = ents
}

func NewMemoryStorage(fsm fsm, id, capacity uint64) *MemoryStorage {
	klog.Infof("Memory Storage capacity is: %v.", capacity)
	return &MemoryStorage{
		fsm:      fsm,
		id:       id,
		capacity: capacity,
		size:     capacity,
		ents:     make([]*proto.Entry, capacity),
	}
}

func NewStorage(raftServer *raft.RaftServer) *MemoryStorage {
	return NewMemoryStorage(raftServer, 1, 8192)
}

func DefaultMemoryStorage() *MemoryStorage {
	return NewMemoryStorage(nil, 0, 4096)
}

func init() {
	numCpu := runtime.NumCPU()
	runtime.GOMAXPROCS(numCpu)
	klog.Infof("[System], Cpu Num = [%d]\r\n", numCpu)
}
