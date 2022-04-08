package speaker

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"k8s-lx1036/k8s/network/loadbalancer/bgplb/pkg/apis/bgplb.k9s.io/v1"
	"k8s-lx1036/k8s/network/loadbalancer/bgplb/pkg/client/clientset/versioned"
	"k8s-lx1036/k8s/network/loadbalancer/bgplb/pkg/client/informers/externalversions"
	listerv1 "k8s-lx1036/k8s/network/loadbalancer/bgplb/pkg/client/listers/bgplb.k9s.io/v1"
	"k8s-lx1036/k8s/network/loadbalancer/bgplb/pkg/utils"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	gobgpapi "github.com/osrg/gobgp/v3/api"
	bgppacket "github.com/osrg/gobgp/v3/pkg/packet/bgp"
	gobgp "github.com/osrg/gobgp/v3/pkg/server"
	apb "google.golang.org/protobuf/types/known/anypb"

	"google.golang.org/grpc"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

const (
	maxSize = 4 << 20 //4MB

	DefaultBGPPort = 179 // 本地测试使用 1790，不要用默认的 179
)

type SpeakerController struct {
	sync.RWMutex

	kubeClient kubernetes.Interface
	crdClient  *versioned.Clientset
	events     record.EventRecorder
	queue      workqueue.RateLimitingInterface

	crdFactory      externalversions.SharedInformerFactory
	bgpPeerInformer cache.SharedIndexInformer
	svcIndexer      cache.Indexer
	svcInformer     cache.Controller
	epIndexer       cache.Indexer
	epInformer      cache.Controller
	bgpPeerLister   listerv1.BGPPeerLister
	bgppeerInformer cache.SharedIndexInformer

	syncFuncs        []cache.InformerSynced
	bgpServer        *gobgp.BgpServer
	nodeName         string
	nodeIP           net.IP
	bgpServerStarted bool
	peerConnected    bool

	utils.Backoff
}

type svcKey string
type bgppeer string

