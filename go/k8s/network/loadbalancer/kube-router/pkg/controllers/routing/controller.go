package routing

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"sync"
	"time"

	"k8s-lx1036/k8s/network/loadbalancer/kube-router/cmd/app/options"
	"k8s-lx1036/k8s/network/loadbalancer/kube-router/pkg/utils"

	gobgpapi "github.com/osrg/gobgp/api"
	gobgp "github.com/osrg/gobgp/pkg/server"
	"github.com/vishvananda/netlink"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	nodeutil "k8s.io/kubernetes/pkg/controller/util/node"
)

const (
	// @see https://github.com/vishvananda/netlink/blob/master/link_linux.go#L1672-L1693
	IfaceNotFound = "Link not found"

	BridgeName = "kube-bridge"
)

type NetworkRoutingController struct {
	CondMutex *sync.Cond
	mu        sync.Mutex

	nodeInformer     cache.SharedIndexInformer
	serviceInformer  cache.SharedIndexInformer
	endpointInformer cache.SharedIndexInformer
	nodeLister       corev1.NodeLister
	serviceLister    corev1.ServiceLister
	endpointLister   corev1.EndpointsLister

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

	nodeSubnet net.IPNet // pod subnet for node
	podCidr    string    // INFO: 使用 kube-controller-manager IPAM 分配的 pod cidr
	nodeIP     net.IP
	nodeName   string
}

func NewNetworkRoutingController(
	option *options.Options,
	clientSet kubernetes.Interface,
) (*NetworkRoutingController, error) {

	factory := informers.NewSharedInformerFactory(clientSet, 0)
	nodeInformer := factory.Core().V1().Nodes().Informer()
	nodeLister := factory.Core().V1().Nodes().Lister()
	serviceInformer := factory.Core().V1().Services().Informer()
	serviceLister := factory.Core().V1().Services().Lister()
	endpointInformer := factory.Core().V1().Endpoints().Informer()
	endpointLister := factory.Core().V1().Endpoints().Lister()

	controller := &NetworkRoutingController{
		CondMutex: sync.NewCond(&sync.Mutex{}),

		nodeInformer:     nodeInformer,
		serviceInformer:  serviceInformer,
		endpointInformer: endpointInformer,
		nodeLister:       nodeLister,
		serviceLister:    serviceLister,
		endpointLister:   endpointLister,

		enableOverlays:          option.EnableOverlays,
		enablePodEgress:         option.EnablePodEgress,
		enableIBGP:              option.EnableIBGP,
		autoMTU:                 option.AutoMTU,
		advertiseClusterIP:      option.AdvertiseClusterIP,
		advertiseExternalIP:     option.AdvertiseExternalIP,
		advertiseLoadBalancerIP: option.AdvertiseLoadBalancerIP,
		advertisePodCidr:        option.AdvertisePodCidr,

		defaultNodeAsnNumber: 64512, // this magic number is first of the private ASN range, use it as default
		bgpPort:              option.BGPPort,
	}

	nodeInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    nodeutil.CreateAddNodeHandler(controller.onNodeAdd),
		UpdateFunc: nodeutil.CreateUpdateNodeHandler(controller.onNodeUpdate),
		DeleteFunc: nodeutil.CreateDeleteNodeHandler(controller.onNodeDelete),
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

	cidr, err := utils.GetPodCidrFromNodeSpec(clientSet)
	if err != nil {
		klog.Fatalf("Failed to get pod CIDR from node spec. kube-router relies on kube-controller-manager to "+
			"allocate pod CIDR for the node or an annotation `kube-router.io/pod-cidr`. Error: %v", err)
		return nil, fmt.Errorf("failed to get pod CIDR details from Node.spec: %s", err.Error())
	}
	controller.podCidr = cidr

	return controller, nil
}

