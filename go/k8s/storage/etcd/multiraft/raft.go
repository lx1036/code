package multiraft

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"unsafe"

	"k8s-lx1036/k8s/storage/etcd/multiraft/storage"
	"k8s-lx1036/k8s/storage/raft/proto"
	"k8s-lx1036/k8s/storage/raft/util"

	"k8s.io/klog/v2"
)

// NoLeader is a placeholder nodeID used when there is no leader.
const NoLeader uint64 = 0

// RaftConfig contains the parameters to create a raft.
type RaftConfig struct {
	ID           uint64
	Term         uint64
	Leader       uint64
	Applied      uint64
	Peers        []proto.Peer
	Storage      storage.Storage
	StateMachine StateMachine
}

// validate returns an error if any required elements of the ReplConfig are missing or invalid.
func (c *RaftConfig) validate() error {
	if c.ID == 0 {
		return errors.New("ID is required")
	}
	if len(c.Peers) == 0 {
		return errors.New("Peers is required")
	}
	if c.Storage == nil {
		return errors.New("Storage is required")
	}
	if c.StateMachine == nil {
		return errors.New("StateMachine is required")
	}

	return nil
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

type Raft struct {
	mu sync.Mutex

	raftFsm           *raftFsm
	config            *Config
	raftConfig        *RaftConfig
	restoringSnapshot util.AtomicBool
	curApplied        util.AtomicUInt64
	curSoftSt         unsafe.Pointer
	prevSoftSt        softState
	prevHardSt        proto.HardState

	peerState peerState

	pending map[uint64]*Future
	//snapping          map[uint64]*snapshotStatus
	//mStatus           *monitorStatus
	propc       chan *proposal
	applyc      chan *apply
	receiveChan chan *proto.Message
	//snapRecvc         chan *snapshotRequest
	truncatec chan uint64
	//readIndexC        chan *Future
	//statusc           chan chan *Status
	//entryRequestC     chan *entryRequest
	readyc chan struct{}
	tickc  chan struct{}
	electc chan struct{}
	stopc  chan struct{}
	done   chan struct{}
}

func newRaft(config *Config, raftConfig *RaftConfig) (*Raft, error) {
	if err := raftConfig.validate(); err != nil {
		return nil, err
	}

	raftFSM, err := NewRaftFsm(config, raftConfig)
	if err != nil {
		return nil, err
	}

	raft := &Raft{
		raftFsm:    raftFSM,
		config:     config,
		raftConfig: raftConfig,
		mStatus:    mStatus,
		pending:    make(map[uint64]*Future),
		//snapping:      make(map[uint64]*snapshotStatus),
		receiveChan: make(chan *proto.Message, config.ReqBufferSize),
		applyc:      make(chan *apply, config.AppBufferSize),
		propc:       make(chan *proposal, 256),
		//snapRecvc:     make(chan *snapshotRequest, 1),
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

	//go raft.runApply()
	go raft.run()
	//go raft.monitor()

	return raft, nil
}

func (r *Raft) run() {
	for {
		select {
		case <-r.tickc:
			klog.Infof(fmt.Sprintf("raft tick"))

		case message := <-r.receiveChan:
			klog.Infof(fmt.Sprintf("%+v", *message))
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