func NewSpeakerController(restConfig *restclient.Config, grpcPort int, nodeName string, debug bool) *SpeakerController {
	kubeClient := kubernetes.NewForConfigOrDie(restConfig)
	crdClient := versioned.NewForConfigOrDie(restConfig)

	broadcaster := record.NewBroadcaster()
	//broadcaster.StartStructuredLogging(0)
	broadcaster.StartRecordingToSink(&typedv1.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	recorder := broadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "bgp-speaker-controller"})
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	c := &SpeakerController{
		kubeClient: kubeClient,
		crdClient:  crdClient,
		events:     recorder,
		queue:      queue,

		nodeName: nodeName,
	}

	node, err := kubeClient.CoreV1().Nodes().Get(context.TODO(), c.nodeName, metav1.GetOptions{})
	if err != nil {
		klog.Fatal(err)
	}
	c.nodeIP, err = GetNodeIP(node)
	if err != nil {
		klog.Fatal(err)
	}

	addr := "127.0.0.1:50053"
	if grpcPort != 0 {
		addr = fmt.Sprintf("%s,%s:%d", addr, c.nodeIP.String(), grpcPort)
	}
	bgpServer := gobgp.NewBgpServer(gobgp.GrpcListenAddress(addr), gobgp.GrpcOption([]grpc.ServerOption{
		grpc.MaxRecvMsgSize(maxSize),
		grpc.MaxSendMsgSize(maxSize),
	}))
	c.bgpServer = bgpServer
	if debug {
		c.bgpServer.SetLogLevel(context.TODO(), &gobgpapi.SetLogLevelRequest{
			Level: gobgpapi.SetLogLevelRequest_DEBUG,
		})
	}

	// only watch nodeName bgppeer @see https://github.com/kubernetes/kubernetes/blob/v1.23.5/pkg/kubelet/kubelet.go#L408-L416
	c.crdFactory = externalversions.NewSharedInformerFactoryWithOptions(crdClient, 0, externalversions.WithTweakListOptions(func(options *metav1.ListOptions) {
		options.FieldSelector = fields.Set{metav1.ObjectNameField: c.nodeName}.String()
	}))
	c.bgppeerInformer = c.crdFactory.Bgplb().V1().BGPPeers().Informer()
	c.bgpPeerLister = c.crdFactory.Bgplb().V1().BGPPeers().Lister()
	c.bgppeerInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				c.queue.Add(bgppeer(key))
			}
		},
		DeleteFunc: func(obj interface{}) {
			bgpp, ok := obj.(*v1.BGPPeer)
			if !ok {
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					klog.Errorf("unexpected object type: %v", obj)
					return
				}
				if bgpp, ok = tombstone.Obj.(*v1.BGPPeer); !ok {
					klog.Errorf("unexpected object type: %v", obj)
					return
				}
			}

			if err = c.bgpServer.DeletePeer(context.TODO(), &gobgpapi.DeletePeerRequest{Address: bgpp.Spec.PeerAddress}); err != nil {
				klog.Errorf(fmt.Sprintf("delete BGP peer %s err:%v", bgpp.Spec.PeerAddress, err))
			}
		},
	})

	svcWatcher := cache.NewListWatchFromClient(kubeClient.CoreV1().RESTClient(), "services", corev1.NamespaceAll, fields.Everything())
	c.svcIndexer, c.svcInformer = cache.NewIndexerInformer(svcWatcher, &corev1.Service{}, 0, cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			switch t := obj.(type) {
			case *corev1.Service:
				return t.Spec.Type == corev1.ServiceTypeLoadBalancer &&
					len(t.Status.LoadBalancer.Ingress) != 0 &&
					len(t.Status.LoadBalancer.Ingress[0].IP) != 0 // only watch LoadBalancer service
			default:
				runtime.HandleError(fmt.Errorf("object passed to %T that is not expected: %T", c, obj))
				return false
			}
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				key, err := cache.MetaNamespaceKeyFunc(obj)
				if err == nil {
					c.queue.Add(svcKey(key))
				}
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				key, err := cache.MetaNamespaceKeyFunc(newObj)
				if err == nil {
					c.queue.Add(svcKey(key))
				}
			},
			DeleteFunc: func(obj interface{}) {
				svc, ok := obj.(*corev1.Service)
				if !ok {
					tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
					if !ok {
						klog.Errorf("unexpected object type: %v", obj)
						return
					}
					if svc, ok = tombstone.Obj.(*corev1.Service); !ok {
						klog.Errorf("unexpected object type: %v", obj)
						return
					}
				}

				// withdraw route
				ip := svc.Status.LoadBalancer.Ingress[0].IP
				if c.isIPAdvertised(ip) {
					if err = c.withdrawIP(context.TODO(), ip); err != nil {
						klog.Errorf(fmt.Sprintf("%v", err))
					} else {
						klog.Infof(fmt.Sprintf("withdraw route %s via nextHop %s", ip, c.nodeIP.String()))
					}
				}
			},
		},
	}, cache.Indexers{})

	endpointWatcher := cache.NewListWatchFromClient(kubeClient.CoreV1().RESTClient(), "endpoints", corev1.NamespaceAll, fields.Everything())
	c.epIndexer, c.epInformer = cache.NewIndexerInformer(endpointWatcher, &corev1.Endpoints{}, 0,
		cache.ResourceEventHandlerFuncs{}, cache.Indexers{})

	c.syncFuncs = append(c.syncFuncs, c.svcInformer.HasSynced, c.epInformer.HasSynced, c.bgppeerInformer.HasSynced)

	return c
}

func (c *SpeakerController) Run(ctx context.Context, workers int) {
	defer runtime.HandleCrash()
	defer c.queue.ShutDown()

	klog.Info("Starting BGP speaker controller")
	defer klog.Info("Shutting down BGP speaker controller")

	go c.bgpServer.Serve()
	c.crdFactory.Start(ctx.Done())
	go c.svcInformer.Run(ctx.Done())
	go c.epInformer.Run(ctx.Done())

	if !cache.WaitForNamedCacheSync("bgp-speaker", ctx.Done(), c.syncFuncs...) {
		return
	}

	klog.Info("cache is synced")

	defer c.bgpServer.Stop()

	for i := 0; i < workers; i++ {
		go wait.UntilWithContext(ctx, c.worker, time.Second)
	}

	// 参考自 kube-router，使用周期函数来确保所有 BGP advertise LoadBalancer service ingress ip
	go wait.UntilWithContext(ctx, c.advertiseAllServiceIP, time.Second*60)

	<-ctx.Done()
}

