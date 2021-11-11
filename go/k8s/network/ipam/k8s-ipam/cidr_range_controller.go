package k8s_ipam

import (
	"fmt"
	"net"
	"sync"

	"k8s-lx1036/k8s/network/ipam/k8s-ipam/cidrset"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	informers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	nodeutil "k8s.io/kubernetes/pkg/controller/util/node"
	utilnode "k8s.io/kubernetes/pkg/util/node"
)

const (
	cidrUpdateQueueSize = 5000

	cidrUpdateRetries = 3
)

type CIDRAllocatorParams struct {
	// ClusterCIDRs is list of cluster cidrs
	ClusterCIDRs []*net.IPNet
	// ServiceCIDR is primary service cidr for cluster
	ServiceCIDR *net.IPNet
	// SecondaryServiceCIDR is secondary service cidr for cluster
	SecondaryServiceCIDR *net.IPNet
	// NodeCIDRMaskSizes is list of node cidr mask sizes
	NodeCIDRMaskSizes []int
}

type rangeAllocator struct {
	// Keep a set of nodes that are currently being processed to avoid races in CIDR allocation
	lock sync.Mutex

	client      clientset.Interface
	nodeLister  corelisters.NodeLister
	nodesSynced cache.InformerSynced

	nodesInProcessing sets.String

	// 支持多个 cluster cidr
	cidrSets []*cidrset.CidrSet

	clusterCIDRs []*net.IPNet

	nodeCIDRUpdateChannel chan nodeReservedCIDRs
}

func NewCIDRRangeAllocator(client clientset.Interface,
	nodeInformer informers.NodeInformer,
	nodeList *corev1.NodeList,
	allocatorParams CIDRAllocatorParams) (*rangeAllocator, error) {
	cidrSets := make([]*cidrset.CidrSet, len(allocatorParams.ClusterCIDRs))
	for idx, cidr := range allocatorParams.ClusterCIDRs {
		cidrSet, err := cidrset.NewCIDRSet(cidr, allocatorParams.NodeCIDRMaskSizes[idx])
		if err != nil {
			return nil, err
		}
		cidrSets[idx] = cidrSet
	}

	allocator := &rangeAllocator{
		client:                client,
		nodeLister:            nodeInformer.Lister(),
		nodesSynced:           nodeInformer.Informer().HasSynced,
		nodeCIDRUpdateChannel: make(chan nodeReservedCIDRs, cidrUpdateQueueSize),

		nodesInProcessing: sets.NewString(),

		cidrSets:     cidrSets,
		clusterCIDRs: allocatorParams.ClusterCIDRs,
	}

	if allocatorParams.ServiceCIDR != nil {
		allocator.filterOutServiceRange(allocatorParams.ServiceCIDR)
	} else {
		klog.Info("No Service CIDR provided. Skipping filtering out service addresses.")
	}

	// IPAM 启动后，先初始化 node PodCIDRs: occupy 已有的 CIDR
	if nodeList != nil {
		for _, node := range nodeList.Items {
			if len(node.Spec.PodCIDRs) == 0 {
				klog.V(4).Infof("Node %v has no CIDR, ignoring", node.Name)
				continue
			}
			klog.V(4).Infof("Node %v has CIDR %s, occupying it in CIDR map", node.Name, node.Spec.PodCIDR)
			if err := allocator.occupyCIDRs(&node); err != nil {
				// This will happen if:
				// 1. We find garbage in the podCIDRs field. Retrying is useless.
				// 2. CIDR out of range: This means a node CIDR has changed.
				// This error will keep crashing controller-manager.
				return nil, err
			}
		}
	}

	nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: nodeutil.CreateAddNodeHandler(allocator.AllocateOrOccupyCIDR),
		UpdateFunc: nodeutil.CreateUpdateNodeHandler(func(_, newNode *corev1.Node) error {
			// If the PodCIDRs list is not empty we either:
			// - already processed a Node that already had CIDRs after NC restarted
			//   (cidr is marked as used),
			// - already processed a Node successfully and allocated CIDRs for it
			//   (cidr is marked as used),
			// - already processed a Node but we did saw a "timeout" response and
			//   request eventually got through in this case we haven't released
			//   the allocated CIDRs (cidr is still marked as used).
			// There's a possible error here:
			// - NC sees a new Node and assigns CIDRs X,Y.. to it,
			// - Update Node call fails with a timeout,
			// - Node is updated by some other component, NC sees an update and
			//   assigns CIDRs A,B.. to the Node,
			// - Both CIDR X,Y.. and CIDR A,B.. are marked as used in the local cache,
			//   even though Node sees only CIDR A,B..
			// The problem here is that in in-memory cache we see CIDR X,Y.. as marked,
			// which prevents it from being assigned to any new node. The cluster
			// state is correct.
			// Restart of NC fixes the issue.
			if len(newNode.Spec.PodCIDRs) == 0 {
				return allocator.AllocateOrOccupyCIDR(newNode)
			}
			return nil
		}),
		DeleteFunc: nodeutil.CreateDeleteNodeHandler(allocator.ReleaseCIDR),
	})

	return allocator, nil
}

