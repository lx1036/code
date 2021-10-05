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

var (
	emptyState = pb.HardState{}

	// ErrStopped is returned by methods on Nodes that have been stopped.
	ErrStopped = errors.New("raft: stopped")
)

// Ready encapsulates the entries and messages that are ready to read,
// be saved to stable storage, committed or sent to other peers.
// All fields in Ready are read-only.
type Ready struct {
	Entries []pb.Entry

	pb.HardState

	*SoftState

	// ReadStates can be used for node to serve linearizable read requests locally
	// when its applied index is greater than the index in ReadState.
	// Note that the readState will be returned when raft receives msgReadIndex.
	// The returned is only valid for the request that requested to read.
	ReadStates []ReadState

	// INFO: 已经提交到 state-machine 中的 []pb.Entry
	CommittedEntries []pb.Entry

	Snapshot pb.Snapshot

	Messages []pb.Message

	MustSync bool
}

func newReady(r *raft, prevSoftSt *SoftState, prevHardSt pb.HardState) Ready {
	ready := Ready{
		Entries:          r.raftLog.unstableEntries(),
		CommittedEntries: r.raftLog.nextEnts(),
		Messages:         r.msgs,
	}
	if softSt := r.softState(); !softSt.equal(prevSoftSt) {
		ready.SoftState = softSt
	}
	if hardSt := r.hardState(); !isHardStateEqual(hardSt, prevHardSt) {
		ready.HardState = hardSt
	}
	if r.raftLog.unstable.snapshot != nil {
		ready.Snapshot = *r.raftLog.unstable.snapshot
	}
	if len(r.readStates) != 0 {
		ready.ReadStates = r.readStates
	}
	ready.MustSync = MustSync(r.hardState(), prevHardSt, len(ready.Entries))

	return ready
}

// MustSync returns true if the hard state and count of Raft entries indicate
// that a synchronous write to persistent storage is required.
func MustSync(st, prevst pb.HardState, entsnum int) bool {
	// Persistent state on all servers:
	// (Updated on stable storage before responding to RPCs)
	// currentTerm
	// votedFor
	// log entries[]
	return entsnum != 0 || st.Vote != prevst.Vote || st.Term != prevst.Term
}

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

	// Campaign INFO: Follower 变成 Candidate, 竞选成 Leader
	Campaign(ctx context.Context) error

	// ReadIndex INFO: 线性一致性读, https://time.geekbang.org/column/article/335932
	ReadIndex(ctx context.Context, rctx []byte) error

	// Propose INFO: 提交数据到 raft log
	Propose(ctx context.Context, data []byte) error

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

func StartNode(c *Config, peers []Peer) Node {
	if len(peers) == 0 {
		panic("no peers given; use RestartNode instead")
	}
	rawNode, err := NewRawNode(c)
	if err != nil {
		panic(err)
	}
	err = rawNode.Bootstrap(peers)
	if err != nil {
		klog.Errorf(fmt.Sprintf("error occurred during starting a new node: %v", err))
	}

	n := newNode(rawNode)

	go n.run()

	return n
}

type msgWithResult struct {
	message pb.Message
	result  chan error
}

// INFO: node in raft cluster
type node struct {
	rawNode *RawNode

	receiveChan chan pb.Message

	readyChan   chan Ready
	advanceChan chan struct{} // 只是个标记 channel
	proposeChan chan msgWithResult

	doneChan chan struct{}
	stopChan chan struct{}
}

func newNode(rawNode *RawNode) *node {
	return &node{
		rawNode: rawNode,

		readyChan: make(chan Ready),

		receiveChan: make(chan pb.Message),
		advanceChan: make(chan struct{}),
		proposeChan: make(chan msgWithResult),

		doneChan: make(chan struct{}),
		stopChan: make(chan struct{}),
	}
}

func (n *node) run() {
	var readyChan chan Ready
	var ready Ready
	var advanceChan chan struct{}
	var proposeChan chan msgWithResult

	r := n.rawNode.raft
	lead := None

	for {
		if advanceChan != nil {
			readyChan = nil
		} else if n.rawNode.HasReady() {
			ready = n.rawNode.readyWithoutAccept()
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
		case msgResult := <-proposeChan:
			message := msgResult.message
			message.From = r.id
			err := r.Step(message)
			if msgResult.result != nil {
				msgResult.result <- err
				close(msgResult.result)
			}
		case message := <-n.receiveChan:
			// filter out response message from unknown From.
			if pr := r.prs.Progress[message.From]; pr != nil || !IsResponseMsg(message.Type) {
				r.Step(message)
			}
		case readyChan <- ready:
			klog.Infof(fmt.Sprintf("readyChan <- ready"))
			n.rawNode.acceptReady(ready)
			advanceChan = n.advanceChan
		case <-advanceChan:
			klog.Infof(fmt.Sprintf("<-advanceChan"))
			n.rawNode.Advance(ready)
			ready = Ready{}
			advanceChan = nil
		case <-n.stopChan:
			close(n.doneChan)
			return
		}
	}
}

func (n *node) ReadIndex(ctx context.Context, readCtx []byte) error {
	return n.step(ctx, pb.Message{Type: pb.MsgReadIndex, Entries: []pb.Entry{{Data: readCtx}}})
}

// Campaign INFO: 参加竞选，选为 leader
func (n *node) Campaign(ctx context.Context) error {
	return n.step(ctx, pb.Message{Type: pb.MsgHup})
}

// Propose INFO: 提交数据到 raft log
func (n *node) Propose(ctx context.Context, data []byte) error {
	return n.stepWait(ctx, pb.Message{Type: pb.MsgProp, Entries: []pb.Entry{{Data: data}}})
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
		case n.receiveChan <- message: // run() 里监听了 n.receiveChan
			return nil
		case <-ctx.Done():
			return ctx.Err()
		case <-n.doneChan:
			return ErrStopped
		}
	}

	// TODO: pb.MsgProp
	ch := n.proposeChan
	msgResult := msgWithResult{message: message}
	if wait {
		msgResult.result = make(chan error, 1)
	}
	select {
	case ch <- msgResult: // INFO: 参考 run() msgResult := <-proposeChan
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

func (n *node) Ready() <-chan Ready {
	return n.readyChan
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
