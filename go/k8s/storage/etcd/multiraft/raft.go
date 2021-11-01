package multiraft

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"unsafe"

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
	Storage      Storage
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
	/*if c.Storage == nil {
		return errors.New("Storage is required")
	}
	if c.StateMachine == nil {
		return errors.New("StateMachine is required")
	}*/

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
	config            *NodeConfig
	raftConfig        *RaftConfig
	restoringSnapshot util.AtomicBool
	curApplied        util.AtomicUInt64

	curSoftSt  unsafe.Pointer
	prevSoftSt softState

	prevHardSt proto.HardState

	peerState peerState

	//pending map[uint64]*Future
	//snapping          map[uint64]*snapshotStatus
	//mStatus           *monitorStatus
	//applyc      chan *apply
	receiveChan chan *proto.Message
	proposeChan chan *proposal

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

func newRaft(config *NodeConfig, raftConfig *RaftConfig) (*Raft, error) {
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
		//mStatus:    mStatus,
		//pending:    make(map[uint64]*Future),
		//snapping:      make(map[uint64]*snapshotStatus),

		receiveChan: make(chan *proto.Message, config.ReqBufferSize),
		proposeChan: make(chan *proposal, 256),

		//applyc:      make(chan *apply, config.AppBufferSize),
		//snapRecvc:     make(chan *snapshotRequest, 1),
		truncatec: make(chan uint64, 1),
		//readIndexC:    make(chan *Future, 256),
		//statusc:       make(chan chan *Status, 1),
		//entryRequestC: make(chan *entryRequest, 16),
		tickc:  make(chan struct{}, 64),
		readyc: make(chan struct{}, 1),
		electc: make(chan struct{}, 1),
		stopc:  make(chan struct{}),
		done:   make(chan struct{}),
	}
	raft.curApplied.Set(raftFSM.log.applied)
	raft.peerState.replace(raftConfig.Peers)

	//go raft.runApply()
	go raft.run()
	//go raft.monitor()

	return raft, nil
}

func (r *Raft) run() {
	r.prevHardSt.Term = r.raftFsm.term
	r.prevHardSt.Vote = r.raftFsm.vote
	r.prevHardSt.Commit = r.raftFsm.log.committed
	r.updateCurrentSoftState()

	var readyc chan struct{}
	for {
		if readyc == nil && r.containsUpdate() {
			readyc = r.readyc
			readyc <- struct{}{}
		}

		select {
		case <-r.tickc:
			r.updateCurrentSoftState()
			klog.Infof(fmt.Sprintf("[Raft run]raft tick"))

		case message := <-r.receiveChan:
			klog.Infof(fmt.Sprintf("[Raft run]message: %+v", *message))
			switch message.Type {
			// INFO: 只有 Follower 处理 ReqMsgHeartBeat message
			case proto.ReqMsgHeartBeat:
				if r.raftFsm.leader == message.From { // debug in local
					//if r.raftFsm.leader == message.From && message.From != r.config.NodeID {
					r.raftFsm.Step(message)
				}

			// INFO: 只有 Leader 处理 RespMsgHeartBeat message
			case proto.RespMsgHeartBeat:
				if r.raftFsm.leader == r.config.NodeID { // debug in local
					//if r.raftFsm.leader == r.config.NodeID && message.From != r.config.NodeID {
					r.raftFsm.Step(message)
				}

			// INFO: 所有非心跳消息，发给 raft peers 推动 raft peer 状态机转动
			default:
				r.raftFsm.Step(message)
			}

		case propose := <-r.proposeChan:

			msg := proto.NewMessage()
			msg.Type = proto.LocalMsgProp
			msg.From = r.config.NodeID
			startIndex := r.raftFsm.log.LastIndex() + 1 // leader raft 中的 log LastIndex+1
			msg.Entries = append(msg.Entries, &proto.Entry{Term: r.raftFsm.term, Index: startIndex, Type: propose.cmdType, Data: propose.data})

			// TODO: 这里从 r.proposeChan 批量取值，总共最多批量取 64 个。后续重构！！！
			flag := false
			for i := 1; i < 64; i++ {
				startIndex = startIndex + 1
				select {
				case propose2 := <-r.proposeChan:
					msg.Entries = append(msg.Entries, &proto.Entry{Term: r.raftFsm.term, Index: startIndex, Type: propose2.cmdType, Data: propose2.data})
				default:
					flag = true
				}
				if flag {
					break
				}
			}

			// INFO: leader raft 收到外部输入，然后推动 leader raft 和 follower raft 转动
			r.raftFsm.Step(msg)

		case <-readyc:
			// Send all messages.
			for _, msg := range r.raftFsm.msgs {
				if msg.Type == proto.ReqMsgSnapShot {
					//r.sendSnapshot(msg)
					continue
				}
				r.sendMessage(msg)
			}
			r.raftFsm.msgs = nil

			readyc = nil
		}
	}

}

