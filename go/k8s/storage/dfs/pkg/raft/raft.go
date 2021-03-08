package raft

import (
	"sync"
	"sync/atomic"
	"unsafe"

	"k8s-lx1036/k8s/storage/dfs/pkg/raft/proto"
	"k8s-lx1036/k8s/storage/dfs/pkg/raft/util"
)

type proposal struct {
	cmdType proto.EntryType
	future  *Future
	data    []byte
}

type apply struct {
	term        uint64
	index       uint64
	future      *Future
	command     interface{}
	readIndexes []*Future
}

type softState struct {
	leader uint64
	term   uint64
}

type peerState struct {
	peers map[uint64]proto.Peer
	mu    sync.RWMutex
}

func (s *peerState) get() (nodes []uint64) {
	s.mu.RLock()
	for n := range s.peers {
		nodes = append(nodes, n)
	}
	s.mu.RUnlock()
	return
}

type monitorStatus struct {
	conErrCount    uint8
	replicasErrCnt map[uint64]uint8
}

type Raft struct {
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

func (r *Raft) isLeader() bool {
	leader, _ := r.leaderTerm()
	return leader == r.config.NodeID
}

func (r *Raft) leaderTerm() (leader, term uint64) {
	st := (*softState)(atomic.LoadPointer(&r.curSoftSt))
	if st == nil {
		return NoLeader, 0
	}
	return st.leader, st.term
}

func (r *Raft) getPeers() (peers []uint64) {
	return r.peerState.get()
}

func (r *Raft) tick() {
	/*if r.restoringSnapshot.Get() {
		return
	}*/

	select {
	case <-r.stopc:
	case r.tickc <- struct{}{}:
	default:
		return
	}
}

func (r *Raft) stop() {
	select {
	case <-r.done:
		return
	default:
		r.doStop()
	}

	<-r.done
}

func (r *Raft) doStop() {
	r.mu.Lock()
	defer r.mu.Unlock()

	select {
	case <-r.stopc:
		return
	default:
		close(r.stopc)
		r.restoringSnapshot.Set(false)
	}
}

// Read
func (r *Raft) readIndex(future *Future) {
	if !r.isLeader() {
		future.respond(nil, ErrNotLeader)
		return
	}

	select {
	case <-r.stopc:
		future.respond(nil, ErrStopped)
	case r.readIndexC <- future: // 写给 readIndexC channel
	}
}

// Write 提交个 propose 到 propose channel
func (r *Raft) propose(cmd []byte, future *Future) {
	if !r.isLeader() {
		future.respond(nil, ErrNotLeader)
		return
	}

	pr := pool.getProposal()
	pr.cmdType = proto.EntryNormal
	pr.data = cmd
	pr.future = future

	select {
	case <-r.stopc:
		future.respond(nil, ErrStopped)
	case r.propc <- pr: // 向 propose channel 里提交 cmd，会在 run() 函数里读 propose channel 数据
	}
}

func (r *Raft) run() {

	for {
		select {
		case propose := <-r.propc: // 读取 cmd 然后处理数据

		}
	}
}

func newRaft(config *Config, raftConfig *RaftConfig) (*Raft, error) {
	defer util.HandleCrash()

	if err := raftConfig.validate(); err != nil {
		return nil, err
	}

	r, err := newRaftFsm(config, raftConfig)
	if err != nil {
		return nil, err
	}

	mStatus := &monitorStatus{
		conErrCount:    0,
		replicasErrCnt: make(map[uint64]uint8),
	}
	raft := &Raft{
		raftFsm:       r,
		config:        config,
		raftConfig:    raftConfig,
		mStatus:       mStatus,
		pending:       make(map[uint64]*Future),
		snapping:      make(map[uint64]*snapshotStatus),
		recvc:         make(chan *proto.Message, config.ReqBufferSize),
		applyc:        make(chan *apply, config.AppBufferSize),
		propc:         make(chan *proposal, 256),
		snapRecvc:     make(chan *snapshotRequest, 1),
		truncatec:     make(chan uint64, 1),
		readIndexC:    make(chan *Future, 256),
		statusc:       make(chan chan *Status, 1),
		entryRequestC: make(chan *entryRequest, 16),
		tickc:         make(chan struct{}, 64),
		readyc:        make(chan struct{}, 1),
		electc:        make(chan struct{}, 1),
		stopc:         make(chan struct{}),
		done:          make(chan struct{}),
	}
	raft.curApplied.Set(r.raftLog.applied)
	raft.peerState.replace(raftConfig.Peers)

	util.RunWorker(raft.runApply, raft.handlePanic)
	util.RunWorker(raft.run, raft.handlePanic)
	util.RunWorker(raft.monitor, raft.handlePanic)
	return raft, nil

}
