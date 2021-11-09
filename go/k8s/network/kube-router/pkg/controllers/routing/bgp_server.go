package routing

import (
	"context"
	"fmt"
	"time"

	"k8s-lx1036/k8s/network/kube-router/pkg/metrics"

	gobgpapi "github.com/osrg/gobgp/api"
	gobgp "github.com/osrg/gobgp/pkg/server"
	"k8s.io/klog/v2"
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

// connectToExternalBGPPeers adds all the configured eBGP peers (global or node specific) as neighbours
func connectToExternalBGPPeers(server *gobgp.BgpServer, peerNeighbors []*gobgpapi.Peer, bgpGracefulRestart bool,
	bgpGracefulRestartDeferralTime time.Duration, bgpGracefulRestartTime time.Duration, peerMultihopTTL uint8) error {
	for _, neighbor := range peerNeighbors {
		if bgpGracefulRestart {
			neighbor.GracefulRestart = &gobgpapi.GracefulRestart{
				Enabled:         true,
				RestartTime:     uint32(bgpGracefulRestartTime.Seconds()),
				DeferralTime:    uint32(bgpGracefulRestartDeferralTime.Seconds()),
				LocalRestarting: true,
			}

			neighbor.AfiSafis = []*gobgpapi.AfiSafi{
				{
					Config: &gobgpapi.AfiSafiConfig{
						Family:  &gobgpapi.Family{Afi: gobgpapi.Family_AFI_IP, Safi: gobgpapi.Family_SAFI_UNICAST},
						Enabled: true,
					},
					MpGracefulRestart: &gobgpapi.MpGracefulRestart{
						Config: &gobgpapi.MpGracefulRestartConfig{
							Enabled: true,
						},
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
					},
				},
			}
		}
		if peerMultihopTTL > 1 {
			neighbor.EbgpMultihop = &gobgpapi.EbgpMultihop{
				Enabled:     true,
				MultihopTtl: uint32(peerMultihopTTL),
			}
		}
		err := server.AddPeer(context.Background(), &gobgpapi.AddPeerRequest{Peer: neighbor})
		if err != nil {
			return fmt.Errorf("error peering with peer router %q due to: %v", neighbor.Conf.NeighborAddress, err)
		}

		klog.Infof("Successfully configured %s in ASN %v as BGP peer to the node",
			neighbor.Conf.NeighborAddress, neighbor.Conf.PeerAs)
	}

	return nil
}
