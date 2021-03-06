package master

import (
	"github.com/tiglabs/raft/proto"
	"sync"
)

// MetaReplica defines the replica of a meta partition
type MetaReplica struct {
	Addr       string
	start      uint64 // lower bound of the inode id
	end        uint64 // upper bound of the inode id
	nodeID     uint64
	ReportTime int64
	Status     int8 // unavailable, readOnly, readWrite
	IsLeader   bool
	metaNode   *MetaNode
}

// MetaPartition defines the structure of a meta partition
type MetaPartition struct {
	PartitionID  uint64
	Start        uint64
	End          uint64
	MaxInodeID   uint64
	Size         uint64
	Replicas     []*MetaReplica
	ReplicaNum   uint8
	Status       int8
	volID        uint64
	volName      string
	Hosts        []string
	Peers        []proto.Peer
	MissNodes    map[string]int64
	LoadResponse []*proto.MetaPartitionLoadResponse
	sync.RWMutex
}
