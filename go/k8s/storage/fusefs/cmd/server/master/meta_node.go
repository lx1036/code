package master

import (
	"math/rand"
	"sync"
	"time"

	"k8s-lx1036/k8s/storage/fusefs/pkg/proto"
)

const (
	defaultMetaNodeReservedMem uint64 = 1 << 30 // 1GB

	defaultMetaPartitionMemUsageThreshold float32 = 0.75 // memory usage threshold on a meta partition

)

// MetaNode defines the structure of a meta node
type MetaNode struct {
	sync.RWMutex

	ID                uint64 `json:"id"`
	Addr              string `json:"addr"`
	IsActive          bool   `json:"isActive"`
	Sender            *AdminTaskManager
	RackName          string    `json:"rackName"`
	MaxMemAvailWeight uint64    `json:"maxMemAvailWeight"`
	Total             uint64    `json:"total"`
	Used              uint64    `json:"used"`
	Ratio             float64   `json:"ratio"`
	SelectCount       uint64    `json:"selectCount"`
	Carry             float64   `json:"carry"`
	Threshold         float32   `json:"threshold"`
	ReportTime        time.Time `json:"reportTime"`
	//metaPartitionInfos []*proto.MetaPartitionReport
	MetaPartitionCount        int      `json:"metaPartitionCount"`
	NodeSetID                 uint64   `json:"nodeSetID"`
	PersistenceMetaPartitions []uint64 `json:"persistenceMetaPartitions"`
}

func newMetaNode(addr, clusterID string) *MetaNode {
	return &MetaNode{
		Addr:   addr,
		Sender: newAdminTaskManager(addr, clusterID),
		Carry:  rand.Float64(),
	}
}

func (metaNode *MetaNode) SetCarry(carry float64) {
	metaNode.Lock()
	defer metaNode.Unlock()
	metaNode.Carry = carry
}

func (metaNode *MetaNode) SelectNodeForWrite() {
	metaNode.Lock()
	defer metaNode.Unlock()
	metaNode.SelectCount++
	metaNode.Carry = metaNode.Carry - 1.0
}

// A carry node is the meta node whose carry is greater than one.
func (metaNode *MetaNode) isCarryNode() (ok bool) {
	metaNode.RLock()
	defer metaNode.RUnlock()
	return metaNode.Carry >= 1
}

func (metaNode *MetaNode) createHeartbeatTask(masterAddr string) *proto.AdminTask {
	panic("not implemented")

	/*request := &proto.HeartBeatRequest{
		CurrTime:   time.Now().Unix(),
		MasterAddr: masterAddr,
	}

	return proto.NewAdminTask(proto.OpMetaNodeHeartbeat, metaNode.Addr, request)*/
}

func (metaNode *MetaNode) checkHeartbeat() {
	metaNode.Lock()
	defer metaNode.Unlock()
	if time.Since(metaNode.ReportTime) > time.Second*time.Duration(defaultNodeTimeOutSec) {
		metaNode.IsActive = false
	}
}

func (metaNode *MetaNode) isWritable() (ok bool) {
	return metaNode.IsActive && metaNode.MaxMemAvailWeight > defaultMetaNodeReservedMem &&
		!metaNode.reachesThreshold() && metaNode.MetaPartitionCount < defaultMaxMetaPartitionCountOnEachNode
}

func (metaNode *MetaNode) reachesThreshold() bool {
	if metaNode.Threshold <= 0 {
		metaNode.Threshold = defaultMetaPartitionMemUsageThreshold
	}

	return float32(float64(metaNode.Used)/float64(metaNode.Total)) > metaNode.Threshold
}
