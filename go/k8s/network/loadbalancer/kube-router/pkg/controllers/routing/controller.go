package routing

import (
	"context"
	"errors"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"net"
	"sync"
	"time"

	"k8s-lx1036/k8s/network/loadbalancer/kube-router/cmd/app/options"
	"k8s-lx1036/k8s/network/loadbalancer/kube-router/pkg/utils"

	gobgpapi "github.com/osrg/gobgp/api"
	gobgp "github.com/osrg/gobgp/pkg/server"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

const (
	// @see https://github.com/vishvananda/netlink/blob/master/link_linux.go#L1672-L1693
	IfaceNotFound = "Link not found"

	BridgeName = "kube-bridge"
)

type NetworkRoutingController struct {
	CondMutex *sync.Cond
	mu        sync.Mutex

	serviceInformer  cache.SharedIndexInformer
	endpointInformer cache.SharedIndexInformer
	serviceLister    cache.Indexer
	endpointLister   cache.Indexer

	enableOverlays          bool
	enablePodEgress         bool
	enableIBGP              bool
	isIpv6                  bool
	autoMTU                 bool
	advertiseClusterIP      bool
	advertiseExternalIP     bool
	advertiseLoadBalancerIP bool
	advertisePodCidr        bool

	bgpServerStarted               bool
	bgpServer                      *gobgp.BgpServer
	globalPeerRouters              []*gobgpapi.Peer
	bgpGracefulRestart             bool
	bgpGracefulRestartTime         time.Duration
	bgpGracefulRestartDeferralTime time.Duration
	peerMultihopTTL                uint8
	defaultNodeAsnNumber           uint32
	bgpPort                        uint32
	routerID                       string
	localAddressList               []string
	overrideNextHop                bool
	nodeCommunities                []string

	nodeSubnet net.IPNet // pod subnet for node
	podCidr    string    // INFO: 使用 kube-controller-manager IPAM 分配的 pod cidr
	nodeIP     net.IP
	nodeName   string
}

func NewNetworkRoutingController(
	option *options.Options,
	clientSet kubernetes.Interface,
	svcInformer cache.SharedIndexInformer,
	epInformer cache.SharedIndexInformer,
) (*NetworkRoutingController, error) {
	controller := &NetworkRoutingController{
		CondMutex: sync.NewCond(&sync.Mutex{}),

		serviceInformer:  svcInformer,
		serviceLister:    svcInformer.GetIndexer(),
		endpointInformer: epInformer,
		endpointLister:   epInformer.GetIndexer(),

		enableOverlays:          option.EnableOverlays,
		enablePodEgress:         option.EnablePodEgress,
		enableIBGP:              option.EnableIBGP,
		autoMTU:                 option.AutoMTU,
		advertiseClusterIP:      option.AdvertiseClusterIP,
		advertiseExternalIP:     option.AdvertiseExternalIP,
		advertiseLoadBalancerIP: option.AdvertiseLoadBalancerIP,
		advertisePodCidr:        option.AdvertisePodCidr,

		defaultNodeAsnNumber:           64512, // this magic number is first of the private ASN range, use it as default
		bgpPort:                        option.BGPPort,
		bgpGracefulRestart:             option.BGPGracefulRestart,
		bgpGracefulRestartTime:         option.BGPGracefulRestartTime,
		bgpGracefulRestartDeferralTime: option.BGPGracefulRestartDeferralTime,
		overrideNextHop:                option.OverrideNextHop,
	}

	svcInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			controller.OnServiceUpdate(obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			controller.OnServiceUpdate(newObj)
		},
		DeleteFunc: func(obj interface{}) {
			controller.OnServiceUpdate(obj)
		},
	})
	epInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    nil,
		UpdateFunc: nil,
		DeleteFunc: nil,
	})

	node, err := utils.GetNodeObject(clientSet)
	if err != nil {
		return nil, errors.New("failed getting node object from API server: " + err.Error())
	}
	controller.nodeName = node.Name
	nodeIP, err := utils.GetNodeIP(node)
	if err != nil {
		return nil, errors.New("failed getting IP address from node object: " + err.Error())
	}
	controller.nodeIP = nodeIP
	controller.isIpv6 = nodeIP.To4() == nil
	controller.routerID = nodeIP.String()
	controller.localAddressList = append(controller.localAddressList, controller.nodeIP.String())

	controller.globalPeerRouters = append(controller.globalPeerRouters, &gobgpapi.Peer{
		Conf: &gobgpapi.PeerConf{
			NeighborAddress: option.PeerRouterAddr,
			PeerAs:          65100,
		},
		Timers: &gobgpapi.Timers{Config: &gobgpapi.TimersConfig{HoldTime: uint64(option.BGPHoldTime)}},
		Transport: &gobgpapi.Transport{
			RemotePort: option.PeerRouterPort,
		},
	})

	cidr, err := utils.GetPodCidrFromNodeSpec(clientSet)
	if err != nil {
		klog.Fatalf("Failed to get pod CIDR from node spec. kube-router relies on kube-controller-manager to "+
			"allocate pod CIDR for the node or an annotation `kube-router.io/pod-cidr`. Error: %v", err)
		return nil, fmt.Errorf("failed to get pod CIDR details from Node.spec: %s", err.Error())
	}
	controller.podCidr = cidr

	return controller, nil
}

func (controller *NetworkRoutingController) OnServiceUpdate(obj interface{}) {
	svc := obj.(*corev1.Service)
	if utils.IsHeadlessService(svc) {
		return
	}

	toAdvertise, toWithdraw, err := controller.getActiveVIPs(svc)
	if err != nil {
		klog.Errorf("error getting routes for services: %s", err)
		return
	}

	// update export policies so that new VIP's gets added to clusterip prefixset and vip gets advertised to peers
	err = controller.AddPolicies()
	if err != nil {
		klog.Errorf("Error adding BGP policies: %s", err.Error())
	}

	controller.advertiseVIPs(toAdvertise)
	controller.withdrawVIPs(toWithdraw)
}

func (controller *NetworkRoutingController) Run(stopCh <-chan struct{}) {
	controller.CondMutex.Broadcast()

	// INFO: start bgp server
	klog.Infof("Starting BGP Server...")
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

	// INFO: loop forever and block
	wait.Until(func() {
		// advertise or withdraw IPs for the services to be reachable via host
		toAdvertise, toWithdraw, err := controller.getAllActiveVIPs()
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
	}, time.Second*5, stopCh)
}