func (controller *NetworkRoutingController) Run(stopCh <-chan struct{}) {
	go controller.nodeInformer.Run(stopCh)
	go controller.serviceInformer.Run(stopCh)
	go controller.endpointInformer.Run(stopCh)

	ok := cache.WaitForCacheSync(stopCh, controller.nodeInformer.HasSynced, controller.serviceInformer.HasSynced, controller.endpointInformer.HasSynced)
	if !ok {
		klog.Fatal(fmt.Sprintf("failed to sync cache for node/service/endpoint"))
	}

	// INFO: (1) initial
	var err error
	if controller.enableCNI {
		controller.updateCNIConfig()
	}

	klog.V(1).Info("Populating ipsets.")
	err = controller.syncNodeIPSets()
	if err != nil {
		klog.Errorf("Failed initial ipset setup: %s", err)
	}

	// In case of cluster provisioned on AWS disable source-destination check
	if controller.disableSrcDstCheck {
		controller.disableSourceDestinationCheck()
		controller.initSrcDstCheckDone = true
	}
	// enable IP forwarding for the packets coming in/out from the pods
	err = controller.enableForwarding()
	if err != nil {
		klog.Errorf("Failed to enable IP forwarding of traffic from pods: %s", err.Error())
	}

	controller.CondMutex.Broadcast()

	// INFO: route based policy, and custom route table kube-router
	// Handle ipip tunnel overlay
	if controller.enableOverlays {
		klog.Info("IPIP Tunnel Overlay enabled in configuration.")
		klog.Info("Setting up overlay networking.")
		err = controller.enablePolicyBasedRouting()
		if err != nil {
			klog.Errorf("Failed to enable required policy based routing: %s", err.Error())
		}
	} else {
		klog.Info("IPIP Tunnel Overlay disabled in configuration.")
		klog.Info("Cleaning up old overlay networking if needed.")
		err = controller.disablePolicyBasedRouting()
		if err != nil {
			klog.Errorf("Failed to disable policy based routing: %s", err.Error())
		}
	}

	// INFO: ensure pod egress iptable rule
	klog.V(1).Info("Performing cleanup of depreciated rules/ipsets (if needed).")
	err = controller.deleteBadPodEgressRules()
	if err != nil {
		klog.Errorf("Error cleaning up old/bad Pod egress rules: %s", err.Error())
	}
	// Handle Pod egress masquerading configuration
	if controller.enablePodEgress {
		klog.V(1).Infoln("Enabling Pod egress.")
		err = controller.createPodEgressRule()
		if err != nil {
			klog.Errorf("Error enabling Pod egress: %s", err.Error())
		}
	} else {
		klog.V(1).Infoln("Disabling Pod egress.")
		err = controller.deletePodEgressRule()
		if err != nil {
			klog.Warningf("Error cleaning up Pod Egress related networking: %s", err)
		}
	}

	// INFO: create 'kube-bridge' interface to which pods will be connected
	kubeBridgeIf, err := netlink.LinkByName(BridgeName)
	if err != nil && err.Error() == IfaceNotFound {
		linkAttrs := netlink.NewLinkAttrs()
		linkAttrs.Name = BridgeName
		bridge := &netlink.Bridge{LinkAttrs: linkAttrs}
		if err = netlink.LinkAdd(bridge); err != nil {
			klog.Errorf("Failed to create `kube-router` bridge due to %s. Will be created by CNI bridge "+
				"plugin when pod is launched.", err.Error())
		}
		kubeBridgeIf, err = netlink.LinkByName(BridgeName)
		if err != nil {
			klog.Errorf("Failed to find created `kube-router` bridge due to %s. Will be created by CNI "+
				"bridge plugin when pod is launched.", err.Error())
		}
		err = netlink.LinkSetUp(kubeBridgeIf)
		if err != nil {
			klog.Errorf("Failed to bring `kube-router` bridge up due to %s. Will be created by CNI bridge "+
				"plugin at later point when pod is launched.", err.Error())
		}
	}
	if controller.autoMTU {
		mtu, err := utils.GetMTUFromNodeIP(controller.nodeIP, controller.enableOverlays)
		if err != nil {
			klog.Errorf("Failed to find MTU for node IP: %s for intelligently setting the kube-bridge MTU "+
				"due to %s.", controller.nodeIP, err.Error())
		}
		if mtu > 0 {
			klog.Infof("Setting MTU of kube-bridge interface to: %d", mtu)
			err = netlink.LinkSetMTU(kubeBridgeIf, mtu)
			if err != nil {
				klog.Errorf("Failed to set MTU for kube-bridge interface due to: %s", err.Error())
			}
		} else {
			klog.Infof("Not setting MTU of kube-bridge interface")
		}
	}
	// enable netfilter for the bridge
	if _, err := exec.Command("modprobe", "br_netfilter").CombinedOutput(); err != nil {
		klog.Errorf("Failed to enable netfilter for bridge. Network policies and service proxy may "+
			"not work: %s", err.Error())
	}
	sysctlErr := utils.SetSysctl(utils.BridgeNFCallIPTables, 1)
	if sysctlErr != nil {
		klog.Errorf("Failed to enable iptables for bridge. Network policies and service proxy may "+
			"not work: %s", sysctlErr.Error())
	}
	if controller.isIpv6 {
		sysctlErr = utils.SetSysctl(utils.BridgeNFCallIP6Tables, 1)
		if sysctlErr != nil {
			klog.Errorf("Failed to enable ip6tables for bridge. Network policies and service proxy may "+
				"not work: %s", sysctlErr.Error())
		}
	}

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

		if controller.enableIBGP {
			controller.syncFullMeshIBGPPeers()
		}
	}, time.Second*5, stopCh)
}
