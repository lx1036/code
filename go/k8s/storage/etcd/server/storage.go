package server

import (
	"k8s-lx1036/k8s/storage/etcd/storage/wal"

	"go.etcd.io/etcd/raft/v3/raftpb"
	"go.etcd.io/etcd/server/v3/etcdserver/api/snap"
)

// Storage INFO: WAL 存储，其实就是一个文件
type Storage interface {
	Save(st raftpb.HardState, ents []raftpb.Entry) error

	SaveSnap(snap raftpb.Snapshot) error

	Close() error

	// Release releases the locked wal files older than the provided snapshot.
	Release(snap raftpb.Snapshot) error

	// Sync WAL
	Sync() error
}

type storage struct {
	*wal.WAL

	*snap.Snapshotter
}

func NewStorage(w *wal.WAL, s *snap.Snapshotter) Storage {
	return &storage{
		WAL:         w,
		Snapshotter: s,
	}
}

func (s *storage) Release(snap raftpb.Snapshot) error {
	panic("implement me")
}

func (s *storage) Sync() error {
	panic("implement me")
}
