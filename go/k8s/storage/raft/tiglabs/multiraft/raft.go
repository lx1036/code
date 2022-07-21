package multiraft

import (
	"fmt"
	"k8s.io/klog/v2"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"k8s-lx1036/k8s/storage/raft/proto"
	"k8s-lx1036/k8s/storage/raft/util"
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

// handle user's get log entries request
type entryRequest struct {
	future     *Future
	index      uint64
	maxSize    uint64
	onlyCommit bool
}

type softState struct {
	leader uint64
	term   uint64
}

type peerState struct {
	sync.RWMutex
	peers map[uint64]proto.Peer
}

func (s *peerState) get() []uint64 {
	s.RLock()
	defer s.RUnlock()

	var nodes []uint64
	for n := range s.peers {
		nodes = append(nodes, n)
	}

	return nodes
}

func (s *peerState) replace(peers []proto.Peer) {
	s.Lock()
	defer s.Unlock()

	s.peers = nil
	s.peers = make(map[uint64]proto.Peer)
	for _, p := range peers {
		s.peers[p.ID] = p
	}
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
	receiveChan       chan *proto.Message
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

// INFO:
func newRaft(config *Config, raftConfig *RaftConfig) (*Raft, error) {
	defer util.HandleCrash()

	if err := raftConfig.validate(); err != nil {
		return nil, err
	}

	raftFSM, err := NewRaftFsm(config, raftConfig)
	if err != nil {
		return nil, err
	}

	mStatus := &monitorStatus{
		conErrCount:    0,
		replicasErrCnt: make(map[uint64]uint8),
	}
	raft := &Raft{
		raftFsm:       raftFSM,
		config:        config,
		raftConfig:    raftConfig,
		mStatus:       mStatus,
		pending:       make(map[uint64]*Future),
		snapping:      make(map[uint64]*snapshotStatus),
		receiveChan:   make(chan *proto.Message, config.ReqBufferSize),
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
	raft.curApplied.Set(raftFSM.raftLog.Applied)
	raft.peerState.replace(raftConfig.Peers)

	go raft.runApply()
	go raft.run()
	go raft.monitor()

	return raft, nil
}

func (r *Raft) runApply() {

	for {
		select {
		case <-r.stopc:
			return
		case apply := <-r.applyc:
			// TODO:
		}
	}

}

func (r *Raft) monitor() {
	statusTicker := time.NewTicker(5 * time.Second)
	leaderTicker := time.NewTicker(1 * time.Minute)
	for {
		select {
		case <-r.stopc:
			statusTicker.Stop()
			return

		case <-statusTicker.C:
			// TODO:

		case <-leaderTicker.C:
			// TODO:
		}
	}
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
			if r.raftFsm.leader != r.config.NodeID {
				propose.future.respond(nil, ErrNotLeader)
				pool.returnProposal(propose)
				break
			}

			msg := proto.NewMessage()
			msg.Type = proto.LocalMsgProp
			msg.From = r.config.NodeID
			starti := r.raftFsm.raftLog.LastIndex() + 1
			r.pending[starti] = propose.future
			msg.Entries = append(msg.Entries, &proto.Entry{Term: r.raftFsm.term, Index: starti, Type: propose.cmdType, Data: propose.data})
			pool.returnProposal(propose)

			flag := false
			for i := 1; i < 64; i++ {
				starti = starti + 1
				select {
				case pr := <-r.propc:
					r.pending[starti] = pr.future
					msg.Entries = append(msg.Entries, &proto.Entry{Term: r.raftFsm.term, Index: starti, Type: pr.cmdType, Data: pr.data})
					pool.returnProposal(pr)
				default:
					flag = true
				}
				if flag {
					break
				}
			}
			r.raftFsm.Step(msg)
		}
	}
}

func (r *Raft) receiveMessage(message *proto.Message) {
	if r.restoringSnapshot.Get() {
		return
	}

	select {
	case <-r.stopc:
	case r.receiveChan <- message:
	default:
		klog.Warningf(fmt.Sprintf("[receiveMessage]raft[%v] discard message(%v)", r.raftConfig.ID, message.ToString()))
		return
	}
}
