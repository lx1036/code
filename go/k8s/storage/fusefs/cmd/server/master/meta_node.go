package master

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"sync"
	"time"

	"k8s-lx1036/k8s/storage/fusefs/pkg/proto"

	"k8s.io/klog/v2"
)

const (
	defaultMetaNodeReservedMem uint64 = 1 << 30 // 1GB

	defaultMetaPartitionMemUsageThreshold float32 = 0.75 // memory usage threshold on a meta partition
)

type NodeSets struct {
	setIndexForMetaNode int
	metaNodes           sync.Map
	nodeSetMap          map[uint64]*nodeSet
	nsLock              sync.RWMutex
}

func newNodeSets() *NodeSets {
	return &NodeSets{
		nodeSetMap: make(map[uint64]*nodeSet),
	}
}

func (t *NodeSets) getAllNodeSet() (nsc nodeSetCollection) {
	t.nsLock.RLock()
	defer t.nsLock.RUnlock()

	nsc = make(nodeSetCollection, 0)
	for _, ns := range t.nodeSetMap {
		nsc = append(nsc, ns)
	}
	return
}

func (t *NodeSets) getAvailNodeSetForMetaNode() *nodeSet {
	allNodeSet := t.getAllNodeSet()
	sort.Sort(allNodeSet) // 快速排序
	for _, ns := range allNodeSet {
		if ns.MetaNodeLen < ns.Capacity {
			return ns
		}
	}

	return nil
}

func (t *NodeSets) putNodeSet(ns *nodeSet) {
	t.nsLock.Lock()
	defer t.nsLock.Unlock()

	t.nodeSetMap[ns.ID] = ns
}

func (t *NodeSets) allocNodeSetForMetaNode(excludeNodeSet *nodeSet, replicaNum uint8) (ns *nodeSet, err error) {
	t.nsLock.RLock()
	defer t.nsLock.RUnlock()

	nset := t.getAllNodeSet()
	if nset.Len() == 0 {
		return nil, fmt.Errorf("no node set available for creating a meta partition")
	}

	for i := 0; i < len(nset); i++ {
		if t.setIndexForMetaNode >= len(nset) {
			t.setIndexForMetaNode = 0
		}
		ns = nset[t.setIndexForMetaNode]
		t.setIndexForMetaNode++
		if excludeNodeSet != nil && ns.ID == excludeNodeSet.ID {
			continue
		}
		if ns.canWriteForMetaNode(int(replicaNum)) {
			return
		}
	}

	return nil, fmt.Errorf("no node set available for creating a meta partition")
}

type nodeSetCollection []*nodeSet

func (nsc nodeSetCollection) Len() int {
	return len(nsc)
}

func (nsc nodeSetCollection) Less(i, j int) bool {
	return nsc[i].MetaNodeLen < nsc[j].MetaNodeLen
}

func (nsc nodeSetCollection) Swap(i, j int) {
	nsc[i], nsc[j] = nsc[j], nsc[i]
}

type nodeSet struct {
	sync.RWMutex

	ID          uint64 `json:"id"`
	Capacity    int    `json:"capacity"`
	MetaNodeLen int    `json:"metaNodeLen"`
	metaNodes   sync.Map
}

func newNodeSet(id uint64, cap int) *nodeSet {
	return &nodeSet{
		ID:       id,
		Capacity: cap,
	}
}

func (ns *nodeSet) putMetaNode(metaNode *MetaNode) {
	ns.metaNodes.Store(metaNode.Addr, metaNode)
}

func (ns *nodeSet) deleteMetaNode(metaNode *MetaNode) {
	ns.metaNodes.Delete(metaNode.Addr)
}

func (ns *nodeSet) increaseMetaNodeLen() {
	ns.Lock()
	defer ns.Unlock()
	ns.MetaNodeLen++
}

func (ns *nodeSet) decreaseMetaNodeLen() {
	ns.Lock()
	defer ns.Unlock()
	ns.MetaNodeLen--
}

func (ns *nodeSet) canWriteForMetaNode(replicaNum int) bool {
	if ns.MetaNodeLen < replicaNum {
		return false
	}
	var count int
	ns.metaNodes.Range(func(key, value interface{}) bool {
		node := value.(*MetaNode)
		if node.isWritable() {
			count++
		}
		if count >= replicaNum {
			return false
		}
		return true
	})
	klog.Infof(fmt.Sprintf("canWriteForMetaNode count[%v] replicaNum[%v]", count, replicaNum))
	return count >= replicaNum
}

func (ns *nodeSet) getAvailMetaNodeHosts(excludeHosts []string, replicaNum int) (newHosts []string, peers []proto.Peer, err error) {
	if replicaNum == 0 {
		return
	}

	maxTotal := ns.getMetaNodeMaxTotal()
	nodes, count := ns.getAllCarryNodes(maxTotal, excludeHosts)
	if len(nodes) < replicaNum {
		err = fmt.Errorf(fmt.Sprintf("ActiveNodeCount:%v  MatchNodeCount:%v", ns.MetaNodeLen, len(nodes)))
		return
	}

	var orderHosts []string
	nodes.setNodeCarry(count, replicaNum)
	sort.Sort(nodes)
	for i := 0; i < replicaNum; i++ {
		node := nodes[i].Ptr.(*MetaNode)
		node.SelectNodeForWrite()
		orderHosts = append(orderHosts, node.Addr)
		peer := proto.Peer{ID: node.ID, Addr: node.Addr}
		peers = append(peers, peer)
	}
	if newHosts, err = reshuffleHosts(orderHosts); err != nil {
		err = fmt.Errorf("getAvailMetaNodeHosts err:%v  orderHosts is nil", err)
		return
	}
	return
}

