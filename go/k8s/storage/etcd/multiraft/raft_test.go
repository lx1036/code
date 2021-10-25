package multiraft

import (
	"testing"
	"time"

	"k8s-lx1036/k8s/storage/raft/proto"

	"k8s.io/klog/v2"
)

func TestRaft(test *testing.T) {
	nodeConfig := DefaultConfig()
	nodeConfig.NodeID = 1
	err := nodeConfig.validate()
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
	memoryStorage := DefaultMemoryStorage()
	raftConfig := &RaftConfig{
		ID:           1,
		Term:         1,
		Leader:       1,
		Applied:      0,
		Peers:        peers,
		Storage:      memoryStorage,
		StateMachine: nil,
	}
	r, err := newRaft(nodeConfig, raftConfig)
	if err != nil {
		klog.Fatal(err)
	}

	ticker := time.Tick(time.Second * 5)
	stopc := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker:
				r.propose([]byte("foo=bar"))
			case <-stopc:
				return
			}
		}
	}()

	<-stopc
}
