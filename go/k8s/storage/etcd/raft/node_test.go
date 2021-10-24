package raft

import (
	"bytes"
	"context"
	"reflect"
	"testing"
	"time"

	"go.etcd.io/etcd/raft/v3/raftpb"
	"k8s.io/klog/v2"
)

// INFO: (1) 测试 node.Propose()
func TestNodePropose(test *testing.T) {
	msgs := []raftpb.Message{}
	appendStep := func(m raftpb.Message) error {
		msgs = append(msgs, m)
		return nil
	}

	storage := newTestMemoryStorage(withPeers(1))
	config := &Config{
		ID:              1,
		ElectionTick:    20,
		HeartbeatTick:   2,
		Storage:         storage,
		MaxSizePerMsg:   noLimit,
		MaxInflightMsgs: 256,
	}
	n := newNode(config)
	defer n.Stop()
	r := n.raft
	go n.run()

	// campaign leader
	if err := n.Campaign(context.TODO()); err != nil {
		test.Fatal(err)
	}

	for {
		rd := <-n.Ready()
		storage.Append(rd.Entries)
		// change the step function to appendStep until this raft becomes leader
		if rd.SoftState.Lead == r.id {
			r.step = appendStep
			n.Advance()
			break
		}
		n.Advance()
	}

	n.Propose(context.TODO(), []byte("somedata"))

	if len(msgs) != 1 {
		test.Fatalf("len(msgs) = %d, want %d", len(msgs), 1)
	}
	if msgs[0].Type != raftpb.MsgProp {
		test.Errorf("msg type = %d, want %d", msgs[0].Type, raftpb.MsgProp)
	}
	if !bytes.Equal(msgs[0].Entries[0].Data, []byte("somedata")) {
		test.Errorf("data = %v, want %v", msgs[0].Entries[0].Data, []byte("somedata"))
	}
}

func TestNodeReadIndex(t *testing.T) {
	msgs := []raftpb.Message{}
	appendStep := func(m raftpb.Message) error {
		msgs = append(msgs, m)
		return nil
	}

	wrs := []ReadState{{Index: uint64(1), RequestCtx: []byte("somedata")}}

	storage := newTestMemoryStorage(withPeers(1))
	config := &Config{
		ID:              1,
		ElectionTick:    20,
		HeartbeatTick:   2,
		Storage:         storage,
		MaxSizePerMsg:   noLimit,
		MaxInflightMsgs: 256,
	}
	n := newNode(config)
	defer n.Stop()
	r := n.raft
	r.readStates = wrs
	go n.run()

	n.Campaign(context.TODO())
	for {
		rd := <-n.Ready()
		if !reflect.DeepEqual(rd.ReadStates, wrs) {
			t.Errorf("ReadStates = %v, want %v", rd.ReadStates, wrs)
		}

		storage.Append(rd.Entries)

		if rd.SoftState.Lead == r.id {
			n.Advance()
			break
		}
		n.Advance()
	}

	r.step = appendStep
	wrequestCtx := []byte("somedata2")
	n.ReadIndex(context.TODO(), wrequestCtx)

	if len(msgs) != 1 {
		t.Fatalf("len(msgs) = %d, want %d", len(msgs), 1)
	}
	if msgs[0].Type != raftpb.MsgReadIndex {
		t.Errorf("msg type = %d, want %d", msgs[0].Type, raftpb.MsgReadIndex)
	}
	if !bytes.Equal(msgs[0].Entries[0].Data, wrequestCtx) {
		t.Errorf("data = %v, want %v", msgs[0].Entries[0].Data, wrequestCtx)
	}
}

// INFO: raft node 可以 proposeChan/readyChan
func TestNode(test *testing.T) {
	cc := raftpb.ConfChange{Type: raftpb.ConfChangeAddNode, NodeID: 1}
	data, err := cc.Marshal()
	if err != nil {
		test.Fatalf("unexpected marshal error: %v", err)
	}
	wants := []Ready{
		{
			HardState: raftpb.HardState{Term: 1, Commit: 1, Vote: 0},
			Entries: []raftpb.Entry{
				{Type: raftpb.EntryConfChange, Term: 1, Index: 1, Data: data},
			},
			CommittedEntries: []raftpb.Entry{
				{Type: raftpb.EntryConfChange, Term: 1, Index: 1, Data: data},
			},
			MustSync: true,
		},
		{
			HardState:        raftpb.HardState{Term: 2, Commit: 3, Vote: 1},
			Entries:          []raftpb.Entry{{Term: 2, Index: 3, Data: []byte("foo")}},
			CommittedEntries: []raftpb.Entry{{Term: 2, Index: 3, Data: []byte("foo")}},
			MustSync:         true,
		},
	}

	storage := NewMemoryStorage()
	config := &Config{
		ID:              1,
		ElectionTick:    10,
		HeartbeatTick:   1,
		Storage:         storage,
		MaxSizePerMsg:   noLimit,
		MaxInflightMsgs: 256,
	}
	node := StartNode(config, []Peer{{ID: 1}})
	defer node.Stop()

	ready := <-node.Ready()
	if !reflect.DeepEqual(ready, wants[0]) {
		test.Fatalf("#%d: g = %+v,\n want %+v", 1, ready, wants[0])
	} else {
		storage.Append(ready.Entries)
		node.Advance()
	}

	// INFO: 参加竞选
	if err := node.Campaign(context.TODO()); err != nil {
		test.Fatal(err)
	}
	rd := <-node.Ready()
	storage.Append(rd.Entries)
	node.Advance()

	node.Propose(context.TODO(), []byte("foo"))
	if ready2 := <-node.Ready(); !reflect.DeepEqual(ready2, wants[1]) {
		test.Errorf("#%d: g = %+v,\n want %+v", 2, ready2, wants[1])
	} else {
		storage.Append(ready2.Entries)
		node.Advance()
	}

	select {
	case rd := <-node.Ready():
		test.Errorf("unexpected Ready: %+v", rd)
	case <-time.After(time.Millisecond * 100):
	}
}

func TestSelectChannel(test *testing.T) {
	advanceCh := make(chan struct{})
	doneChan := make(chan struct{})

	go func() {
		<-advanceCh
	}()

	select {
	case advanceCh <- struct{}{}:
		klog.Info("advanceCh <- struct{}{}")
	case <-doneChan:
		klog.Info("<-doneChan")
	}

	klog.Info("done")

	// output:
	// advanceCh <- struct{}{}
	// done
}
