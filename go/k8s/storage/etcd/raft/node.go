package raft

import (
	"context"
	"errors"
	"fmt"

	pb "go.etcd.io/etcd/raft/v3/raftpb"
	"k8s.io/klog/v2"
)

// INFO:
//  ////////////////////////////////////////////////////////////////////////////////////////////
//  整套代码里，开启一个对象，然后通过 stop/done channel 来 start()/stop() 一个 loop，这种方式值得学习!!!
//  ///////////////////////////////////////////////////////////////////////////////////////////

type SnapshotStatus int

const (
	SnapshotFinish  SnapshotStatus = 1
	SnapshotFailure SnapshotStatus = 2
)

var (
	emptyState = pb.HardState{}

	// ErrStopped is returned by methods on Nodes that have been stopped.
	ErrStopped = errors.New("raft: stopped")
)

// SoftState provides state that is useful for logging and debugging.
// The state is volatile and does not need to be persisted to the WAL.
type SoftState struct {
	Lead      uint64 // must use atomic operations to access; keep 64-bit aligned.
	RaftState StateType
}

func (a *SoftState) equal(b *SoftState) bool {
	return a.Lead == b.Lead && a.RaftState == b.RaftState
}

type Node interface {

	// Tick INFO: 每一个 node 的逻辑时钟，用来判断 heartbeatTimeout 和 electionTimeout
	Tick()

	// Campaign INFO: Follower 变成 Candidate, 竞选成 Leader
	Campaign(ctx context.Context) error

	// ReadIndex INFO: 线性一致性读, https://time.geekbang.org/column/article/335932
	ReadIndex(ctx context.Context, rctx []byte) error

	// Propose INFO: 提交数据到 raft log
	Propose(ctx context.Context, data []byte) error

	// ProposeConfChange proposes a configuration change.
	ProposeConfChange(ctx context.Context, cc pb.ConfChangeI) error

	// Step advances the state machine using the given message. ctx.Err() will be returned, if any.
	Step(ctx context.Context, msg pb.Message) error

	// Ready returns a channel that returns the current point-in-time state.
	// Users of the Node must call Advance after retrieving the state returned by Ready.
	//
	// NOTE: No committed entries from the next Ready may be applied until all committed entries
	// and snapshots from the previous one have finished.
	Ready() <-chan Ready

	// Advance The application should generally call Advance after it applies the entries in last Ready.
	Advance()

	Stop()
}

type Peer struct {
	ID      uint64
	Context []byte
}

// StartNode INFO: 启动 raft node，同时 ConfChangeAddNode 添加 peers
func StartNode(config *Config, peers []Peer) Node {
	if len(peers) == 0 {
		panic("no peers given; use RestartNode instead")
	}

	n := newNode(config)
	err := n.Bootstrap(peers)
	if err != nil {
		klog.Errorf(fmt.Sprintf("error occurred during starting a new node: %v", err))
	}

	go n.run()

	return n
}

// RestartNode INFO: 启动 raft node，没有添加 peers
func RestartNode(config *Config) Node {
	n := newNode(config)

	go n.run()

	return n
}

// ReadState INFO: 线性一致性读, 返回给 read-only node 该 ReadState
type ReadState struct {
	Index      uint64
	RequestCtx []byte
}

type ReadOnlyOption int

const (
	// ReadOnlySafe guarantees the linearizability of the read only request by
	// communicating with the quorum. It is the default and suggested option.
	ReadOnlySafe ReadOnlyOption = iota

	// ReadOnlyLeaseBased ensures linearizability of the read only request by
	// relying on the leader lease. It can be affected by clock drift.
	// If the clock drift is unbounded, leader might keep the lease longer than it
	// should (clock can move backward/pause without any bound). ReadIndex is not safe
	// in that case.
	ReadOnlyLeaseBased
)

type readIndexStatus struct {
	req   pb.Message
	index uint64
	// NB: this never records 'false', but it's more convenient to use this
	// instead of a map[uint64]struct{} due to the API of quorum.VoteResult. If
	// this becomes performance sensitive enough (doubtful), quorum.VoteResult
	// can change to an API that is closer to that of CommittedIndex.
	acks map[uint64]bool
}

