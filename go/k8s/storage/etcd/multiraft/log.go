package multiraft

import (
	"k8s-lx1036/k8s/storage/etcd/multiraft/storage"
	"k8s-lx1036/k8s/storage/raft/proto"
)

// raftLog 主要用来持久化 operation log
type RaftLog struct {
	Unstable           unstable
	Storage            storage.Storage
	Committed, Applied uint64
}

func NewRaftLog(storage storage.Storage) (*RaftLog, error) {
	log := &RaftLog{
		Storage: storage,
	}

	firstIndex, err := storage.FirstIndex()
	if err != nil {
		return nil, err
	}
	lastIndex, err := storage.LastIndex()
	if err != nil {
		return nil, err
	}

	log.Unstable.offset = lastIndex + 1
	log.Unstable.entries = make([]*proto.Entry, 0, 256)
	log.Committed = firstIndex - 1
	log.Applied = firstIndex - 1

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
