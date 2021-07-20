package raftlog

import (
	"errors"
	"fmt"
	"math"

	"k8s-lx1036/k8s/storage/raft/proto"
	"k8s-lx1036/k8s/storage/raft/storage"
	"k8s-lx1036/k8s/storage/raft/util"

	"k8s.io/klog/v2"
)

// INFO: raft log
//  Raft协议详解-Log Replication: https://zhuanlan.zhihu.com/p/29730357

const noLimit = math.MaxUint64

var (
	ErrCompacted = errors.New("requested index is unavailable due to compaction.")
)

// raftLog is responsible for the operation of the log.
// raftLog 主要用来持久化 operation log
type RaftLog struct {
	unstable           unstable
	storage            storage.Storage
	committed, applied uint64
}

// INFO: 获取 raftlog 的 last index
func (log *RaftLog) LastIndex() uint64 {
	if i, ok := log.unstable.maybeLastIndex(); ok {
		return i
	}
	i, err := log.storage.LastIndex()
	if err != nil {
		errMsg := fmt.Sprintf("[raftLog->lastIndex]get lastIndex from storage err:[%v]", err)
		klog.Errorf(errMsg)
		panic(fmt.Errorf("occurred application logic panic error: %s", errMsg))
	}
	return i
}

func (log *RaftLog) append(ents ...*proto.Entry) uint64 {
	if len(ents) == 0 {
		return log.LastIndex()
	}

	if after := ents[0].Index - 1; after < log.committed {
		errMsg := fmt.Sprintf("[raftLog->append]after(%d) is out of range [committed(%d)]", after, log.committed)
		klog.Error(errMsg)
		panic(fmt.Errorf(errMsg))
	}

	// unstable 存储还未提交到 Storage 的 entry
	log.unstable.truncateAndAppend(ents)

	return log.LastIndex()
}

func (log *RaftLog) entries(i uint64, maxsize uint64) ([]*proto.Entry, error) {
	if i > log.LastIndex() {
		return nil, nil
	}

	return log.slice(i, log.LastIndex()+1, maxsize)
}

func (log *RaftLog) firstIndex() uint64 {
	index, err := log.storage.FirstIndex()
	if err != nil {
		errMsg := fmt.Sprintf("[raftLog->firstIndex]get firstindex from storage err:[%v].", err)
		klog.Error(errMsg)
		panic(fmt.Errorf(errMsg))
	}
	return index
}

// log.firstIndex <= lo <= hi <= log.firstIndex + len(log.entries)
func (log *RaftLog) mustCheckOutOfBounds(lo, hi uint64) error {
	if lo > hi {
		errMsg := fmt.Sprintf("[raftLog->mustCheckOutOfBounds]invalid slice %d > %d", lo, hi)
		klog.Error(errMsg)
		panic(fmt.Errorf(errMsg))
	}
	fi := log.firstIndex()
	if lo < fi {
		return ErrCompacted
	}
	lastIndex := log.LastIndex()
	if hi > lastIndex+1 {
		errMsg := fmt.Sprintf("[raftLog->mustCheckOutOfBounds]slice[%d,%d) out of bound [%d,%d]", lo, hi, fi, lastIndex)
		klog.Error(errMsg)
		panic(fmt.Errorf(errMsg))
	}

	return nil
}

