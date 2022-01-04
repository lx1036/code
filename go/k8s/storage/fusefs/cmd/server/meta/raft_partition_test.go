package meta

import (
	"encoding/json"
	"net"
	"strings"
	"testing"

	//"k8s-lx1036/k8s/storage/fusefs/pkg/raftstore"
	"k8s-lx1036/k8s/storage/fusefs/cmd/server/meta/partition/raftstore"

	"github.com/tiglabs/raft"
	"github.com/tiglabs/raft/proto"
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

func (partition *MetaPartition) HandleFatalEvent(err *raft.FatalError) {
	panic("implement me")
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

func TestRaftPartition(test *testing.T) {
	nodeId := uint64(1)
	raftDir := "data/meta/raft"
	localAddr := "127.0.0.1"
	raftHeartbeatPort := 9093
	raftReplicaPort := 9094
	raftConf := &raftstore.Config{
		NodeID:            nodeId,
		RaftPath:          raftDir,
		IPAddr:            localAddr,
		HeartbeatPort:     raftHeartbeatPort,
		ReplicaPort:       raftReplicaPort,
		NumOfLogsToRetain: 1000000,
	}
	raftStore, err := raftstore.NewRaftStore(raftConf)
	if err != nil {
		klog.Fatal(err)
	}

	applyID := uint64(1)
	partitionId := uint64(1)
	var peersAddress []raftstore.PeerAddress
	type Peer struct {
		ID   uint64
		Addr string
	}
	peers := []Peer{
		{ID: 0, Addr: "127.0.0.1:9021"},
		{ID: 1, Addr: "127.0.0.1:9022"},
		{ID: 2, Addr: "127.0.0.1:9023"},
	}
	for _, peer := range peers {
		addr := strings.Split(peer.Addr, ":")[0]
		peerAddress := raftstore.PeerAddress{
			Peer: proto.Peer{
				ID: peer.ID,
			},
			Address:       addr,
			HeartbeatPort: raftHeartbeatPort,
			ReplicaPort:   raftReplicaPort,
		}
		peersAddress = append(peersAddress, peerAddress)
	}
	partition := &MetaPartition{}
	partitionConfig := &raftstore.PartitionConfig{
		ID:      partitionId,
		Applied: applyID,
		Peers:   peersAddress,
		SM:      partition,
	}
	raftPartition, err := raftStore.CreatePartition(partitionConfig)
	if err != nil {
		klog.Fatal(err)
	}

	cmd, _ := json.Marshal(&RaftCmd{
		Op: opFSMCreateInode,
		K:  "hello",
		V:  []byte("world"),
	})
	response, err := raftPartition.Submit(cmd)
	if err != nil {
		klog.Fatal(err)
	}

	klog.Info(response)
}

func TestNetDial(test *testing.T) {
	_, err := net.Dial("tcp", "127.0.0.1:8581")
	if err != nil {
		klog.Error(err) // "dial tcp 127.0.0.1:8581: connect: connection refused"
	}
}