// INFO: 如果 serviceCIDR 和 clusterCIDR 重叠，重叠部分标记为 occupy
func (allocator *rangeAllocator) filterOutServiceRange(serviceCIDR *net.IPNet) {
	for idx, cidr := range allocator.clusterCIDRs {
		// INFO: 如果两个 cidr 不重叠，可以借鉴
		if !cidr.Contains(serviceCIDR.IP.Mask(cidr.Mask)) && !serviceCIDR.Contains(cidr.IP.Mask(serviceCIDR.Mask)) {
			continue
		}

		// at this point, len(cidrSet) == len(clusterCidr)
		if err := allocator.cidrSets[idx].Occupy(serviceCIDR); err != nil {
			klog.Errorf("Error filtering out service cidr out cluster cidr:%v (index:%v) %v: %v", cidr, idx, serviceCIDR, err)
		}
	}
}

// marks node.PodCIDRs[...] as used in allocator's tracked cidrSet
func (allocator *rangeAllocator) occupyCIDRs(node *corev1.Node) error {
	defer allocator.removeNodeFromProcessing(node.Name)
	if len(node.Spec.PodCIDRs) == 0 {
		return nil
	}

	for idx, cidr := range node.Spec.PodCIDRs {
		_, podCIDR, err := net.ParseCIDR(cidr)
		if err != nil {
			return fmt.Errorf("failed to parse node %s, CIDR %s", node.Name, node.Spec.PodCIDR)
		}
		if idx >= len(allocator.cidrSets) {
			return fmt.Errorf("node:%s has an allocated cidr: %v at index:%v that does not exist in cluster cidrs configuration", node.Name, cidr, idx)
		}

		if err := allocator.cidrSets[idx].Occupy(podCIDR); err != nil {
			return fmt.Errorf("failed to mark cidr[%v] at idx [%v] as occupied for node: %v: %v", podCIDR, idx, node.Name, err)
		}
	}
	return nil
}

// ReleaseCIDR marks node.podCIDRs[...] as unused in our tracked cidrSets
func (allocator *rangeAllocator) ReleaseCIDR(node *corev1.Node) error {
	if node == nil || len(node.Spec.PodCIDRs) == 0 {
		return nil
	}

	for idx, cidr := range node.Spec.PodCIDRs {
		_, podCIDR, err := net.ParseCIDR(cidr)
		if err != nil {
			return fmt.Errorf("failed to parse CIDR %s on Node %v: %v", cidr, node.Name, err)
		}

		if idx >= len(allocator.cidrSets) {
			return fmt.Errorf("node:%s has an allocated cidr: %v at index:%v that does not exist in cluster cidrs configuration", node.Name, cidr, idx)
		}

		klog.Infof("release CIDR %s for node:%v", cidr, node.Name)
		if err = allocator.cidrSets[idx].Release(podCIDR); err != nil {
			return fmt.Errorf("error when releasing CIDR %v: %v", cidr, err)
		}
	}

	return nil
}

type nodeReservedCIDRs struct {
	allocatedCIDRs []*net.IPNet
	nodeName       string
}

func (allocator *rangeAllocator) AllocateOrOccupyCIDR(node *corev1.Node) error {
	if node == nil {
		return nil
	}
	if !allocator.insertNodeToProcessing(node.Name) {
		klog.V(2).Infof("Node %v is already in a process of CIDR assignment.", node.Name)
		return nil
	}

	if len(node.Spec.PodCIDRs) > 0 { // 如果 node 在 Create 之前已经有了 PodCIDRs
		return allocator.occupyCIDRs(node)
	}

	// allocate and queue the assignment
	allocated := nodeReservedCIDRs{
		nodeName:       node.Name,
		allocatedCIDRs: make([]*net.IPNet, len(allocator.cidrSets)),
	}

	for idx := range allocator.cidrSets {
		podCIDR, err := allocator.cidrSets[idx].AllocateNext() // INFO: 因为支持多个 clusterCIDR，这里从每一个 CIDR 分别分配一个 podCIDR
		if err != nil {
			allocator.removeNodeFromProcessing(node.Name)
			//nodeutil.RecordNodeStatusChange(allocator.recorder, node, "CIDRNotAvailable")
			return fmt.Errorf("failed to allocate cidr from cluster cidr at idx:%v: %v", idx, err)
		}
		allocated.allocatedCIDRs[idx] = podCIDR
	}

	//queue the assignment
	klog.V(4).Infof("Putting node %s with CIDR %v into the work queue", node.Name, allocated.allocatedCIDRs)
	allocator.nodeCIDRUpdateChannel <- allocated
	return nil
}

