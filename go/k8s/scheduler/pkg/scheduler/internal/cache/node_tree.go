package cache

import (
	"k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	utilnode "k8s.io/kubernetes/pkg/util/node"
)

// nodeTree is a tree-like data structure that holds node names in each zone. Zone names are
// keys to "NodeTree.tree" and values of "NodeTree.tree" are arrays of node names.
// NodeTree is NOT thread-safe, any concurrent updates/reads from it must be synchronized by the caller.
// It is used only by schedulerCache, and should stay as such.
type nodeTree struct {
	tree      map[string]*nodeArray // a map from zone (region-zone) to an array of nodes in the zone.
	zones     []string              // a list of all the zones in the tree (keys)
	zoneIndex int
	numNodes  int
}

// nodeArray is a struct that has nodes that are in a zone.
// We use a slice (as opposed to a set/map) to store the nodes because iterating over the nodes is
// a lot more frequent than searching them by name.
type nodeArray struct {
	nodes     []string
	lastIndex int
}

// addNode adds a node and its corresponding zone to the tree. If the zone already exists, the node
// is added to the array of nodes in that zone.
func (nt *nodeTree) addNode(n *v1.Node) {
	zone := utilnode.GetZoneKey(n)
	if na, ok := nt.tree[zone]; ok {
		for _, nodeName := range na.nodes {
			if nodeName == n.Name {
				klog.Warningf("node %q already exist in the NodeTree", n.Name)
				return
			}
		}
		na.nodes = append(na.nodes, n.Name)
	} else {
		nt.zones = append(nt.zones, zone)
		nt.tree[zone] = &nodeArray{nodes: []string{n.Name}, lastIndex: 0}
	}
	klog.V(2).Infof("Added node %q in group %q to NodeTree", n.Name, zone)
	nt.numNodes++
}

// newNodeTree creates a NodeTree from nodes.
func newNodeTree(nodes []*v1.Node) *nodeTree {
	nt := &nodeTree{
		tree: make(map[string]*nodeArray),
	}
	for _, n := range nodes {
		nt.addNode(n)
	}
	return nt
}
