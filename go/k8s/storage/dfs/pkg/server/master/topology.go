package master

import (
	"sort"
	"sync"
)

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

type Topology struct {
	setIndexForMetaNode int
	metaNodes           sync.Map
	nodeSetMap          map[uint64]*nodeSet
	nsLock              sync.RWMutex
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

func newTopology() *Topology {
	return &Topology{
		nodeSetMap: make(map[uint64]*nodeSet),
	}
}

type nodeSet struct {
	ID          uint64
	Capacity    int
	metaNodeLen int
	metaNodes   sync.Map
	sync.RWMutex
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

func newNodeSet(id uint64, cap int) *nodeSet {
	return &nodeSet{
		ID:       id,
		Capacity: cap,
	}
}
