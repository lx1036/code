package raft

import (
	"github.com/tiglabs/raft/proto"
	"github.com/tiglabs/raft/util"
	"sync"
	"unsafe"
)

type raft struct {
	raftFsm           *raftFsm
	config            *Config
	raftConfig        *RaftConfig
	restoringSnapshot util.AtomicBool
	curApplied        util.AtomicUInt64
	curSoftSt         unsafe.Pointer
	prevSoftSt        softState
	prevHardSt        proto.HardState
	peerState         peerState
	pending           map[uint64]*Future
	snapping          map[uint64]*snapshotStatus
	mStatus           *monitorStatus
	propc             chan *proposal
	applyc            chan *apply
	recvc             chan *proto.Message
	snapRecvc         chan *snapshotRequest
	truncatec         chan uint64
	readIndexC        chan *Future
	statusc           chan chan *Status
	entryRequestC     chan *entryRequest
	readyc            chan struct{}
	tickc             chan struct{}
	electc            chan struct{}
	stopc             chan struct{}
	done              chan struct{}
	mu                sync.Mutex
}