func (c *SpeakerController) worker(ctx context.Context) {
	for c.processNextWorkItem(ctx) {
	}
}

func (c *SpeakerController) processNextWorkItem(ctx context.Context) bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	var err error
	switch t := key.(type) {
	case svcKey:
		if !c.peerConnected { // wait for create BGPPeer
			c.queue.AddAfter(key, time.Second*5)
			break
		}
		err = c.syncService(ctx, string(t))
	case bgppeer:
		err = c.syncBgpPeer(ctx, string(t))
	}

	if err == nil {
		c.queue.Forget(key)
		return true
	} else {
		runtime.HandleError(fmt.Errorf("error processing %v err:%v", key, err))
	}

	return true
}

func (c *SpeakerController) syncBgpPeer(ctx context.Context, key string) error {
	c.Lock()
	defer c.Unlock()

	bgpp, err := c.bgpPeerLister.Get(key)
	if err != nil {
		return err
	}

	var listenPort int32
	var sourceAddress string
	if !c.bgpServerStarted {
		if bgpp.Spec.SourcePort == 0 {
			listenPort = DefaultBGPPort
		} else {
			listenPort = int32(bgpp.Spec.SourcePort)
		}
		if len(bgpp.Spec.SourceAddress) == 0 {
			sourceAddress = c.nodeIP.String()
		} else {
			sourceAddress = bgpp.Spec.SourceAddress
		}

		if err = c.bgpServer.StartBgp(ctx, &gobgpapi.StartBgpRequest{
			Global: &gobgpapi.Global{
				Asn:             uint32(bgpp.Spec.MyAsn),
				RouterId:        sourceAddress,
				ListenPort:      listenPort,
				ListenAddresses: []string{sourceAddress},
			},
		}); err != nil {
			return err
		} else {
			// add import route policy
			// - inject any route advertised from peer
			/*err = c.bgpServer.AddPolicyAssignment(ctx, &gobgpapi.AddPolicyAssignmentRequest{
				Assignment: &gobgpapi.PolicyAssignment{
					Name:          "global",
					Direction:     gobgpapi.PolicyDirection_IMPORT,
					DefaultAction: gobgpapi.RouteAction_REJECT,
				},
			})
			if err != nil {
				return err
			}*/
		}
	}

	var remotePort uint32
	if bgpp.Spec.PeerPort == 0 {
		remotePort = DefaultBGPPort
	} else {
		remotePort = uint32(bgpp.Spec.PeerPort)
	}
	err = c.bgpServer.AddPeer(ctx, &gobgpapi.AddPeerRequest{
		Peer: &gobgpapi.Peer{
			Conf: &gobgpapi.PeerConf{
				NeighborAddress: bgpp.Spec.PeerAddress,
				PeerAsn:         uint32(bgpp.Spec.PeerAsn),
			},
			EbgpMultihop: &gobgpapi.EbgpMultihop{ // https://github.com/osrg/gobgp/blob/master/docs/sources/ebgp-multihop.md
				Enabled:     true,
				MultihopTtl: 5,
			},
			Transport: &gobgpapi.Transport{
				//LocalAddress: c.nodeIP.String(),
				RemotePort: remotePort,
			},
			GracefulRestart: &gobgpapi.GracefulRestart{ // https://github.com/osrg/gobgp/blob/master/docs/sources/graceful-restart.md
				Enabled:      true,
				RestartTime:  uint32((120 * time.Second).Seconds()), // route server 会在 RestartTime 之后 withdraw route
				DeferralTime: uint32((300 * time.Second).Seconds()),
				//LocalRestarting: true, // 如果打开，则需要 DeferralTime 后才会发送 BGP Update 报文，路由信息
			},
			AfiSafis: []*gobgpapi.AfiSafi{
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
			},
		},
	})

	if err == nil {
		klog.Infof(fmt.Sprintf("add BGPPeer neighbor %s:%d asn:%d for local %s:%d asn:%d",
			bgpp.Spec.PeerAddress, remotePort, bgpp.Spec.PeerAsn, sourceAddress, listenPort, bgpp.Spec.MyAsn))

		c.bgpServerStarted = true
	}

	// check peer established
	go wait.UntilWithContext(ctx, func(ctx context.Context) {
		peerConnected, err := c.isPeerEstablished(ctx, bgpp.Spec.PeerAddress)
		if err != nil {
			klog.Errorf(fmt.Sprintf("peer %s established err:%v", bgpp.Spec.PeerAddress, err))
			return
		}

		c.peerConnected = peerConnected
	}, time.Second)

	return err
}

