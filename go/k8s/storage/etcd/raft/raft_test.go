package raft

import (
	"fmt"
	"reflect"
	"testing"

	pb "go.etcd.io/etcd/raft/v3/raftpb"
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
	r.applyConfChange(pb.ConfChange{NodeID: 1, Type: pb.ConfChangeAddNode}.AsV2())
	nodes := r.progress.VoterNodes()
	wnodes := []uint64{1, 2}
	if !reflect.DeepEqual(nodes, wnodes) {
		test.Errorf("nodes = %v, want %v", nodes, wnodes)
	}

	// INFO: (2) Remove node
	r.applyConfChange(pb.ConfChange{NodeID: 2, Type: pb.ConfChangeRemoveNode}.AsV2())
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
	r.applyConfChange(pb.ConfChange{NodeID: 1, Type: pb.ConfChangeRemoveNode}.AsV2())
}

func TestAddLearnerNode(t *testing.T) {
	r := newTestRaft(1, 10, 1, newTestMemoryStorage(withPeers(1)))
	// INFO: (1) Add new learner peer
	r.applyConfChange(pb.ConfChange{NodeID: 2, Type: pb.ConfChangeAddLearnerNode}.AsV2())
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
	r.applyConfChange(pb.ConfChange{NodeID: 2, Type: pb.ConfChangeAddNode}.AsV2())
	if r.progress.Progress[2].IsLearner {
		t.Fatal("expected 2 to be voter")
	}
	if r.isLearner {
		t.Fatal("expected 2 to be voter")
	}
	klog.Infof(fmt.Sprintf("VoterNodes: %+v, LearnerNodes: %+v", r.progress.VoterNodes(), r.progress.LearnerNodes()))

	// INFO: (3) Demote voter to learner
	r.applyConfChange(pb.ConfChange{NodeID: 1, Type: pb.ConfChangeAddLearnerNode}.AsV2())
	if !r.progress.Progress[1].IsLearner {
		t.Fatal("expected 1 to be learner")
	}
	if !r.isLearner {
		t.Fatal("expected 1 to be learner")
	}
	klog.Infof(fmt.Sprintf("VoterNodes: %+v, LearnerNodes: %+v", r.progress.VoterNodes(), r.progress.LearnerNodes()))

	// INFO: (4) Promote learner to voter
	r.applyConfChange(pb.ConfChange{NodeID: 1, Type: pb.ConfChangeAddNode}.AsV2())
	if r.progress.Progress[1].IsLearner {
		t.Fatal("expected 1 to be voter")
	}
	if r.isLearner {
		t.Fatal("expected 1 to be voter")
	}
	klog.Infof(fmt.Sprintf("VoterNodes: %+v, LearnerNodes: %+v", r.progress.VoterNodes(), r.progress.LearnerNodes()))

	// INFO: (5) Remove voter
	r.applyConfChange(pb.ConfChange{NodeID: 2, Type: pb.ConfChangeRemoveNode}.AsV2())
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

func (r *raft) readMessages() []pb.Message {
	msgs := r.msgs
	r.msgs = make([]pb.Message, 0)

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
		r.Step(pb.Message{From: 1, To: 1, Type: pb.MsgProp, Entries: []pb.Entry{{Data: []byte("somedata")}}})
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
		r.Step(pb.Message{From: 1, To: 1, Type: pb.MsgProp, Entries: []pb.Entry{{Data: []byte("somedata")}}})
		ms := r.readMessages()
		if len(ms) != 0 { // INFO: 这里 len == 0
			test.Fatalf("#%d: len(ms) = %d, want 0", i, len(ms))
		}
	}
}

// INFO: (1) leader election
// TestHandleHeartbeat ensures that the follower commits to the commit in the message.
func TestHandleHeartbeat(t *testing.T) {
	commit := uint64(2)
	fixtures := []struct {
		message pb.Message
		wCommit uint64
	}{
		{pb.Message{From: 2, To: 1, Type: pb.MsgHeartbeat, Term: 2, Commit: commit + 1}, commit + 1},
		{pb.Message{From: 2, To: 1, Type: pb.MsgHeartbeat, Term: 2, Commit: commit - 1}, commit}, // do not decrease commit
	}

	for i, fixture := range fixtures {
		storage := newTestMemoryStorage(withPeers(1, 2))
		storage.Append([]pb.Entry{{Index: 1, Term: 1}, {Index: 2, Term: 2}, {Index: 3, Term: 3}})
		r := newTestRaft(1, 5, 1, storage)
		r.becomeFollower(2, 2)
		r.raftLog.commitTo(commit)
		r.handleHeartbeat(fixture.message)
		if r.raftLog.committed != fixture.wCommit {
			t.Errorf("#%d: committed = %d, want %d", i, r.raftLog.committed, fixture.wCommit)
		}
		m := r.readMessages()
		if len(m) != 1 {
			t.Fatalf("#%d: msg = nil, want 1", i)
		}
		if m[0].Type != pb.MsgHeartbeatResp {
			t.Errorf("#%d: type = %v, want MsgHeartbeatResp", i, m[0].Type)
		}

		klog.Infof(fmt.Sprintf("[TestHandleHeartbeat]%+v", m))
	}
}

// INFO: (2) log replication

// INFO: (3) safety

// INFO: (4) snapshot

// INFO: 起始只有 {1,2}，snapshot 里有 {1,2,3}
func TestRestore(test *testing.T) {
	snapshot := pb.Snapshot{
		Metadata: pb.SnapshotMetadata{
			Index:     11, // magic number
			Term:      11, // magic number
			ConfState: pb.ConfState{Voters: []uint64{1, 2, 3}, Learners: []uint64{4}},
		},
	}
	storage := newTestMemoryStorage(withPeers(1, 2), withLearners(4))
	r := newTestRaft(1, 10, 1, storage)
	if !r.restore(snapshot) {
		test.Fatal("restore fail, want succeed")
	}

	if r.raftLog.lastIndex() != snapshot.Metadata.Index {
		test.Errorf("log.lastIndex = %d, want %d", r.raftLog.lastIndex(), snapshot.Metadata.Index)
	}
	if term, err := r.raftLog.term(snapshot.Metadata.Index); err != nil || term != snapshot.Metadata.Term {
		test.Errorf("log.lastTerm = %d, want %d", term, snapshot.Metadata.Term)
	}
	voters := r.progress.VoterNodes()
	if !reflect.DeepEqual(voters, snapshot.Metadata.ConfState.Voters) {
		test.Errorf("sm.Voters = %+v, want %+v", voters, snapshot.Metadata.ConfState.Voters)
	}
	learners := r.progress.LearnerNodes()
	if !reflect.DeepEqual(learners, snapshot.Metadata.ConfState.Learners) {
		test.Errorf("sm.Learners = %+v, length not equal with %+v", learners, snapshot.Metadata.ConfState.Learners)
	}

	if ok := r.restore(snapshot); ok {
		test.Fatal("restore succeed, want fail")
	}
	// It should not campaign before actually applying data.
	for i := 0; i < r.randomizedElectionTimeout; i++ {
		r.tick()
	}
	if r.state != StateFollower {
		test.Errorf("state = %d, want %d", r.state, StateFollower)
	}
}
