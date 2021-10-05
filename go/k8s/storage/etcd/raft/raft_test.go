package raft

import (
	"fmt"
	"reflect"
	"testing"

	"go.etcd.io/etcd/raft/v3/raftpb"
	"k8s.io/klog/v2"
)

func newTestConfig(id uint64, election, heartbeat int, storage Storage) *Config {
	return &Config{
		ID:              id,
		ElectionTick:    election,
		HeartbeatTick:   heartbeat,
		Storage:         storage,
		MaxSizePerMsg:   noLimit,
		MaxInflightMsgs: 10,
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

// INFO: 加节点, 主要就是修改 progress conf
func TestAddAndRemoveNode(test *testing.T) {
	r := newTestRaft(2, 10, 1, newTestMemoryStorage(withPeers(2)))

	// INFO: (1) Add node, 加一个 node=1 节点
	r.applyConfChange(raftpb.ConfChange{NodeID: 1, Type: raftpb.ConfChangeAddNode}.AsV2())
	nodes := r.progress.VoterNodes()
	wnodes := []uint64{1, 2}
	if !reflect.DeepEqual(nodes, wnodes) {
		test.Errorf("nodes = %v, want %v", nodes, wnodes)
	}

	// INFO: (2) Remove node
	r.applyConfChange(raftpb.ConfChange{NodeID: 2, Type: raftpb.ConfChangeRemoveNode}.AsV2())
	w := []uint64{1}
	if g := r.progress.VoterNodes(); !reflect.DeepEqual(g, w) {
		test.Errorf("nodes = %v, want %v", g, w)
	}

	// Removing the remaining voter will panic.
	defer func() {
		if err := recover(); err == nil {
			test.Error("did not panic")
		}
	}()
	r.applyConfChange(raftpb.ConfChange{NodeID: 1, Type: raftpb.ConfChangeRemoveNode}.AsV2())
}

func TestAddLearnerNode(t *testing.T) {
	r := newTestRaft(1, 10, 1, newTestMemoryStorage(withPeers(1)))
	// INFO: (1) Add new learner peer
	r.applyConfChange(raftpb.ConfChange{NodeID: 2, Type: raftpb.ConfChangeAddLearnerNode}.AsV2())
	if r.isLearner {
		// 这里 raft.isLearner 是当前 local node 状态
		t.Fatal("expected 1 to be voter")
	}
	nodes := r.progress.LearnerNodes()
	wantedNodes := []uint64{2}
	if !reflect.DeepEqual(nodes, wantedNodes) {
		t.Errorf("nodes = %v, want %v", nodes, wantedNodes)
	}
	if !r.progress.Progress[2].IsLearner {
		t.Fatal("expected 2 to be learner")
	}
	klog.Infof(fmt.Sprintf("VoterNodes: %+v, LearnerNodes: %+v", r.progress.VoterNodes(), r.progress.LearnerNodes()))

	// INFO: (2) Promote learner to voter
	r.applyConfChange(raftpb.ConfChange{NodeID: 2, Type: raftpb.ConfChangeAddNode}.AsV2())
	if r.progress.Progress[2].IsLearner {
		t.Fatal("expected 2 to be voter")
	}
	if r.isLearner {
		t.Fatal("expected 2 to be voter")
	}
	klog.Infof(fmt.Sprintf("VoterNodes: %+v, LearnerNodes: %+v", r.progress.VoterNodes(), r.progress.LearnerNodes()))

	// INFO: (3) Demote voter to learner
	r.applyConfChange(raftpb.ConfChange{NodeID: 1, Type: raftpb.ConfChangeAddLearnerNode}.AsV2())
	if !r.progress.Progress[1].IsLearner {
		t.Fatal("expected 1 to be learner")
	}
	if !r.isLearner {
		t.Fatal("expected 1 to be learner")
	}
	klog.Infof(fmt.Sprintf("VoterNodes: %+v, LearnerNodes: %+v", r.progress.VoterNodes(), r.progress.LearnerNodes()))

	// INFO: (4) Promote learner to voter
	r.applyConfChange(raftpb.ConfChange{NodeID: 1, Type: raftpb.ConfChangeAddNode}.AsV2())
	if r.progress.Progress[1].IsLearner {
		t.Fatal("expected 1 to be voter")
	}
	if r.isLearner {
		t.Fatal("expected 1 to be voter")
	}
	klog.Infof(fmt.Sprintf("VoterNodes: %+v, LearnerNodes: %+v", r.progress.VoterNodes(), r.progress.LearnerNodes()))

	// INFO: (5) Remove voter
	r.applyConfChange(raftpb.ConfChange{NodeID: 2, Type: raftpb.ConfChangeRemoveNode}.AsV2())
	if len(r.progress.VoterNodes()) != 1 {
		t.Fatal("expected nodes number to be 1")
	}
	klog.Infof(fmt.Sprintf("VoterNodes: %+v, LearnerNodes: %+v", r.progress.VoterNodes(), r.progress.LearnerNodes()))

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

// INFO:
//  日志复制：https://zhuanlan.zhihu.com/p/378025106

func TestRaftFlowControl(test *testing.T) {
	r := newTestRaft(1, 10, 1, newTestMemoryStorage(withPeers(1, 2)))
	r.becomeCandidate()
	r.becomeLeader()

	pr2 := r.progress.Progress[2]
	// force the progress to be in replicate state
	pr2.BecomeReplicate()
	for i := 0; i < r.progress.MaxInflight; i++ {
		r.Step(raftpb.Message{From: 1, To: 1, Type: raftpb.MsgProp, Entries: []raftpb.Entry{{Data: []byte("somedata")}}})
		message := r.readMessages()
		if len(message) != 1 {
			test.Fatalf("#%d: len(ms) = %d, want 1", i, len(message))
		}

		klog.Infof(fmt.Sprintf("%+v", message))
	}

	// ensure 1
	if !pr2.Inflights.Full() {
		test.Fatalf("inflights.full = %t, want %t", pr2.Inflights.Full(), true)
	}

	// ensure 2
	for i := 0; i < 10; i++ {
		r.Step(raftpb.Message{From: 1, To: 1, Type: raftpb.MsgProp, Entries: []raftpb.Entry{{Data: []byte("somedata")}}})
		ms := r.readMessages()
		if len(ms) != 0 { // INFO: 这里 len == 0
			test.Fatalf("#%d: len(ms) = %d, want 0", i, len(ms))
		}
	}
}
