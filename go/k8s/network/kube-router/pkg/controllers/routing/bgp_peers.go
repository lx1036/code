package routing

import (
	"context"
	"fmt"
	gobgpapi "github.com/osrg/gobgp/api"
	"k8s-lx1036/k8s/network/kube-router/pkg/metrics"
	"k8s-lx1036/k8s/network/kube-router/pkg/utils"
	"k8s.io/apimachinery/pkg/labels"
	"strings"
	"time"

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

// Refresh the peer relationship with rest of the nodes in the cluster (iBGP peers). Node add/remove
// events should ensure peer relationship with only currently active nodes. In case
// we miss any events from API server this method which is called periodically
// ensures peer relationship with removed nodes is deleted.
func (controller *NetworkRoutingController) syncFullMeshIBGPPeers() {
	controller.mu.Lock()
	defer controller.mu.Unlock()

	start := time.Now()
	defer func() {
		endTime := time.Since(start)
		metrics.ControllerBGPInternalPeersSyncTime.Observe(endTime.Seconds())
		klog.Infof("Syncing BGP peers for the node took %v", endTime)
	}()

	nodes, err := controller.nodeLister.List(labels.Everything())
	if err != nil {
		klog.Infof(fmt.Sprintf("[syncIBGPPeers]list nodes err:%v", err))
		return
	}

	metrics.ControllerBPGPeers.Set(float64(len(nodes)))

	// establish peer and add Pod CIDRs with current set of nodes
	for _, node := range nodes {
		nodeIP, err := utils.GetNodeIP(node)
		if err != nil {
			klog.Errorf("Failed to find a node IP and therefore cannot sync internal BGP Peer: %v", err)
			continue
		}

		// skip self
		if nodeIP.String() == controller.nodeIP.String() {
			continue
		}

		controller.activeNodes[nodeIP.String()] = true
		peer := &gobgpapi.Peer{
			Conf: &gobgpapi.PeerConf{
				NeighborAddress: nodeIP.String(),
				PeerAs:          controller.nodeAsnNumber,
			},
			Transport: &gobgpapi.Transport{
				RemotePort: controller.bgpPort,
			},
		}
		if controller.bgpGracefulRestart {
			peer.GracefulRestart = &gobgpapi.GracefulRestart{
				Enabled:         true,
				RestartTime:     uint32(controller.bgpGracefulRestartTime.Seconds()),
				DeferralTime:    uint32(controller.bgpGracefulRestartDeferralTime.Seconds()),
				LocalRestarting: true,
			}

			peer.AfiSafis = []*gobgpapi.AfiSafi{
				{
					Config: &gobgpapi.AfiSafiConfig{
						Family:  &gobgpapi.Family{Afi: gobgpapi.Family_AFI_IP, Safi: gobgpapi.Family_SAFI_UNICAST},
						Enabled: true,
					},
					MpGracefulRestart: &gobgpapi.MpGracefulRestart{
						Config: &gobgpapi.MpGracefulRestartConfig{
							Enabled: true,
						},
						State: &gobgpapi.MpGracefulRestartState{},
					},
				},
				{
					Config: &gobgpapi.AfiSafiConfig{
						Family:  &gobgpapi.Family{Afi: gobgpapi.Family_AFI_IP6, Safi: gobgpapi.Family_SAFI_UNICAST},
						Enabled: true,
					},
					MpGracefulRestart: &gobgpapi.MpGracefulRestart{
						Config: &gobgpapi.MpGracefulRestartConfig{
							Enabled: true,
						},
						State: &gobgpapi.MpGracefulRestartState{},
					},
				},
			}
		}

		// we are rr-server peer with other rr-client with reflection enabled
		if controller.bgpRRServer {
			if _, ok := node.ObjectMeta.Annotations[rrClientAnnotation]; ok {
				// add rr options with clusterId
				peer.RouteReflector = &gobgpapi.RouteReflector{
					RouteReflectorClient:    true,
					RouteReflectorClusterId: fmt.Sprint(controller.bgpClusterID),
				}
			}
		}

		// TODO: check if a node is already added as neighbor in a better way than add and catch error
		if err := controller.bgpServer.AddPeer(context.Background(), &gobgpapi.AddPeerRequest{
			Peer: peer,
		}); err != nil {
			// https://github.com/osrg/gobgp/blob/master/pkg/server/server.go#L2891-L2899
			if !strings.Contains(err.Error(), "can't overwrite the existing peer") {
				klog.Errorf("Failed to add node %s as peer due to %s", nodeIP.String(), err)
			}
		}

	}

}
