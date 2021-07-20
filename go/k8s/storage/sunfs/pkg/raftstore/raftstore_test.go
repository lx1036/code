package raftstore

import (
	"testing"

	"github.com/tiglabs/raft/proto"

	"k8s.io/klog/v2"
)

func TestRaftStoreCreatePartition(test *testing.T) {
	config := &Config{
		NodeID:            1,
		RaftPath:          "raft",
		IPAddr:            "",
		HeartbeatPort:     0,
		ReplicaPort:       0,
		NumOfLogsToRetain: 0,
		TickInterval:      0,
		ElectionTick:      0,
	}

	raftstore, err := NewRaftStore(config)
	if err != nil {
		klog.Fatal(err)
	}

	p, err := raftstore.CreatePartition(&PartitionConfig{
		ID:      1,
		Applied: 0,
		Leader:  3,
		Term:    0,
		Peers: []PeerAddress{
			{Peer: proto.Peer{ID: uint64(1)}, Address: "127.0.0.1"},
			{Peer: proto.Peer{ID: uint64(2)}, Address: "127.0.0.2"},
			{Peer: proto.Peer{ID: uint64(3)}, Address: "127.0.0.3"},
		},
		SM: nil,
	})
	if err != nil {
		klog.Fatal(err)
	}

	klog.Info(p.IsRaftLeader())
}
