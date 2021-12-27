package master

import (
	"sync"

	"k8s-lx1036/k8s/storage/fusefs/pkg/proto"
)

// The following defines the status of a disk or a partition.
const (
	ReadOnly    = 1
	ReadWrite   = 2
	Unavailable = -1
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
	PartitionID uint64
	Start       uint64
	End         uint64
	MaxInodeID  uint64
	Size        uint64
	Replicas    []*MetaReplica
	ReplicaNum  int
	Status      int8
	volID       uint64
	volName     string
	Hosts       []string
	Peers       []proto.Peer
	MissNodes   map[string]int64
	//LoadResponse []*proto.MetaPartitionLoadResponse
	sync.RWMutex
}

func newMetaPartition(partitionID, start, end uint64, replicaNum int, volName string, volID uint64, isMarkDeleted bool) (mp *MetaPartition) {
	mp = &MetaPartition{PartitionID: partitionID, Start: start, End: end, volName: volName, volID: volID}
	mp.ReplicaNum = replicaNum
	mp.Replicas = make([]*MetaReplica, 0)
	mp.Status = Unavailable
	mp.MissNodes = make(map[string]int64, 0)
	mp.Peers = make([]proto.Peer, 0)
	mp.Hosts = make([]string, 0)
	//mp.LoadResponse = make([]*proto.MetaPartitionLoadResponse, 0)
	//mp.IsMarkDeleted = isMarkDeleted
	return
}

func (mp *MetaPartition) setPeers(peers []proto.Peer) {
	mp.Peers = peers
}

func (mp *MetaPartition) setHosts(hosts []string) {
	mp.Hosts = hosts
}