func (log *RaftLog) slice(lo, hi uint64, maxSize uint64) ([]*proto.Entry, error) {
	if lo == hi {
		return nil, nil
	}
	err := log.mustCheckOutOfBounds(lo, hi)
	if err != nil {
		return nil, err
	}

	var ents []*proto.Entry
	if lo < log.unstable.offset {
		storedhi := util.Min(hi, log.unstable.offset)
		storedEnts, cmp, err := log.storage.Entries(lo, storedhi, maxSize)
		if cmp {
			return nil, ErrCompacted
		} else if err != nil {
			errMsg := fmt.Sprintf("[raftLog->slice]get entries[%d:%d) from storage err:[%v].", lo, storedhi, err)
			klog.Error(errMsg)
			panic(fmt.Errorf(errMsg))
		}
		// check if ents has reached the size limitation
		if uint64(len(storedEnts)) < storedhi-lo {
			return storedEnts, nil
		}
		ents = storedEnts
	}
	if hi > log.unstable.offset {
		unstable := log.unstable.slice(util.Max(lo, log.unstable.offset), hi)
		if len(ents) > 0 {
			ents = append([]*proto.Entry{}, ents...)
			ents = append(ents, unstable...)
		} else {
			ents = unstable
		}
	}

	if maxSize == noLimit {
		return ents, nil
	}

	return limitSize(ents, maxSize), nil
}

func newRaftLog(storage storage.Storage) (*RaftLog, error) {
	log := &RaftLog{
		storage: storage,
	}

	firstIndex, err := storage.FirstIndex()
	if err != nil {
		return nil, err
	}
	lastIndex, err := storage.LastIndex()
	if err != nil {
		return nil, err
	}

	log.unstable.offset = lastIndex + 1
	log.unstable.entries = make([]*proto.Entry, 0, 256)
	log.committed = firstIndex - 1
	log.applied = firstIndex - 1

	return log, nil
}

// unstable temporary deposit the unpersistent log entries.It has log position i+unstable.offset.
// unstable can support group commit.
// Note that unstable.offset may be less than the highest log position in storage;
// this means that the next write to storage might need to truncate the log before persisting unstable.entries.
type unstable struct {
	offset uint64
	// all entries that have not yet been written to storage.
	entries []*proto.Entry
}

// maybeLastIndex returns the last index if it has at least one unstable entry.
func (u *unstable) maybeLastIndex() (uint64, bool) {
	if l := len(u.entries); l != 0 {
		return u.offset + uint64(l) - 1, true
	}
	return 0, false
}

// TODO 这里的逻辑后续继续看下
func (u *unstable) truncateAndAppend(ents []*proto.Entry) {
	after := ents[0].Index
	switch {
	case after == u.offset+uint64(len(u.entries)):
		// after is the next index in the u.entries, directly append
		u.entries = append(u.entries, ents...)

	case after <= u.offset:
		// The log is being truncated to before our current offset portion, so set the offset and replace the entries
		u.offset = after
		u.entries = ents // TODO ???

	default:
		// truncate to after and copy to u.entries then append
		// TODO ???
		u.entries = append(u.entries[0:0], u.slice(u.offset, after)...)
		u.entries = append(u.entries, ents...)
	}
}

func (u *unstable) slice(lo uint64, hi uint64) []*proto.Entry {
	u.mustCheckOutOfBounds(lo, hi)
	return u.entries[lo-u.offset : hi-u.offset]
}

// u.offset <= lo <= hi <= u.offset+len(u.offset)
func (u *unstable) mustCheckOutOfBounds(lo, hi uint64) {
	if lo > hi {
		errMsg := fmt.Sprintf("unstable.slice[%d,%d) is invalid.", lo, hi)
		klog.Error(errMsg)
		panic(fmt.Errorf(errMsg))
	}
	upper := u.offset + uint64(len(u.entries))
	if lo < u.offset || hi > upper {
		errMsg := fmt.Sprintf("unstable.slice[%d,%d) out of bound [%d,%d].", lo, hi, u.offset, upper)
		klog.Error(errMsg)
		panic(fmt.Errorf(errMsg))
	}
}

func limitSize(ents []*proto.Entry, maxSize uint64) []*proto.Entry {
	if len(ents) == 0 || maxSize == noLimit {
		return ents
	}

	size := ents[0].Size()
	limit := 1
	for l := len(ents); limit < l; limit++ {
		size += ents[limit].Size()
		if size > maxSize {
			break
		}
	}
	return ents[:limit]
}
