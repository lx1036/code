package server

import (
	"fmt"
	"time"

	"k8s-lx1036/k8s/storage/etcd/raft"
	"k8s-lx1036/k8s/storage/etcd/storage/wal"

	"go.etcd.io/etcd/api/v3/etcdserverpb"
	"go.etcd.io/etcd/pkg/v3/pbutil"
	"go.etcd.io/etcd/server/v3/etcdserver/api/snap"
	"k8s.io/klog/v2"
)

const (
	NodeID = 1

	ClusterID = 1

	TickMS = 1000

	DataDIR = "tmp"
	SnapDIR = "tmp/snap"
)

type bootstrappedRaft struct {
	heartbeat time.Duration

	peers  []raft.Peer
	config *raft.Config

	storage     *raft.MemoryStorage
	wal         *wal.WAL
	snapShotter *snap.Snapshotter
}

func bootstrapFromWAL() *bootstrappedRaft {

	s := raft.NewMemoryStorage()

	metadata := pbutil.MustMarshal(
		&etcdserverpb.Metadata{
			NodeID:    uint64(NodeID),
			ClusterID: uint64(ClusterID),
		},
	)
	w, err := wal.Create(DataDIR, metadata)
	if err != nil {
		klog.Fatalf(fmt.Sprintf("[bootstrapFromWAL]failed to create WAL err: %v", err))
	}

	snapShotter := snap.New(nil, SnapDIR)

	return &bootstrappedRaft{
		heartbeat:   time.Duration(TickMS) * time.Millisecond,
		config:      raftConfig(NodeID, s),
		storage:     s,
		wal:         w,
		snapShotter: snapShotter,
	}
}

func (b *bootstrappedRaft) newRaftNode() *RaftNode {
	var node raft.Node
	if len(b.peers) == 0 {
		node = raft.RestartNode(b.config)
	} else {
		node = raft.StartNode(b.config, b.peers)
	}

	return newRaftNode(
		RaftNodeConfig{
			Node:        node,
			heartbeat:   b.heartbeat,
			raftStorage: b.storage,
			storage:     NewStorage(b.wal, b.snapShotter),
		},
	)
}

func raftConfig(id uint64, storage *raft.MemoryStorage) *raft.Config {
	return &raft.Config{
		ID:              id,
		ElectionTick:    10,
		HeartbeatTick:   1,
		Storage:         storage,
		MaxSizePerMsg:   maxSizePerMsg,
		MaxInflightMsgs: maxInflightMsgs,
		CheckQuorum:     true,
		PreVote:         true,
	}
}
