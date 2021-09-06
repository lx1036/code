package raftstore

import (
	"fmt"
	"testing"

	"github.com/tiglabs/raft"
	raftproto "github.com/tiglabs/raft/proto"
	"k8s.io/klog/v2"
)

type MetaPartition struct {
}

func (partition *MetaPartition) Apply(command []byte, index uint64) (interface{}, error) {
	klog.Info(command, index)

	return nil, nil
}

func (partition *MetaPartition) ApplyMemberChange(confChange *raftproto.ConfChange, index uint64) (interface{}, error) {
	panic("implement me")
}

func (partition *MetaPartition) Snapshot() (raftproto.Snapshot, error) {
	panic("implement me")
}

func (partition *MetaPartition) ApplySnapshot(peers []raftproto.Peer, iter raftproto.SnapIterator) error {
	panic("implement me")
}

func (partition *MetaPartition) HandleFatalEvent(err *raft.FatalError) {
	klog.Errorf(fmt.Sprintf("ID: %d, err %s", err.ID, err.Err.Error()))
}

func (partition *MetaPartition) HandleLeaderChange(leader uint64) {
	panic("implement me")
}

func (partition *MetaPartition) Put(key, val interface{}) (interface{}, error) {
	panic("implement me")
}

func (partition *MetaPartition) Get(key interface{}) (interface{}, error) {
	panic("implement me")
}

func (partition *MetaPartition) Del(key interface{}) (interface{}, error) {
	panic("implement me")
}

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

	partition := &MetaPartition{}
	p, err := raftstore.CreatePartition(&PartitionConfig{
		ID:      1,
		Applied: 0,
		Leader:  3,
		Term:    0,
		Peers: []PeerAddress{
			{Peer: raftproto.Peer{ID: uint64(1)}, Address: "127.0.0.1:9021"},
			{Peer: raftproto.Peer{ID: uint64(2)}, Address: "127.0.0.1:9022"},
			{Peer: raftproto.Peer{ID: uint64(3)}, Address: "127.0.0.1:9023"},
		},
		SM: partition,
	})
	if err != nil {
		klog.Fatal(err)
	}

	if p.IsRaftLeader() {
		klog.Info("success")
	} else {
		klog.Info("fail")
	}
}
