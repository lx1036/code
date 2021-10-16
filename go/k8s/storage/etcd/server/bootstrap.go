package server

import (
	"fmt"
	"k8s.io/klog/v2"
	"time"

	"k8s-lx1036/k8s/storage/etcd/raft"
	"k8s-lx1036/k8s/storage/etcd/storage/wal"

	"go.etcd.io/etcd/api/v3/etcdserverpb"
	"go.etcd.io/etcd/pkg/v3/pbutil"
)

const (
	NodeID = 1

	ClusterID = 1

	TickMS = 1000

	DataDIR = "tmp"
)

type bootstrappedRaft struct {
	heartbeat time.Duration

	peers  []raft.Peer
	config *raft.Config

	storage *raft.MemoryStorage
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

	return &bootstrappedRaft{
		heartbeat: time.Duration(TickMS) * time.Millisecond,
		config:    raftConfig(NodeID, s),
		storage:   s,
	}
}

func (b *bootstrappedRaft) newRaftNode(ss *snap.Snapshotter, wal *wal.WAL, cl *membership.RaftCluster) *RaftNode {
	var n raft.Node
	if len(b.peers) == 0 {
		n = raft.RestartNode(b.config)
	} else {
		n = raft.StartNode(b.config, b.peers)
	}
	raftStatusMu.Lock()
	raftStatus = n.Status
	raftStatusMu.Unlock()
	return newRaftNode(
		RaftNodeConfig{
			isIDRemoved: func(id uint64) bool { return cl.IsIDRemoved(types.ID(id)) },
			Node:        n,
			heartbeat:   b.heartbeat,
			raftStorage: b.storage,
			storage:     NewStorage(wal),
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
