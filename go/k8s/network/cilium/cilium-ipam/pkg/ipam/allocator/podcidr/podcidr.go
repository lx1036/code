package podcidr

import (
	v2 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2"
	"github.com/cilium/cilium/pkg/revert"
	log "github.com/sirupsen/logrus"
	"sync"
)

// NodesPodCIDRManager will be used to manage podCIDRs for the nodes in the
// cluster.
type NodesPodCIDRManager struct {
	sync.Mutex

	// v4CIDRAllocators contains the CIDRs for IPv4 addresses
	v4CIDRAllocators []CIDRAllocator
}

// Update will re-allocate the node podCIDRs. In case the node already has
// podCIDRs allocated, the podCIDR allocator will try to allocate those CIDRs.
// In case the CIDRs were able to be allocated, the CiliumNode will have its
// podCIDRs fields set with the allocated CIDRs.
// In case the CIDRs were unable to be allocated, this function will return
// true and the node will have its status updated into kubernetes with the
// error message by the NodesPodCIDRManager.
func (n *NodesPodCIDRManager) Update(node *v2.CiliumNode) bool {
	n.Mutex.Lock()
	defer n.Mutex.Unlock()
	return n.update(node)
}

// Needs n.Mutex to be held.
func (n *NodesPodCIDRManager) update(node *v2.CiliumNode) bool {
	cn, allocated, updateStatus, err := n.allocateNode(node)
	if err != nil {
		return false
	}

	// TODO

}

// AllocateNode allocates the podCIDRs for the given node. Returns a DeepCopied
// node with the podCIDRs allocated. In case there weren't CIDRs allocated
// the returned node will be nil.
// If allocated returns false, it means an update of CiliumNode Status should
// be performed into kubernetes as an error have happened while trying to
// allocate a CIDR for this node.
// Needs n.Mutex to be held.
func (n *NodesPodCIDRManager) allocateNode(node *v2.CiliumNode) (cn *v2.CiliumNode, allocated, updateStatus bool, err error) {

	if len(node.Spec.IPAM.PodCIDRs) == 0 {
		// If we can't allocate podCIDRs for now we should store the node
		// temporarily until n.reSync is called.
		if !n.canAllocatePodCIDRs {
			log.Debug("Postponing CIDR allocation")
			n.nodesToAllocate[node.GetName()] = node
			return nil, false, false, nil
		}

		// Allocate the next free CIDRs
		cidrs, allocated, err = n.allocateNext(node.GetName())
		if err != nil {
			// We want to log this error in cilium node
			updateStatus = true
			return
		}

		log.WithFields(logrus.Fields{
			"cidrs":     cidrs.String(),
			"allocated": allocated,
		}).Debug("Allocated new CIDRs")
	} else {

	}

}

// allocateNext returns the next v4 and / or v6 CIDR available in the CIDR
// allocator. The CIDRs are only allocated if the respective CIDR allocators
// are available. If the node had a CIDR previously allocated the same CIDR
// allocated to that node is returned.
// The return value 'allocated' is set to false in case none of the CIDRs were
// re-allocated, for example in the case the node had already allocated CIDRs.
// In case an error is returned no CIDRs were allocated.
// Needs n.Mutex to be held.
func (n *NodesPodCIDRManager) allocateNext(nodeName string) (*nodeCIDRs, bool, error) {
	// Only allocate a v4 CIDR if the v4CIDR allocator is available
	if len(n.v4CIDRAllocators) != 0 {
		revertFunc, v4CIDR, err = allocateFirstFreeCIDR(n.v4CIDRAllocators)
		if err != nil {
			return nil, false, err
		}

		log.WithField("CIDR", v4CIDR).Debug("v4 allocated CIDR")
		cidrs.v4PodCIDRs = []*net.IPNet{v4CIDR}

		revertStack.Push(revertFunc)
	}

}

// allocateFirstFreeCIDR allocates the first CIDR available from the slice of
// cidrAllocators.
func allocateFirstFreeCIDR(cidrAllocators []CIDRAllocator) (revertFunc revert.RevertFunc, cidr *net.IPNet, err error) {
	var (
		firstFreeAllocator *CIDRAllocator
		revertStack        revert.RevertStack
	)
	for _, cidrAllocator := range cidrAllocators {
		// Allocate from the first allocator that still has free CIDRs
		if !cidrAllocator.IsFull() {
			firstFreeAllocator = &cidrAllocator
			break
		}
	}
	if firstFreeAllocator == nil {
		return nil, nil, &ErrAllocatorFull{}
	}
	cidr, err = (*firstFreeAllocator).AllocateNext()
	if err != nil {
		return nil, nil, err
	}
	revertStack.Push(func() error {
		return (*firstFreeAllocator).Release(cidr)
	})
	return revertStack.Revert, cidr, err
}
