package raftstore

import (
	"fmt"
	"testing"

	"k8s-lx1036/k8s/storage/raft/proto"

	"k8s.io/klog/v2"
)

type MetaPartition struct {
}

func (partition *MetaPartition) Apply(command []byte, index uint64) (interface{}, error) {
	klog.Info(command, index)

	return nil, nil
}

func (partition *MetaPartition) ApplyMemberChange(confChange *proto.ConfChange, index uint64) (interface{}, error) {
	panic("implement me")
}

func (partition *MetaPartition) Snapshot() (proto.Snapshot, error) {
	panic("implement me")
}

func (partition *MetaPartition) ApplySnapshot(peers []proto.Peer, iter proto.SnapIterator) error {
	panic("implement me")
}

func (partition *MetaPartition) HandleFatalEvent(err error) {
	klog.Errorf(fmt.Sprintf("err %v", err))
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
		TickInterval:      5000, // 5000ms
		ElectionTick:      0,
	}

	n, err := NewRaftNode(config)
	if err != nil {
		klog.Fatal(err)
	}

	sm := &MetaPartition{}
	r, err := n.CreatePartition(&PartitionConfig{
		ID:      1,
		Applied: 0,
		Leader:  3,
		Term:    0,
		Peers: []PeerAddress{
			{Peer: proto.Peer{ID: uint64(1)}, Address: "127.0.0.1:9021"},
			{Peer: proto.Peer{ID: uint64(2)}, Address: "127.0.0.1:9022"},
			{Peer: proto.Peer{ID: uint64(3)}, Address: "127.0.0.1:9023"},
		},
		SM: sm,
	})
	if err != nil {
		klog.Fatal(err)
	}

	if r.IsRaftLeader() {
		klog.Info("success")
	} else {
		klog.Info("fail")
	}
}
