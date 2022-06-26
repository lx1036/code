package master

import (
	"encoding/json"
	"fmt"
	"strconv"
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

// #metapartition#{volID}#{partitionID}
func (cluster *Cluster) submitMetaPartition(opType uint32, mp *MetaPartition) error {
	cmd := new(RaftCmd)
	cmd.Op = opType
	cmd.K = fmt.Sprintf("%s%s#%s", metaPartitionPrefix, strconv.FormatUint(mp.volID, 10),
		strconv.FormatUint(mp.PartitionID, 10))
	cmd.V, _ = json.Marshal(mp)
	return cluster.submit(cmd)
}
