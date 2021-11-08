package routing

import (
	"context"
	"fmt"
	gobgpapi "github.com/osrg/gobgp/api"
	"k8s-lx1036/k8s/network/kube-router/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

func (controller *NetworkRoutingController) onNodeAdd(obj interface{}) {
	node := obj.(*corev1.Node)
	nodeIP, err := utils.GetNodeIP(node)
	if err != nil {
		klog.Errorf("New node received, but we were unable to add it as we were couldn't find its node IP: %v", err)
		return
	}

	klog.Infof("Received node %s added update from watch API so peer with new node", nodeIP)
	controller.handleNodeUpdate(obj)
}

func (controller *NetworkRoutingController) onNodeUpdate(oldObj, newObj interface{}) {
	// we are only interested in node add/delete, so skip update
}

func (controller *NetworkRoutingController) onNodeDelete(obj interface{}) {
	node, ok := obj.(*corev1.Node)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			klog.Errorf("unexpected object type: %v", obj)
			return
		}
		if node, ok = tombstone.Obj.(*corev1.Node); !ok {
			klog.Errorf("unexpected object type: %v", obj)
			return
		}
	}
	nodeIP, err := utils.GetNodeIP(node)
	// INFO: 如果 node 被删除则 node ip 获取可能获取不了，这样在 nodeLister.List() 时也没有这个 node
	if err == nil {
		klog.Infof("Received node %s removed update from watch API, so remove node from peer", nodeIP)
	} else {
		klog.Infof("Received node (IP unavailable) removed update from watch API, so remove node from peer")
	}

	controller.handleNodeUpdate(obj)
}

func (controller *NetworkRoutingController) handleNodeUpdate(obj interface{}) {
	if !controller.bgpServerStarted {
		return
	}

	// update export policies so that NeighborSet gets updated with new set of nodes
	err := controller.AddPolicies()
	if err != nil {
		klog.Errorf("Error adding BGP policies: %s", err.Error())
	}

	if controller.bgpEnableInternal {
		controller.syncInternalPeers()
	}
}

func (controller *NetworkRoutingController) isPeerEstablished(peerIP string) (bool, error) {
	var peerConnected bool
	err := controller.bgpServer.ListPeer(context.TODO(), &gobgpapi.ListPeerRequest{
		Address: peerIP,
	}, func(peer *gobgpapi.Peer) {
		if peer.Conf.NeighborAddress == peerIP && peer.State.SessionState == gobgpapi.PeerState_ESTABLISHED {
			peerConnected = true
		}
	})
	if err != nil {
		return false, fmt.Errorf("unable to list peers to see if tunnel & routes need to be removed: %v", err)
	}

	return peerConnected, nil
}
