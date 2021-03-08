package master

import (
	"k8s-lx1036/k8s/storage/dfs/pkg/util/proto"
	"math/rand"
	"sync"
	"time"
)

// MetaNode defines the structure of a meta node
type MetaNode struct {
	ID                 uint64
	Addr               string
	IsActive           bool
	Sender             *AdminTaskManager
	RackName           string `json:"Rack"`
	MaxMemAvailWeight  uint64 `json:"MaxMemAvailWeight"`
	Total              uint64 `json:"TotalWeight"`
	Used               uint64 `json:"UsedWeight"`
	Ratio              float64
	SelectCount        uint64
	Carry              float64
	Threshold          float32
	ReportTime         time.Time
	metaPartitionInfos []*proto.MetaPartitionReport
	MetaPartitionCount int
	NodeSetID          uint64
	sync.RWMutex
	PersistenceMetaPartitions []uint64
}

func (metaNode *MetaNode) createHeartbeatTask(masterAddr string) *proto.AdminTask {
	request := &proto.HeartBeatRequest{
		CurrTime:   time.Now().Unix(),
		MasterAddr: masterAddr,
	}

	return proto.NewAdminTask(proto.OpMetaNodeHeartbeat, metaNode.Addr, request)
}

func (metaNode *MetaNode) checkHeartbeat() {
	metaNode.Lock()
	defer metaNode.Unlock()
	if time.Since(metaNode.ReportTime) > time.Second*time.Duration(defaultNodeTimeOutSec) {
		metaNode.IsActive = false
	}
}

func newMetaNode(addr, clusterID string) *MetaNode {
	return &MetaNode{
		Addr:   addr,
		Sender: newAdminTaskManager(addr, clusterID),
		Carry:  rand.Float64(),
	}
}
