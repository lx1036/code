package controller

import (
	"context"
	"fmt"
	"net"

	v1 "k8s-lx1036/k8s/network/bgplb/pkg/apis/bgplb.k9s.io/v1"

	gobgpapi "github.com/osrg/gobgp/api"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

func (controller *BgpLBController) onBGPPeerAdd(obj interface{}) {
	peer := obj.(*v1.BgpPeer)
	klog.Infof("bgpPeer %s/%s was added, enqueuing it for submission", peer.Namespace, peer.Name)

	//filter peer with nodeSelector
	if peer.Spec.NodeSelector != nil {

	}

	clone := peer.DeepCopy()
	bgpPeer, err := convertToBgpPeer(clone)
	if err != nil {
		klog.Error(err)
		return
	}

	err = controller.bgpServer.AddPeer(context.Background(), &gobgpapi.AddPeerRequest{
		Peer: bgpPeer,
	})
	if err != nil {
		klog.Error(err)
		return
	}
}

func (controller *BgpLBController) onBGPPeerUpdate(oldObj, newObj interface{}) {

}

func (controller *BgpLBController) onBGPPeerDelete(obj interface{}) {
	peer, ok := obj.(*v1.BgpPeer)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			klog.Errorf("Couldn't get object from tombstone %#v", obj)
			return
		}
		peer, ok = tombstone.Obj.(*v1.BgpPeer)
		if !ok {
			klog.Errorf("Tombstone contained object that is not expected %#v", obj)
			return
		}
	}

	if peer != nil {
		clone := peer.DeepCopy()
		bgpPeer, err := convertToBgpPeer(clone)
		if err != nil {
			klog.Error(err)
			return
		}

		err = controller.bgpServer.DeletePeer(context.TODO(), &gobgpapi.DeletePeerRequest{
			Address:   bgpPeer.Conf.NeighborAddress,
			Interface: bgpPeer.Conf.NeighborInterface,
		})
		if err != nil {
			klog.Error(err)
			return
		}
	}
}

func defaultFamily(ip net.IP) *v1.Family {
	family := &v1.Family{
		Afi:  "AFI_IP",
		Safi: "SAFI_UNICAST",
	}
	if ip.To4() == nil {
		family = &v1.Family{
			Afi:  "AFI_IP6",
			Safi: "SAFI_UNICAST",
		}
	}

	return family
}

func convertToBgpPeer(peer *v1.BgpPeer) (*gobgpapi.Peer, error) {
	// set default afisafi
	if len(peer.Spec.AfiSafis) == 0 {
		ip := net.ParseIP(peer.Spec.Conf.NeighborAddress)
		if ip == nil {
			return nil, fmt.Errorf("field Spec.Conf.NeighborAddress invalid")
		}
		peer.Spec.AfiSafis = append(peer.Spec.AfiSafis, &v1.AfiSafi{
			Config: &v1.AfiSafiConfig{
				Family:  defaultFamily(ip),
				Enabled: true,
			},
			AddPaths: &v1.AddPaths{
				Config: &v1.AddPathsConfig{
					SendMax: 10,
				},
			},
		})
	}

	bgpPeer, err := peer.Spec.ConvertToGoBgpPeer()
	if err != nil {
		return nil, err
	}

	return bgpPeer, nil
}
