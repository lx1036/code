package raft

import (
	pb "go.etcd.io/etcd/raft/v3/raftpb"
	"testing"
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

func newTestMemoryStorage(opts ...testMemoryStorageOptions) *MemoryStorage {
	ms := NewMemoryStorage()
	for _, o := range opts {
		o(ms)
	}
	return ms
}

func TestAddLearnerNode(t *testing.T) {
	r := newTestRaft(1, 10, 1, newTestMemoryStorage(withPeers(1)))
	// Add new learner peer.
	r.applyConfChange(pb.ConfChange{NodeID: 2, Type: pb.ConfChangeAddLearnerNode}.AsV2())
	if r.isLearner {
		t.Fatal("expected 1 to be voter")
	}

}