func (allocator *rangeAllocator) insertNodeToProcessing(nodeName string) bool {
	allocator.lock.Lock()
	defer allocator.lock.Unlock()
	if allocator.nodesInProcessing.Has(nodeName) {
		return false
	}

	allocator.nodesInProcessing.Insert(nodeName)
	return true
}

func (allocator *rangeAllocator) removeNodeFromProcessing(nodeName string) {
	allocator.lock.Lock()
	defer allocator.lock.Unlock()
	allocator.nodesInProcessing.Delete(nodeName)
}

func (allocator *rangeAllocator) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()

	klog.Infof("Starting range CIDR allocator")
	defer klog.Infof("Shutting down range CIDR allocator")

	if !cache.WaitForNamedCacheSync("cidrallocator", stopCh, allocator.nodesSynced) {
		return
	}

	for i := 0; i < 1; i++ {
		go allocator.worker(stopCh)
	}

	<-stopCh
}

func (allocator *rangeAllocator) worker(stopChan <-chan struct{}) {
	for {
		select {
		case workItem, ok := <-allocator.nodeCIDRUpdateChannel:
			if !ok {
				klog.Warning("Channel nodeCIDRUpdateChannel was unexpectedly closed")
				return
			}
			if err := allocator.updateCIDRsAllocation(workItem); err != nil {
				// Requeue the failed node for update again.
				allocator.nodeCIDRUpdateChannel <- workItem
			}
		case <-stopChan:
			return
		}
	}
}

// INFO: 更新 node spec.PodCIDRs
func (allocator *rangeAllocator) updateCIDRsAllocation(data nodeReservedCIDRs) error {
	var err error
	var node *corev1.Node
	defer allocator.removeNodeFromProcessing(data.nodeName)
	cidrsString := cidrsAsString(data.allocatedCIDRs)
	node, err = allocator.nodeLister.Get(data.nodeName)
	if err != nil {
		klog.Errorf("Failed while getting node %v for updating Node.Spec.PodCIDRs: %v", data.nodeName, err)
		return err
	}

	// skip if CIDR match
	if len(node.Spec.PodCIDRs) == len(data.allocatedCIDRs) {
		match := true
		for idx, cidr := range cidrsString {
			if node.Spec.PodCIDRs[idx] != cidr {
				match = false
				break
			}
		}
		if match {
			klog.V(4).Infof("Node %v already has allocated CIDR %v. It matches the proposed one.", node.Name, data.allocatedCIDRs)
			return nil
		}
	}

	// node has cidrs, release the reserved
	if len(node.Spec.PodCIDRs) != 0 {
		klog.Errorf("Node %v already has a CIDR allocated %v. Releasing the new one.", node.Name, node.Spec.PodCIDRs)
		for idx, cidr := range data.allocatedCIDRs {
			if releaseErr := allocator.cidrSets[idx].Release(cidr); releaseErr != nil {
				klog.Errorf("Error when releasing CIDR idx:%v value: %v err:%v", idx, cidr, releaseErr)
			}
		}
		return nil
	}

	// If we reached here, it means that the node has no CIDR currently assigned. So we set it.
	for i := 0; i < cidrUpdateRetries; i++ {
		if err = utilnode.PatchNodeCIDRs(allocator.client, types.NodeName(node.Name), cidrsString); err == nil {
			klog.Infof("Set node %v PodCIDR to %v", node.Name, cidrsString)
			return nil
		}
	}

	// if failed, release cidr
	klog.Errorf("Failed to update node %v PodCIDR to %v after multiple attempts: %v", node.Name, cidrsString, err)
	for idx, cidr := range data.allocatedCIDRs {
		if releaseErr := allocator.cidrSets[idx].Release(cidr); releaseErr != nil {
			klog.Errorf("Error releasing allocated CIDR for node %v: %v", node.Name, releaseErr)
		}
	}

	return err
}

func cidrsAsString(inCIDRs []*net.IPNet) []string {
	outCIDRs := make([]string, len(inCIDRs))
	for idx, inCIDR := range inCIDRs {
		outCIDRs[idx] = inCIDR.String()
	}
	return outCIDRs
}
