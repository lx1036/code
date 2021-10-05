package raft

import (
	"context"
	"fmt"
	"k8s.io/klog/v2"
	"reflect"
	"testing"
	"time"

	"go.etcd.io/etcd/raft/v3/raftpb"
)

func newTestConfig(id uint64, election, heartbeat int, storage Storage) *Config {
	return &Config{
		ID:              id,
		ElectionTick:    election,
		HeartbeatTick:   heartbeat,
		Storage:         storage,
		MaxSizePerMsg:   noLimit,
		MaxInflightMsgs: 256,
	}
}

func newTestRaft(id uint64, election, heartbeat int, storage Storage) *raft {
	return newRaft(newTestConfig(id, election, heartbeat, storage))
}

type testMemoryStorageOptions func(*MemoryStorage)

func withPeers(peers ...uint64) testMemoryStorageOptions {
	return func(ms *MemoryStorage) {
		ms.snapshot.Metadata.ConfState.Voters = peers
	}
}

func withLearners(learners ...uint64) testMemoryStorageOptions {
	return func(ms *MemoryStorage) {
		ms.snapshot.Metadata.ConfState.Learners = learners
	}
}

func newTestMemoryStorage(opts ...testMemoryStorageOptions) *MemoryStorage {
	ms := NewMemoryStorage()
	for _, o := range opts {
		o(ms)
	}
	return ms
}

func TestAddLearnerNode(t *testing.T) {
	r := newTestRaft(1, 10, 1, newTestMemoryStorage(withPeers(1)))
	// INFO: (1) Add new learner peer
	r.applyConfChange(raftpb.ConfChange{NodeID: 2, Type: raftpb.ConfChangeAddLearnerNode}.AsV2())
	if r.isLearner {
		// 这里 raft.isLearner 是当前 local node 状态
		t.Fatal("expected 1 to be voter")
	}
	nodes := r.prs.LearnerNodes()
	wantedNodes := []uint64{2}
	if !reflect.DeepEqual(nodes, wantedNodes) {
		t.Errorf("nodes = %v, want %v", nodes, wantedNodes)
	}
	if !r.prs.Progress[2].IsLearner {
		t.Fatal("expected 2 to be learner")
	}
	klog.Infof(fmt.Sprintf("VoterNodes: %+v, LearnerNodes: %+v", r.prs.VoterNodes(), r.prs.LearnerNodes()))

	// INFO: (2) Promote learner to voter
	r.applyConfChange(raftpb.ConfChange{NodeID: 2, Type: raftpb.ConfChangeAddNode}.AsV2())
	if r.prs.Progress[2].IsLearner {
		t.Fatal("expected 2 to be voter")
	}
	if r.isLearner {
		t.Fatal("expected 2 to be voter")
	}
	klog.Infof(fmt.Sprintf("VoterNodes: %+v, LearnerNodes: %+v", r.prs.VoterNodes(), r.prs.LearnerNodes()))

	// INFO: (3) Demote voter to learner
	r.applyConfChange(raftpb.ConfChange{NodeID: 1, Type: raftpb.ConfChangeAddLearnerNode}.AsV2())
	if !r.prs.Progress[1].IsLearner {
		t.Fatal("expected 1 to be learner")
	}
	if !r.isLearner {
		t.Fatal("expected 1 to be learner")
	}
	klog.Infof(fmt.Sprintf("VoterNodes: %+v, LearnerNodes: %+v", r.prs.VoterNodes(), r.prs.LearnerNodes()))

	// INFO: (4) Promote learner to voter
	r.applyConfChange(raftpb.ConfChange{NodeID: 1, Type: raftpb.ConfChangeAddNode}.AsV2())
	if r.prs.Progress[1].IsLearner {
		t.Fatal("expected 1 to be voter")
	}
	if r.isLearner {
		t.Fatal("expected 1 to be voter")
	}
	klog.Infof(fmt.Sprintf("VoterNodes: %+v, LearnerNodes: %+v", r.prs.VoterNodes(), r.prs.LearnerNodes()))

	// INFO: (5) Remove voter
	r.applyConfChange(raftpb.ConfChange{NodeID: 2, Type: raftpb.ConfChangeRemoveNode}.AsV2())
	if len(r.prs.VoterNodes()) != 1 {
		t.Fatal("expected nodes number to be 1")
	}
	klog.Infof(fmt.Sprintf("VoterNodes: %+v, LearnerNodes: %+v", r.prs.VoterNodes(), r.prs.LearnerNodes()))

	// INFO: (6) promote state machine to be leader
	fixtures := []struct {
		desc       string
		peers      []uint64
		promotable bool
	}{
		{
			desc:       "{1} promotable",
			peers:      []uint64{1},
			promotable: true,
		},
		{
			desc:       "{1,2,3} promotable",
			peers:      []uint64{1, 2, 3},
			promotable: true,
		},
		{
			desc:       "{} not promotable",
			peers:      []uint64{},
			promotable: false,
		},
		{
			desc:       "{2,3} not promotable",
			peers:      []uint64{2, 3},
			promotable: false,
		},
	}
	for _, fixture := range fixtures {
		t.Run(fixture.desc, func(t *testing.T) {
			r2 := newTestRaft(1, 10, 1, newTestMemoryStorage(withPeers(fixture.peers...)))
			if promotable := r2.promotable(); promotable != fixture.promotable {
				t.Fatalf("promotable = %v, want %v", promotable, fixture.promotable)
			}
		})
	}
}

func (r *raft) readMessages() []raftpb.Message {
	msgs := r.msgs
	r.msgs = make([]raftpb.Message, 0)

	return msgs
}

func TestRaftFlowControl(test *testing.T) {
	r := newTestRaft(1, 10, 1, newTestMemoryStorage(withPeers(1, 2)))
	r.becomeCandidate()
	r.becomeLeader()

	pr2 := r.prs.Progress[2]
	// force the progress to be in replicate state
	pr2.BecomeReplicate()
	for i := 0; i < r.prs.MaxInflight; i++ {
		r.Step(raftpb.Message{From: 1, To: 1, Type: raftpb.MsgProp, Entries: []raftpb.Entry{{Data: []byte("somedata")}}})
		ms := r.readMessages()
		if len(ms) != 1 {
			test.Fatalf("#%d: len(ms) = %d, want 1", i, len(ms))
		}

		klog.Infof(fmt.Sprintf("%+v", r.readMessages()))
	}

}

// ensures that a node can be started correctly. The node should
// start with correct configuration change entries, and can accept and commit
// proposals.
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
	c := &Config{
		ID:              1,
		ElectionTick:    10,
		HeartbeatTick:   1,
		Storage:         storage,
		MaxSizePerMsg:   noLimit,
		MaxInflightMsgs: 256,
	}
	node := StartNode(c, []Peer{{ID: 1}})
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
