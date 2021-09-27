package raft

import (
	"fmt"
	"k8s.io/klog/v2"
	"reflect"
	"testing"

	pb "go.etcd.io/etcd/raft/v3/raftpb"
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
	r.applyConfChange(pb.ConfChange{NodeID: 2, Type: pb.ConfChangeAddLearnerNode}.AsV2())
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
	r.applyConfChange(pb.ConfChange{NodeID: 2, Type: pb.ConfChangeAddNode}.AsV2())
	if r.prs.Progress[2].IsLearner {
		t.Fatal("expected 2 to be voter")
	}
	if r.isLearner {
		t.Fatal("expected 2 to be voter")
	}
	klog.Infof(fmt.Sprintf("VoterNodes: %+v, LearnerNodes: %+v", r.prs.VoterNodes(), r.prs.LearnerNodes()))

	// INFO: (3) Demote voter to learner
	r.applyConfChange(pb.ConfChange{NodeID: 1, Type: pb.ConfChangeAddLearnerNode}.AsV2())
	if !r.prs.Progress[1].IsLearner {
		t.Fatal("expected 1 to be learner")
	}
	if !r.isLearner {
		t.Fatal("expected 1 to be learner")
	}
	klog.Infof(fmt.Sprintf("VoterNodes: %+v, LearnerNodes: %+v", r.prs.VoterNodes(), r.prs.LearnerNodes()))

	// INFO: (4) Promote learner to voter
	r.applyConfChange(pb.ConfChange{NodeID: 1, Type: pb.ConfChangeAddNode}.AsV2())
	if r.prs.Progress[1].IsLearner {
		t.Fatal("expected 1 to be voter")
	}
	if r.isLearner {
		t.Fatal("expected 1 to be voter")
	}
	klog.Infof(fmt.Sprintf("VoterNodes: %+v, LearnerNodes: %+v", r.prs.VoterNodes(), r.prs.LearnerNodes()))

	// INFO: (5) Remove voter
	r.applyConfChange(pb.ConfChange{NodeID: 2, Type: pb.ConfChangeRemoveNode}.AsV2())
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

func (r *raft) readMessages() []pb.Message {
	msgs := r.msgs
	r.msgs = make([]pb.Message, 0)

	return msgs
}

func TestRaftFlowControl(test *testing.T) {
	r := newTestRaft(1, 10, 1, newTestMemoryStorage(withPeers(1)))
	r.becomeCandidate()
	r.becomeLeader()

	// Throw away all the messages relating to the initial election.
	r.readMessages()
}
