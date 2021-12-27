package master

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	"k8s-lx1036/k8s/storage/fusefs/pkg/proto"

	"k8s.io/klog/v2"
)

type Topology struct {
	setIndexForMetaNode int
	metaNodes           sync.Map
	nodeSetMap          map[uint64]*nodeSet
	nsLock              sync.RWMutex
}

func newTopology() *Topology {
	return &Topology{
		nodeSetMap: make(map[uint64]*nodeSet),
	}
}

func (t *Topology) getAllNodeSet() (nsc nodeSetCollection) {
	t.nsLock.RLock()
	defer t.nsLock.RUnlock()

	nsc = make(nodeSetCollection, 0)
	for _, ns := range t.nodeSetMap {
		nsc = append(nsc, ns)
	}
	return
}

func (t *Topology) getAvailNodeSetForMetaNode() *nodeSet {
	allNodeSet := t.getAllNodeSet()
	sort.Sort(allNodeSet) // 快速排序
	for _, ns := range allNodeSet {
		if ns.metaNodeLen < ns.Capacity {
			return ns
		}
	}

	return nil
}

func (t *Topology) putNodeSet(ns *nodeSet) {
	t.nsLock.Lock()
	defer t.nsLock.Unlock()

	t.nodeSetMap[ns.ID] = ns
}

func (t *Topology) allocNodeSetForMetaNode(excludeNodeSet *nodeSet, replicaNum uint8) (ns *nodeSet, err error) {
	t.nsLock.RLock()
	defer t.nsLock.RUnlock()

	nset := t.getAllNodeSet()
	if nset == nil {
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
	return nsc[i].metaNodeLen < nsc[j].metaNodeLen
}

func (nsc nodeSetCollection) Swap(i, j int) {
	nsc[i], nsc[j] = nsc[j], nsc[i]
}

type nodeSet struct {
	sync.RWMutex

	ID          uint64
	Capacity    int
	metaNodeLen int
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
	ns.metaNodeLen++
}

func (ns *nodeSet) decreaseMetaNodeLen() {
	ns.Lock()
	defer ns.Unlock()
	ns.metaNodeLen--
}

func (ns *nodeSet) canWriteForMetaNode(replicaNum int) bool {
	if ns.metaNodeLen < replicaNum {
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
		err = fmt.Errorf(fmt.Sprintf("ActiveNodeCount:%v  MatchNodeCount:%v", ns.metaNodeLen, len(nodes)))
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
