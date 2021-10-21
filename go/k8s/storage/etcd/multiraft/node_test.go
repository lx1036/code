package multiraft

import (
	"k8s-lx1036/k8s/storage/etcd/multiraft/storage"
	"k8s-lx1036/k8s/storage/raft/proto"
	"k8s.io/klog/v2"
	"testing"
)

func TestNode(test *testing.T) {

	nodeConfig := DefaultConfig()
	nodeConfig.NodeID = 1
	node, err := NewNode(nodeConfig)
	if err != nil {
		klog.Fatal(err)
	}

	// INFO: 该 node 上有两个 partition raft
	peers := []proto.Peer{
		{
			ID:   1,
			Type: proto.PeerNormal,
		},
		{
			ID:   2,
			Type: proto.PeerNormal,
		},
	}
	memoryStorage := storage.DefaultMemoryStorage()
	raftConfig := &RaftConfig{
		ID:           1,
		Term:         1,
		Leader:       1,
		Applied:      0,
		Peers:        peers,
		Storage:      memoryStorage,
		StateMachine: nil,
	}
	err = node.CreateRaft(raftConfig)
	if err != nil {
		klog.Fatal(err)
	}
	raftConfig2 := &RaftConfig{
		ID:           2,
		Term:         1,
		Leader:       1,
		Applied:      0,
		Peers:        peers,
		Storage:      memoryStorage,
		StateMachine: nil,
	}
	err = node.CreateRaft(raftConfig2)
	if err != nil {
		klog.Fatal(err)
	}

	<-node.stopc
}
