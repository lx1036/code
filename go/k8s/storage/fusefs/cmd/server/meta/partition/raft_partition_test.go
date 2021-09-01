package partition

import (
	"encoding/json"
	"strings"
	"testing"

	"k8s-lx1036/k8s/storage/fusefs/pkg/raftstore"

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

// MetaItem defines the structure of the metadata operations.
type MetaItem struct {
	Op uint32 `json:"op"`
	K  []byte `json:"k"`
	V  []byte `json:"v"`
}

const (
	opFSMCreateInode uint32 = iota
)

// NewMetaItem returns a new MetaItem.
func NewMetaItem(op uint32, key, value []byte) *MetaItem {
	return &MetaItem{
		Op: op,
		K:  key,
		V:  value,
	}
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
			Peer: raftproto.Peer{
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

	metaItem := NewMetaItem(opFSMCreateInode, []byte("hello"), []byte("world"))
	cmd, _ := json.Marshal(metaItem)
	response, err := raftPartition.Submit(cmd)
	if err != nil {
		klog.Fatal(err)
	}

	klog.Info(response)
}