func (c *SpeakerController) isPeerEstablished(ctx context.Context, peerIP string) (bool, error) {
	var peerConnected bool
	err := c.bgpServer.ListPeer(ctx, &gobgpapi.ListPeerRequest{Address: peerIP}, func(peer *gobgpapi.Peer) {
		if peer.Conf.NeighborAddress == peerIP && peer.State.SessionState == gobgpapi.PeerState_ESTABLISHED {
			peerConnected = true
		}
	})
	if err != nil {
		return false, err
	}

	return peerConnected, nil
}

func (c *SpeakerController) syncService(ctx context.Context, key string) error {
	service, exists, err := c.svcIndexer.GetByKey(key)
	switch {
	case !exists:
		// 不会存在这种情况，这里保留下代码，主要学习 switch 这种代码用法
	case err != nil:
		runtime.HandleError(fmt.Errorf("unable to retrieve service %v from store: %v", key, err))
	default:
		svc := service.(*corev1.Service)
		err = c.processServiceCreateOrUpdate(ctx, svc, key)
	}

	return err
}

func (c *SpeakerController) processServiceCreateOrUpdate(ctx context.Context, service *corev1.Service, key string) error {
	ip := service.Status.LoadBalancer.Ingress[0].IP
	ok := c.shouldAdvertiseService(service)
	if !ok {
		if c.isIPAdvertised(ip) {
			defer klog.Infof(fmt.Sprintf("withdraw route: %s/32 via nextHop %s", ip, c.nodeIP.String()))
			return c.withdrawIP(ctx, ip)
		}

		return nil
	}

	if !c.isIPAdvertised(ip) {
		defer klog.Infof(fmt.Sprintf("advertise route: %s/32 via nextHop %s", ip, c.nodeIP.String()))
		return c.advertiseIP(ctx, ip)
	}

	return nil
}

func (c *SpeakerController) shouldAdvertiseService(svc *corev1.Service) bool {
	if svc.Spec.ExternalTrafficPolicy != corev1.ServiceExternalTrafficPolicyTypeLocal {
		return true
	}

	ok, err := c.hasLocalEndpointsForService(svc)
	if err != nil || !ok {
		return false
	}

	return true
}

