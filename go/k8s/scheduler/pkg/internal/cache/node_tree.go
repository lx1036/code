package cache

import (
	"fmt"
	"k8s.io/api/core/v1"
	utilnode "k8s.io/component-helpers/node/topology"
	"k8s.io/klog/v2"
)

type nodeTree struct {
	tree      map[string][]string // [zone][node1,...]
	zones     []string
	zoneIndex int
	numNodes  int
}

func newNodeTree(nodes []*v1.Node) *nodeTree {
	nt := &nodeTree{
		tree: make(map[string][]string),
	}
	for _, n := range nodes {
		nt.addNode(n)
	}
	return nt
}

// addNode adds a node and its corresponding zone to the tree. If the zone already exists, the node
// is added to the array of nodes in that zone.
func (nt *nodeTree) addNode(node *v1.Node) {
	zone := utilnode.GetZoneKey(node)
	if nodes, ok := nt.tree[zone]; ok {
		for _, nodeName := range nodes {
			if nodeName == node.Name {
				klog.Warningf("node %q already exist in the NodeTree", node.Name)
				return
			}
		}
		nt.tree[zone] = append(nodes, node.Name)
	} else {
		nt.zones = append(nt.zones, zone)
		nt.tree[zone] = []string{node.Name}
	}
	klog.V(2).Infof("Added node %q in group %q to NodeTree", node.Name, zone)
	nt.numNodes++
}

func (nt *nodeTree) updateNode(old, new *v1.Node) {
	var oldZone string
	if old != nil {
		oldZone = utilnode.GetZoneKey(old)
	}
	newZone := utilnode.GetZoneKey(new)
	if oldZone == newZone {
		return
	}

	// 只有 zone 改变了
	nt.removeNode(old)
	nt.addNode(new)
}

func (nt *nodeTree) removeNode(n *v1.Node) error {
	zone := utilnode.GetZoneKey(n)
	if na, ok := nt.tree[zone]; ok {
		for i, nodeName := range na {
			if nodeName == n.Name {
				nt.tree[zone] = append(na[:i], na[i+1:]...)
				if len(nt.tree[zone]) == 0 {
					nt.removeZone(zone)
				}
				klog.V(2).InfoS("Removed node in listed group from NodeTree", "node", klog.KObj(n), "zone", zone)
				nt.numNodes--
				return nil
			}
		}
	}
	klog.ErrorS(nil, "Node in listed group was not found", "node", klog.KObj(n), "zone", zone)
	return fmt.Errorf("node %q in group %q was not found", n.Name, zone)
}

func (nt *nodeTree) removeZone(zone string) {
	delete(nt.tree, zone)
	for i, z := range nt.zones {
		if z == zone {
			nt.zones = append(nt.zones[:i], nt.zones[i+1:]...)
			return
		}
	}
}
