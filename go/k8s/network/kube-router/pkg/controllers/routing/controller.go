package routing

import (
	"context"
	gobgp "github.com/osrg/gobgp/pkg/server"
	"net"
	"time"

	gobgpapi "github.com/osrg/gobgp/api"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

type NetworkRoutingController struct {
	bgpServerStarted  bool
	bgpServer         *gobgp.BgpServer
	globalPeerRouters []*gobgpapi.Peer

	nodeSubnet net.IPNet // pod subnet for node

}

func NewNetworkRoutingController(
	clientSet kubernetes.Interface,
	nodeInformer cache.SharedIndexInformer,
	serviceInformer cache.SharedIndexInformer,
	endpointInformer cache.SharedIndexInformer,
) (*NetworkRoutingController, error) {

	factory := informers.NewSharedInformerFactory(clientSet, 0)
	nodeInformer := factory.Core().V1().Nodes().Informer()
	serviceInformer := factory.Core().V1().Services().Informer()
	endpointInformer := factory.Core().V1().Endpoints().Informer()

	controller := &NetworkRoutingController{}

	nodeInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.onNodeAdd,
		UpdateFunc: controller.onNodeUpdate,
		DeleteFunc: controller.onNodeDelete,
	})
	serviceInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    nil,
		UpdateFunc: nil,
		DeleteFunc: nil,
	})
	endpointInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    nil,
		UpdateFunc: nil,
		DeleteFunc: nil,
	})
}

func (controller *NetworkRoutingController) Start(stopCh chan struct{}) {

	klog.Infof("Starting BGP Server")
	t := time.NewTicker(time.Second * 5)
	defer t.Stop()
	// Wait till we are ready to launch BGP server
	for {
		err := controller.startBgpServer()
		if err != nil {
			klog.Errorf("Failed to start node BGP server: %s", err)
			select {
			case <-stopCh:
				klog.Infof("Shutting down network routes controller")
				return
			case <-t.C:
				klog.Infof("Retrying start of node BGP server")
				continue
			}
		} else {
			break
		}
	}

	controller.bgpServerStarted = true
	if !controller.bgpGracefulRestart {
		defer func() {
			err := controller.bgpServer.StopBgp(context.Background(), &gobgpapi.StopBgpRequest{})
			if err != nil {
				klog.Errorf("error shutting down BGP server: %s", err)
			}
		}()
	}

	// loop forever till notified to stop on stopCh
	for {
		var err error
		select {
		case <-stopCh:
			klog.Infof("Shutting down network routes controller")
			return
		default:
		}

		// Update ipset entries
		if controller.enablePodEgress || controller.enableOverlays {
			klog.V(1).Info("Syncing ipsets")
			err = controller.syncNodeIPSets()
			if err != nil {
				klog.Errorf("Error synchronizing ipsets: %s", err.Error())
			}
		}

		// enable IP forwarding for the packets coming in/out from the pods
		err = controller.enableForwarding()
		if err != nil {
			klog.Errorf("Failed to enable IP forwarding of traffic from pods: %s", err.Error())
		}

		// advertise or withdraw IPs for the services to be reachable via host
		toAdvertise, toWithdraw, err := controller.getActiveVIPs()
		if err != nil {
			klog.Errorf("failed to get routes to advertise/withdraw %s", err)
		}

		klog.Infof("Performing periodic sync of service VIP routes")
		controller.advertiseVIPs(toAdvertise)
		controller.withdrawVIPs(toWithdraw)

		klog.Info("Performing periodic sync of pod CIDR routes")
		err = controller.advertisePodRoute()
		if err != nil {
			klog.Errorf("Error advertising route: %s", err.Error())
		}

		err = controller.AddPolicies()
		if err != nil {
			klog.Errorf("Error adding BGP policies: %s", err.Error())
		}

		if controller.bgpEnableInternal {
			controller.syncInternalPeers()
		}

		if err == nil {
			//healthcheck.SendHeartBeat(healthChan, "NRC")
		} else {
			klog.Errorf("Error during periodic sync in network routing controller. Error: " + err.Error())
			klog.Errorf("Skipping sending heartbeat from network routing controller as periodic sync failed.")
		}

		select {
		case <-stopCh:
			klog.Infof("Shutting down network routes controller")
			return
		case <-t.C:
		}
	}
}
