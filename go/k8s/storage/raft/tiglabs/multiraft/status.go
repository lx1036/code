package multiraft

import (
	"time"
)

// ReplicaStatus  replica status
type ReplicaStatus struct {
	Match       uint64 // 复制进度
	Commit      uint64 // commmit位置
	Next        uint64
	State       string
	Snapshoting bool
	Paused      bool
	Active      bool
	LastActive  time.Time
	Inflight    int
}

// Status raft status
type Status struct {
	ID                uint64
	NodeID            uint64
	Leader            uint64
	Term              uint64
	Index             uint64
	Commit            uint64
	Applied           uint64
	Vote              uint64
	PendQueue         int
	RecvQueue         int
	AppQueue          int
	Stopped           bool
	RestoringSnapshot bool
	State             string // leader、follower、candidate
	Replicas          map[uint64]*ReplicaStatus
}
