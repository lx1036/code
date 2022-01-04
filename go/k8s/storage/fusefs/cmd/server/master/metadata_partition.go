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
	Addr       string    `json:"addr"`
	start      uint64    `json:"start"` // lower bound of the inode id
	end        uint64    `json:"end"`   // upper bound of the inode id
	nodeID     uint64    `json:"nodeID"`
	ReportTime int64     `json:"reportTime"`
	Status     int8      `json:"status"` // unavailable, readOnly, readWrite
	IsLeader   bool      `json:"isLeader"`
	metaNode   *MetaNode `json:"metaNode"`
}

// MetaPartition defines the structure of a meta partition
type MetaPartition struct {
	sync.RWMutex

	PartitionID uint64           `json:"partitionID"`
	Start       uint64           `json:"start"`
	End         uint64           `json:"end"`
	MaxInodeID  uint64           `json:"maxInodeID"`
	Size        uint64           `json:"size"`
	Replicas    []*MetaReplica   `json:"replicas"`
	ReplicaNum  int              `json:"replicaNum"`
	Status      int8             `json:"status"`
	volID       uint64           `json:"volID"`
	volName     string           `json:"volName"`
	Hosts       []string         `json:"hosts"`
	Peers       []proto.Peer     `json:"peers"`
	MissNodes   map[string]int64 `json:"missNodes"`

	InodeCount  uint64
	DentryCount uint64

	//LoadResponse []*proto.MetaPartitionLoadResponse
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