// INFO: 如果这个 service 是 ServiceExternalTrafficPolicyTypeLocal，那 bgp speaker 和 endpoint 在一个 node 上， 该 service ip 才会被宣告
func (c *SpeakerController) hasLocalEndpointsForService(svc *corev1.Service) (bool, error) {
	key, err := cache.MetaNamespaceKeyFunc(svc)
	if err != nil {
		return false, err
	}
	obj, exists, err := c.epIndexer.GetByKey(key)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, fmt.Errorf("endpoint resource doesn't exist for service: %q", svc.Name)
	}
	endpoints := obj.(*corev1.Endpoints)
	for _, subset := range endpoints.Subsets {
		for _, address := range subset.Addresses {
			if address.NodeName != nil {
				if *address.NodeName == c.nodeName {
					return true, nil
				}
			} else {
				if address.IP == c.nodeIP.String() {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

func (c *SpeakerController) isIPAdvertised(ip string) bool {
	existed := false
	err := c.bgpServer.ListPath(context.Background(), &gobgpapi.ListPathRequest{
		TableType: gobgpapi.TableType_GLOBAL,
		Family:    &gobgpapi.Family{Afi: gobgpapi.Family_AFI_IP, Safi: gobgpapi.Family_SAFI_UNICAST},
		Prefixes: []*gobgpapi.TableLookupPrefix{
			{
				Prefix: ip,
			},
		},
	}, func(destination *gobgpapi.Destination) {
		for _, path := range destination.Paths {
			if getNextHop(path) == c.nodeIP.String() {
				existed = true
			}
		}
	})

	return err == nil && existed
}

func (c *SpeakerController) advertiseIP(ctx context.Context, ip string) error {
	a1, _ := apb.New(&gobgpapi.OriginAttribute{
		Origin: uint32(bgppacket.BGP_ORIGIN_ATTR_TYPE_IGP),
	})
	a2, _ := apb.New(&gobgpapi.NextHopAttribute{
		NextHop: c.nodeIP.String(),
	})
	attrs := []*any.Any{a1, a2}
	nlri, _ := apb.New(&gobgpapi.IPAddressPrefix{
		Prefix:    ip,
		PrefixLen: 32,
	})
	_, err := c.bgpServer.AddPath(ctx, &gobgpapi.AddPathRequest{
		TableType: gobgpapi.TableType_GLOBAL,
		Path: &gobgpapi.Path{
			Family: &gobgpapi.Family{Afi: gobgpapi.Family_AFI_IP, Safi: gobgpapi.Family_SAFI_UNICAST},
			Nlri:   nlri,
			Pattrs: attrs,
		},
	})

	return err
}

func (c *SpeakerController) withdrawIP(ctx context.Context, ip string) error {
	a1, _ := apb.New(&gobgpapi.OriginAttribute{
		Origin: 0,
	})
	a2, _ := apb.New(&gobgpapi.NextHopAttribute{
		NextHop: c.nodeIP.String(),
	})
	attrs := []*any.Any{a1, a2}
	nlri, _ := apb.New(&gobgpapi.IPAddressPrefix{
		Prefix:    ip,
		PrefixLen: 32,
	})
	return c.bgpServer.DeletePath(ctx, &gobgpapi.DeletePathRequest{
		TableType: gobgpapi.TableType_GLOBAL,
		Path: &gobgpapi.Path{
			Family: &gobgpapi.Family{Afi: gobgpapi.Family_AFI_IP, Safi: gobgpapi.Family_SAFI_UNICAST},
			Nlri:   nlri,
			Pattrs: attrs,
		},
	})
}

func (c *SpeakerController) advertiseAllServiceIP(ctx context.Context) {
	if !c.peerConnected {
		return
	}

	count := 0
	klog.Infof(fmt.Sprintf("advertise service ip in period start..."))
	for _, obj := range c.svcIndexer.List() {
		svc := obj.(*corev1.Service)
		if svc.Spec.Type != corev1.ServiceTypeLoadBalancer || len(svc.Status.LoadBalancer.Ingress) == 0 {
			continue
		}

		if !c.shouldAdvertiseService(svc) {
			continue
		}

		ip := svc.Status.LoadBalancer.Ingress[0].IP
		if err := c.advertiseIP(ctx, ip); err != nil {
			klog.Errorf(fmt.Sprintf("advertise service %s/%s ip %s err:%v", svc.Namespace, svc.Name, ip, err))
			continue
		}

		count++
		klog.Infof(fmt.Sprintf("advertise service %s/%s ip %s via nextHop %s", svc.Namespace, svc.Name, ip, c.nodeIP.String()))
	}

	klog.Infof(fmt.Sprintf("advertise %d service ip in period end...", count))
}

// GetNodeIP returns the most valid external facing IP address for a node.
// Order of preference:
// 1. NodeInternalIP
// 2. NodeExternalIP (Only set on cloud providers usually)
func GetNodeIP(node *corev1.Node) (net.IP, error) {
	addresses := node.Status.Addresses
	addressMap := make(map[corev1.NodeAddressType][]corev1.NodeAddress)
	for i := range addresses {
		addressMap[addresses[i].Type] = append(addressMap[addresses[i].Type], addresses[i])
	}
	if addr, ok := addressMap[corev1.NodeInternalIP]; ok {
		return net.ParseIP(addr[0].Address), nil
	}
	if addr, ok := addressMap[corev1.NodeExternalIP]; ok {
		return net.ParseIP(addr[0].Address), nil
	}
	return nil, fmt.Errorf("host IP unknown")
}

// INFO: ECMP(Equal Cost Multi-Path) 等价路由: 多条不同链路到达同一目的地址的网络环境，即同一个 dst 多个 next hop

func getNextHop(path *gobgpapi.Path) string {
	for _, pattr := range path.Pattrs {
		var msg ptypes.DynamicAny
		ptypes.UnmarshalAny(pattr, &msg)
		switch t := msg.Message.(type) {
		case *gobgpapi.NextHopAttribute:
			return t.NextHop
		}
	}

	return ""
}