type readOnly struct {
	option           ReadOnlyOption
	pendingReadIndex map[string]*readIndexStatus
	readIndexQueue   []string
}

func newReadOnly(option ReadOnlyOption) *readOnly {
	return &readOnly{
		option:           option,
		pendingReadIndex: make(map[string]*readIndexStatus),
	}
}

// lastPendingRequestCtx returns the context of the last pending read only
// request in readonly struct.
func (ro *readOnly) lastPendingRequestCtx() string {
	if len(ro.readIndexQueue) == 0 {
		return ""
	}
	return ro.readIndexQueue[len(ro.readIndexQueue)-1]
}

type msgWithResult struct {
	message pb.Message
	result  chan error
}

// INFO: node in raft cluster
type node struct {
	raft       *raft
	prevSoftSt *SoftState
	prevHardSt pb.HardState

	tickChan chan struct{}

	proposeChan    chan msgWithResult
	confChangeChan chan pb.ConfChangeV2
	confStateChan  chan pb.ConfState
	receiveChan    chan pb.Message
	readyChan      chan Ready
	advanceChan    chan struct{} // 只是个标记 channel

	doneChan chan struct{}
	stopChan chan struct{}
}

func newNode(config *Config) *node {
	r := newRaft(config)
	n := &node{
		raft: r,

		// INFO: 这里 tickChan 是一个 buffer chan，这样消费者(run goroutine)消费不过来时，可以 buffer 下
		tickChan: make(chan struct{}, 128),

		readyChan: make(chan Ready),

		proposeChan:    make(chan msgWithResult),
		confChangeChan: make(chan pb.ConfChangeV2),
		confStateChan:  make(chan pb.ConfState),
		receiveChan:    make(chan pb.Message),
		advanceChan:    make(chan struct{}),

		doneChan: make(chan struct{}),
		stopChan: make(chan struct{}),
	}
	n.prevHardSt = r.hardState()
	n.prevSoftSt = r.softState()

	return n
}

func (n *node) run() {
	var readyChan chan Ready
	var ready Ready
	var advanceChan chan struct{}
	var proposeChan chan msgWithResult

	r := n.raft
	lead := None

	for {
		if advanceChan != nil {
			readyChan = nil
		} else if n.HasReady() {
			// INFO: 会从 raft.msgs 获取用户提交的 []pb.Message
			ready = n.readyWithoutAccept()
			readyChan = n.readyChan
		}

		if lead != r.lead {
			if r.hasLeader() {
				if lead == None {
					klog.Infof(fmt.Sprintf("raft.node: %x elected leader %x at term %d", r.id, r.lead, r.Term))
				} else {
					klog.Infof(fmt.Sprintf("raft.node: %x changed leader from %x to %x at term %d", r.id, lead, r.lead, r.Term))
				}
				proposeChan = n.proposeChan
			} else {
				klog.Infof(fmt.Sprintf("raft.node: %x lost leader %x at term %d", r.id, lead, r.Term))
				proposeChan = nil
			}

			lead = r.lead
		}

		select {
		case <-n.tickChan:
			n.raft.tick()

		case msgResult := <-proposeChan: // INFO: 用户提交的 Put/Delete Propose
			message := msgResult.message
			message.From = r.id
			err := r.Step(message) // INFO: raft Step 提交消息到状态机
			if msgResult.result != nil {
				msgResult.result <- err
				close(msgResult.result)
			}

		case message := <-n.receiveChan:
			// filter out response message from unknown From.
			if pr := r.progress.Progress[message.From]; pr != nil || !IsResponseMsg(message.Type) {
				r.Step(message)
			}

		case confChange := <-n.confChangeChan:
			_, okBefore := r.progress.Progress[r.id]
			confState := r.applyConfChange(confChange)
			if _, okAfter := r.progress.Progress[r.id]; okBefore && !okAfter {
				var found bool
				for _, sl := range [][]uint64{confState.Voters, confState.VotersOutgoing} {
					for _, id := range sl {
						if id == r.id {
							found = true
							break
						}
					}
					if found {
						break
					}
				}
				if !found {
					proposeChan = nil
				}
			}

			select {
			case n.confStateChan <- confState:
			case <-n.doneChan:
			}

		// INFO: 用户提交的 pb.Message 从这 readyChan 获取
		case readyChan <- ready:
			n.acceptReady(ready)
			advanceChan = n.advanceChan

		// INFO: @see Advance(), 推动用户提交的 []pb.Message 提交到 log 模块中，这里才是最终目标!!!
		case <-advanceChan:
			n.AdvanceRaft(ready)
			ready = Ready{}
			advanceChan = nil

		case <-n.stopChan:
			close(n.doneChan)
			return
		}
	}
}

