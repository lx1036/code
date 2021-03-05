package raftstore

import (
	"github.com/tiglabs/raft"
)

// RaftStore defines the interface for the raft store.
type RaftStore interface {
	CreatePartition(cfg *PartitionConfig) (Partition, error)
	Stop()
	RaftConfig() *raft.Config
	RaftStatus(raftID uint64) (raftStatus *raft.Status)
	NodeManager
	RaftServer() *raft.RaftServer
}