func (ns *nodeSet) getMetaNodeMaxTotal() (maxTotal uint64) {
	ns.metaNodes.Range(func(key, value interface{}) bool {
		metaNode := value.(*MetaNode)
		if metaNode.Total > maxTotal {
			maxTotal = metaNode.Total
		}
		return true
	})
	return
}

func (ns *nodeSet) getAllCarryNodes(maxTotal uint64, excludeHosts []string) (nodes SortedWeightedNodes, availCount int) {
	nodes = make(SortedWeightedNodes, 0)
	ns.metaNodes.Range(func(key, value interface{}) bool {
		metaNode := value.(*MetaNode)
		if contains(excludeHosts, metaNode.Addr) == true {
			return true
		}
		if metaNode.isWritable() == false {
			return true
		}
		if metaNode.isCarryNode() == true {
			availCount++
		}
		nt := new(weightedNode)
		nt.Carry = metaNode.Carry
		if metaNode.Used < 0 {
			nt.Weight = 1.0
		} else {
			nt.Weight = (float64)(maxTotal-metaNode.Used) / (float64)(maxTotal)
		}
		nt.Ptr = metaNode
		nodes = append(nodes, nt)

		return true
	})

	return
}

// key=#nodeset#
func (cluster *Cluster) submitNodeSet(opType uint32, nset *nodeSet) error {
	cmd := new(RaftCmd)
	cmd.Op = opType
	cmd.K = fmt.Sprintf("%s%s", nodeSetPrefix, strconv.FormatUint(nset.ID, 10))
	cmd.V, _ = json.Marshal(nset)
	return cluster.submit(cmd)
}

type Node interface {
	SetCarry(carry float64)
	SelectNodeForWrite()
}

type weightedNode struct {
	Carry  float64
	Weight float64
	Ptr    Node
	ID     uint64
}

// SortedWeightedNodes defines an array sorted by carry
type SortedWeightedNodes []*weightedNode

func (nodes SortedWeightedNodes) Len() int {
	return len(nodes)
}

func (nodes SortedWeightedNodes) Less(i, j int) bool {
	return nodes[i].Carry > nodes[j].Carry
}

func (nodes SortedWeightedNodes) Swap(i, j int) {
	nodes[i], nodes[j] = nodes[j], nodes[i]
}

func (nodes SortedWeightedNodes) setNodeCarry(availCarryCount, replicaNum int) {
	if availCarryCount >= replicaNum {
		return
	}
	for availCarryCount < replicaNum {
		availCarryCount = 0
		for _, nt := range nodes {
			carry := nt.Carry + nt.Weight
			if carry > 10.0 {
				carry = 10.0
			}
			nt.Carry = carry
			nt.Ptr.SetCarry(carry)
			if carry > 1.0 {
				availCarryCount++
			}
		}
	}
}

func contains(arr []string, element string) (ok bool) {
	if arr == nil || len(arr) == 0 {
		return
	}

	for _, e := range arr {
		if e == element {
			ok = true
			break
		}
	}
	return
}

func reshuffleHosts(oldHosts []string) (newHosts []string, err error) {
	if oldHosts == nil || len(oldHosts) == 0 {
		klog.Errorf(fmt.Sprintf("action[reshuffleHosts],err:%v", proto.ErrReshuffleArray))
		err = proto.ErrReshuffleArray
		return
	}

	lenOldHosts := len(oldHosts)
	newHosts = make([]string, lenOldHosts)
	if lenOldHosts == 1 {
		copy(newHosts, oldHosts)
		return
	}

	for i := lenOldHosts; i > 1; i-- {
		rand.Seed(time.Now().UnixNano())
		oCurrPos := rand.Intn(i)
		oldHosts[i-1], oldHosts[oCurrPos] = oldHosts[oCurrPos], oldHosts[i-1]
	}
	copy(newHosts, oldHosts)
	return
}

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
	return proto.NewAdminTask(proto.OpMetaNodeHeartbeat, metaNode.Addr, &proto.HeartBeatRequest{
		CurrTime:   time.Now().Unix(),
		MasterAddr: masterAddr,
	})
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

// key=#metanode#
func (cluster *Cluster) submitMetaNode(opType uint32, metaNode *MetaNode) error {
	cmd := new(RaftCmd)
	cmd.Op = opType
	cmd.K = fmt.Sprintf("%s%s", metaNodePrefix, strconv.FormatUint(metaNode.ID, 10))
	cmd.V, _ = json.Marshal(metaNode)
	return cluster.submit(cmd)
}
