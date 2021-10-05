package raft

import (
	"context"
	"errors"
	"fmt"

	pb "go.etcd.io/etcd/raft/v3/raftpb"
	"k8s.io/klog/v2"
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

type Node interface {

	// Campaign INFO: Follower 变成 Candidate, 竞选成 Leader
	Campaign(ctx context.Context) error

	// ReadIndex INFO: 线性一致性读, https://time.geekbang.org/column/article/335932
	ReadIndex(ctx context.Context, rctx []byte) error
}

type Peer struct {
	ID      uint64
	Context []byte
}

func StartNode(c *Config, peers []Peer) Node {
	if len(peers) == 0 {
		panic("no peers given; use RestartNode instead")
	}
	rn, err := NewRawNode(c)
	if err != nil {
		panic(err)
	}
	err = rn.Bootstrap(peers)
	if err != nil {
		klog.Errorf(fmt.Sprintf("error occurred during starting a new node: %v", err))
	}

	n := newNode(rn)

	go n.run()

	return n
}

// INFO: node in raft cluster
type node struct {
	rawNode *RawNode

	receiveChan chan pb.Message

	doneChan chan struct{}
	stopChan chan struct{}
}

func newNode(rawNode *RawNode) *node {
	return &node{
		rawNode: rawNode,

		receiveChan: make(chan pb.Message),
		doneChan:    make(chan struct{}),
		stopChan:    make(chan struct{}),
	}
}

func (n *node) run() {
	r := n.rawNode.raft

	for {
		select {
		case message := <-n.receiveChan:
			// filter out response message from unknown From.
			if pr := r.prs.Progress[message.From]; pr != nil || !IsResponseMsg(message.Type) {
				r.Step(message)
			}
		case <-n.stopChan:
			close(n.doneChan)
			return
		}
	}

}

func (n *node) ReadIndex(ctx context.Context, readCtx []byte) error {
	return n.step(ctx, pb.Message{Type: pb.MsgReadIndex, Entries: []pb.Entry{{Data: readCtx}}})
}

func (n *node) step(ctx context.Context, m pb.Message) error {
	return n.stepWithWaitOption(ctx, m, false)
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

	return nil
}

func (n *node) Campaign(ctx context.Context) error {
	panic("implement me")
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