func (n *node) Bootstrap(peers []Peer) error {
	if len(peers) == 0 {
		return errors.New("must provide at least one peer to Bootstrap")
	}

	lastIndex := n.raft.raftLog.storage.LastIndex()
	if lastIndex != 0 {
		return errors.New("can't bootstrap a nonempty Storage")
	}

	n.prevHardSt = emptyState

	n.raft.becomeFollower(1, None)
	ents := make([]pb.Entry, len(peers))
	for i, peer := range peers {
		cc := pb.ConfChange{Type: pb.ConfChangeAddNode, NodeID: peer.ID, Context: peer.Context}
		data, err := cc.Marshal()
		if err != nil {
			return err
		}

		ents[i] = pb.Entry{Type: pb.EntryConfChange, Term: 1, Index: uint64(i + 1), Data: data}
	}
	n.raft.raftLog.append(ents...)

	n.raft.raftLog.committed = uint64(len(ents))
	for _, peer := range peers {
		n.raft.applyConfChange(pb.ConfChange{NodeID: peer.ID, Type: pb.ConfChangeAddNode}.AsV2())
	}

	return nil
}

// Propose INFO: (INPUT)提交数据到 raft log
func (n *node) Propose(ctx context.Context, data []byte) error {
	return n.stepWait(ctx, pb.Message{Type: pb.MsgProp, Entries: []pb.Entry{{Data: data}}})
}

// Ready INFO: OUTPUT
func (n *node) Ready() <-chan Ready {
	return n.readyChan
}

// Tick INFO: raft node 逻辑时钟，用来判断 heartbeatTimeout 和 electionTimeout
func (n *node) Tick() {
	select {
	case n.tickChan <- struct{}{}:
	case <-n.doneChan:
	default:
		klog.Warningf(fmt.Sprintf("[Tick]%x tick missed to fire. Node blocks too long!", n.raft.id))
	}
}

// HasReady called when RawNode user need to check if any Ready pending.
// Checking logic in this method should be consistent with Ready.containsUpdates().
func (n *node) HasReady() bool {
	r := n.raft
	if !r.softState().equal(n.prevSoftSt) {
		return true
	}
	if hardSt := r.hardState(); !IsEmptyHardState(hardSt) && !isHardStateEqual(hardSt, n.prevHardSt) {
		return true
	}
	if r.raftLog.hasPendingSnapshot() {
		return true
	}
	if len(r.msgs) > 0 || len(r.raftLog.unstableEntries()) > 0 || r.raftLog.hasNextEnts() {
		return true
	}
	if len(r.readStates) != 0 {
		return true
	}

	return false
}

func (n *node) readyWithoutAccept() Ready {
	return n.raft.newReady(n.prevSoftSt, n.prevHardSt)
}

// acceptReady is called when the consumer of the RawNode has decided to go
// ahead and handle a Ready. Nothing must alter the state of the RawNode between
// this call and the prior call to Ready().
func (n *node) acceptReady(ready Ready) {
	if ready.SoftState != nil {
		n.prevSoftSt = ready.SoftState
	}
	if len(ready.ReadStates) != 0 {
		n.raft.readStates = nil
	}

	n.raft.msgs = nil
}