type proposal struct {
	cmdType proto.EntryType
	//future  *Future
	data []byte
}

// INFO: raft 输入来自于 propose(cmd)
func (r *Raft) propose(cmd []byte) {
	if !r.isLeader() {
		return
	}

	propose := &proposal{
		data:    cmd,
		cmdType: proto.EntryNormal,
	}

	select {
	case <-r.stopc:
	case r.proposeChan <- propose:
	}
}

func (r *Raft) proposeMemberChange(cc *proto.ConfChange) {
	if !r.isLeader() {
		return
	}

	data := cc.Encode()
	propose := &proposal{
		data:    data,
		cmdType: proto.EntryConfChange,
	}

	select {
	case <-r.stopc:
	case r.proposeChan <- propose:
	}
}

// INFO: 每一个 partition 是一个 raft，并且只有 leader node 上的 partition raft 都是 leader
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

func (r *Raft) applied() uint64 {
	return r.curApplied.Get()
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
	/*if r.restoringSnapshot.Get() {
		return
	}*/

	select {
	case <-r.stopc:
	case r.receiveChan <- message:
	default:
		klog.Warningf(fmt.Sprintf("[Raft receiveMessage]raft[%v] discard message(%v)", r.raftConfig.ID, message.ToString()))
		return
	}
}

func (r *Raft) updateCurrentSoftState() {
	updated := false
	if r.prevSoftSt.term != r.raftFsm.term {
		updated = true
		r.prevSoftSt.term = r.raftFsm.term
		//r.resetTick()
	}

	preLeader := r.prevSoftSt.leader
	if preLeader != r.raftFsm.leader {
		updated = true
		r.prevSoftSt.leader = r.raftFsm.leader
		klog.Infof(fmt.Sprintf("[Raft updateCurrentSoftState]change leader from %d to %d", preLeader, r.raftFsm.leader))
		if r.raftFsm.leader != r.config.NodeID {
			/*if preLeader != r.config.NodeID {
				r.resetPending(ErrNotLeader)
			}
			r.stopSnapping()*/
		}

		//r.raftConfig.StateMachine.HandleLeaderChange(r.raftFsm.leader)
	}

	if updated {
		atomic.StorePointer(&r.curSoftSt, unsafe.Pointer(&softState{leader: r.raftFsm.leader, term: r.raftFsm.term}))
	}

	curSoftState := (*softState)(atomic.LoadPointer(&r.curSoftSt))
	klog.Infof(fmt.Sprintf("[Raft updateCurrentSoftState]current leader:%d, term:%d", curSoftState.leader, curSoftState.term))
}

func (r *Raft) sendMessage(m *proto.Message) {
	if r.config.transport != nil {
		r.config.transport.Send(m)
	} else {
		if len(m.Entries) != 0 {
			klog.Infof(fmt.Sprintf("[raft sendMessage]%s", m.Entries[0].Data))
		}
	}
}

func (r *Raft) containsUpdate() bool {
	return len(r.raftFsm.log.unstableEntries()) > 0 || r.raftFsm.log.committed > r.raftFsm.log.applied || len(r.raftFsm.msgs) > 0 ||
		r.raftFsm.log.committed != r.prevHardSt.Commit || r.raftFsm.term != r.prevHardSt.Term || r.raftFsm.vote != r.prevHardSt.Vote //||
	//r.raftFsm.readOnly.containsUpdate(r.curApplied.Get())
}

func (r *Raft) status() *Status {
	panic("not implemented")
}
