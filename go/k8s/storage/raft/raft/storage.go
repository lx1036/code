package raft

import "sync"

// INFO: 支持读写并发
type Storage struct {
	sync.RWMutex

	wal *Wal
}

func NewStorage(dir string) (RaftLog, error) {
	w, err := NewWal(dir, &Option{
		SegmentNum:  4,
		SegmentSize: 20 * 1024, // 20 KB
		IsSync:      true,
	})
	if err != nil {
		return nil, err
	}

	storage := &Storage{
		wal: w,
	}

	return storage, nil
}
