package routing

import (
	"context"
	"fmt"
	"github.com/vishvananda/netlink"
	"k8s-lx1036/k8s/network/kube-router/pkg/metrics"

	gobgpapi "github.com/osrg/gobgp/api"
	gobgp "github.com/osrg/gobgp/pkg/server"
	"k8s.io/klog/v2"
)

const (
	// Taken from: https://github.com/torvalds/linux/blob/master/include/uapi/linux/rtnetlink.h#L284
	zebraRouteOriginator = 0x11
)

func (controller *NetworkRoutingController) startBgpServer() error {
	addr := fmt.Sprintf("%s:50051,127.0.0.1:50051", controller.nodeIP.String())
	controller.bgpServer = gobgp.NewBgpServer(gobgp.GrpcListenAddress(addr))
	go controller.bgpServer.Serve()

	global := &gobgpapi.Global{
		As:              nodeAsnNumber,
		RouterId:        controller.routerID,
		ListenAddresses: controller.localAddressList,
		ListenPort:      int32(controller.bgpPort),
	}
	if err := controller.bgpServer.StartBgp(context.Background(), &gobgpapi.StartBgpRequest{Global: global}); err != nil {
		return fmt.Errorf(fmt.Sprintf("failed to start BGP server due to:%v", err))
	}

	go controller.watchBgpUpdates()

	if len(controller.globalPeerRouters) != 0 {
		err := connectToExternalBGPPeers(controller.bgpServer, controller.globalPeerRouters, controller.bgpGracefulRestart,
			controller.bgpGracefulRestartDeferralTime, controller.bgpGracefulRestartTime, controller.peerMultihopTTL)
		if err != nil {
			err2 := controller.bgpServer.StopBgp(context.Background(), &gobgpapi.StopBgpRequest{})
			if err2 != nil {
				klog.Errorf("Failed to stop bgpServer: %s", err2)
			}

			return fmt.Errorf("failed to peer with Global Peer Router(s): %s", err)
		}
	} else {
		klog.Infof("No Global Peer Routers configured. Peering skipped.")
	}

	return nil
}

func (controller *NetworkRoutingController) watchBgpUpdates() {
	pathWatch := func(path *gobgpapi.Path) {
		metrics.ControllerBGPAdvertisementsReceived.Inc()

		if path.NeighborIp == "<nil>" {
			return
		}
		klog.Infof("Processing bgp route advertisement from peer: %s", path.NeighborIp)
		if err := controller.injectRoute(path); err != nil {
			klog.Errorf(fmt.Sprintf("Failed to inject routes due to: %v", err))
		}
	}

	err := controller.bgpServer.MonitorTable(context.Background(), &gobgpapi.MonitorTableRequest{
		TableType: gobgpapi.TableType_GLOBAL,
		Family: &gobgpapi.Family{
			Afi:  gobgpapi.Family_AFI_IP,
			Safi: gobgpapi.Family_SAFI_UNICAST,
		},
	}, pathWatch)
	if err != nil {
		klog.Errorf("failed to register monitor global routing table callback due to : " + err.Error())
	}
}

func (controller *NetworkRoutingController) injectRoute(path *gobgpapi.Path) error {
	klog.Infof("injectRoute Path Looks Like: %s", path.String())
	var route *netlink.Route
	var link netlink.Link

	dst, nextHop, err := parseBGPPath(path)
	if err != nil {
		return err
	}

	tunnelName := generateTunnelName(nextHop.String())
	sameSubnet := controller.nodeSubnet.Contains(nextHop)

	// INFO: 如果是删除路由请求
	if path.IsWithdraw {
		klog.Infof("Removing route: '%s via %s' from peer in the routing table", dst, nextHop)

		// The path might be withdrawn because the peer became unestablished or it may be withdrawn because just the
		// path was withdrawn. Check to see if the peer is still established before deciding whether to clean the
		// tunnel and tunnel routes or whether to just delete the destination route.
		peerEstablished, err := controller.isPeerEstablished(nextHop.String())
		if err != nil {
			klog.Errorf("encountered error while checking peer status: %v", err)
		}
		if err == nil && !peerEstablished {
			klog.Infof("Peer '%s' was not found any longer, removing tunnel and routes", nextHop.String())
			controller.cleanupTunnel(dst, tunnelName)
			return nil
		}

		return deleteRoutesByDestination(dst)
	}

	shouldCreateTunnel := func() bool {
		if !controller.enableOverlays {
			return false
		}
		if controller.overlayType == "full" {
			return true
		}
		if controller.overlayType == "subnet" && !sameSubnet {
			return true
		}
		return false
	}

	// create IPIP tunnels only when node is not in same subnet or overlay-type is set to 'full'
	// if the user has disabled overlays, don't create tunnels. If we're not creating a tunnel, check to see if there is
	// any cleanup that needs to happen.
	if shouldCreateTunnel() {
		link, err = controller.setupOverlayTunnel(tunnelName, nextHop)
		if err != nil {
			return err
		}
	} else {
		// knowing that a tunnel shouldn't exist for this route, check to see if there are any lingering tunnels /
		// routes that need to be cleaned up.
		controller.cleanupTunnel(dst, tunnelName)
	}

	switch {
	case link != nil:
		// if we setup an overlay tunnel link, then use it for destination routing
		route = &netlink.Route{
			LinkIndex: link.Attrs().Index,
			Src:       controller.nodeIP,
			Dst:       dst,
			Protocol:  zebraRouteOriginator,
		}
	case sameSubnet:
		// if the nextHop is within the same subnet, add a route for the destination so that traffic can bet routed
		// at layer 2 and minimize the need to traverse a router
		route = &netlink.Route{
			Dst:      dst,
			Gw:       nextHop,
			Protocol: zebraRouteOriginator,
		}
	default:
		// otherwise, let BGP do its thing, nothing to do here
		return nil
	}

	// Alright, everything is in place, and we have our route configured, let's add it to the host's routing table
	klog.Infof("Inject route: '%s via %s' from peer to routing table", dst, nextHop)
	return netlink.RouteReplace(route)
	//return nil // for debug in local
}