// AdvanceRaft INFO: 推动用户提交的 []pb.Message 提交到 log 模块中，这里才是最终目标!!!
func (n *node) AdvanceRaft(ready Ready) {
	if !IsEmptyHardState(ready.HardState) {
		n.prevHardSt = ready.HardState
	}

	n.raft.advance(ready)
}

func (n *node) ReadIndex(ctx context.Context, readCtx []byte) error {
	return n.step(ctx, pb.Message{Type: pb.MsgReadIndex, Entries: []pb.Entry{{Data: readCtx}}})
}

// Campaign INFO: 参加竞选，选为 leader
func (n *node) Campaign(ctx context.Context) error {
	return n.step(ctx, pb.Message{Type: pb.MsgHup})
}

func confChangeToMsg(c pb.ConfChangeI) (pb.Message, error) {
	typ, data, err := pb.MarshalConfChange(c)
	if err != nil {
		return pb.Message{}, err
	}
	return pb.Message{Type: pb.MsgProp, Entries: []pb.Entry{{Type: typ, Data: data}}}, nil
}

// ProposeConfChange INFO: raft 成员变更也是一个 pb.MsgProp message, 和普通的 log entry 一样，
func (n *node) ProposeConfChange(ctx context.Context, cc pb.ConfChangeI) error {
	msg, err := confChangeToMsg(cc)
	if err != nil {
		return err
	}

	return n.Step(ctx, msg)
}

func (n *node) ApplyConfChange(cc pb.ConfChangeI) *pb.ConfState {
	var confState pb.ConfState
	select {
	case n.confChangeChan <- cc.AsV2():
	case <-n.doneChan:
	}
	// block until async goroutine done
	select {
	case confState = <-n.confStateChan:
	case <-n.doneChan:
	}

	return &confState
}

func (n *node) Step(ctx context.Context, m pb.Message) error {
	return n.step(ctx, m)
}
func (n *node) step(ctx context.Context, m pb.Message) error {
	return n.stepWithWaitOption(ctx, m, false)
}
func (n *node) stepWait(ctx context.Context, m pb.Message) error {
	return n.stepWithWaitOption(ctx, m, true)
}

// Step advances the state machine using message. The ctx.Err() will be returned, if any.
func (n *node) stepWithWaitOption(ctx context.Context, message pb.Message, wait bool) error {
	if message.Type != pb.MsgProp {
		select {
		// INFO: campaign 走这个 n.receiveChan, run() 里监听了 n.receiveChan
		case n.receiveChan <- message:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		case <-n.doneChan:
			return ErrStopped
		}
	}

	msgResult := msgWithResult{message: message}
	if wait {
		msgResult.result = make(chan error, 1)
	}
	select {
	case n.proposeChan <- msgResult: // INFO: 参考 run() msgResult := <-proposeChan
		if !wait {
			return nil
		}
	case <-ctx.Done():
		return ctx.Err()
	case <-n.doneChan:
		return ErrStopped
	}

	select {
	case err := <-msgResult.result:
		if err != nil {
			return err
		}
	case <-ctx.Done():
		return ctx.Err()
	case <-n.doneChan:
		return ErrStopped
	}

	return nil
}

func (n *node) Advance() {
	select {
	case n.advanceChan <- struct{}{}:
		klog.Infof(fmt.Sprintf("n.advanceChan <- struct{}{}"))
	case <-n.stopChan:
	}
}

func (n *node) Stop() {
	select {
	case n.stopChan <- struct{}{}:
		// Not already stopped, so trigger it
	case <-n.doneChan:
		// Node has already been stopped - no need to do anything
		return
	}

	// INFO: 在 run() 里会 ack 这个 channel
	<-n.doneChan
}

func IsEmptyHardState(st pb.HardState) bool {
	return isHardStateEqual(st, emptyState)
}

func isHardStateEqual(a, b pb.HardState) bool {
	return a.Term == b.Term && a.Vote == b.Vote && a.Commit == b.Commit
}

// IsEmptySnap returns true if the given Snapshot is empty.
func IsEmptySnap(sp pb.Snapshot) bool {
	return sp.Metadata.Index == 0
}
