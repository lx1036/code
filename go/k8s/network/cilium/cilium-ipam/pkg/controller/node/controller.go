package node

import (
	"context"
	"fmt"
	"k8s-lx1036/k8s/network/cilium/cilium-ipam/pkg/ipam/allocator/clusterpool"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"net"
	"time"

	apiv1 "k8s-lx1036/k8s/network/cilium/cilium-ipam/pkg/apis/ipam.k9s.io/v1"
	"k8s-lx1036/k8s/network/cilium/cilium-ipam/pkg/client/clientset/versioned"
	"k8s-lx1036/k8s/network/cilium/cilium-ipam/pkg/client/informers/externalversions"
	"k8s-lx1036/k8s/network/cilium/cilium-ipam/pkg/client/listers/ipam.k9s.io/v1"

	ipamTypes "github.com/cilium/cilium/pkg/ipam/types"
	ciliumAPIV2 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2"
	ciliumClientSet "github.com/cilium/cilium/pkg/k8s/client/clientset/versioned"
	ciliumExternalversions "github.com/cilium/cilium/pkg/k8s/client/informers/externalversions"
	ciliumListerV2 "github.com/cilium/cilium/pkg/k8s/client/listers/cilium.io/v2"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	listerv1 "k8s.io/client-go/listers/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

type nodeKey string
type ippoolKey string

type Controller struct {
	kubeClient   kubernetes.Interface
	crdClient    *versioned.Clientset
	ciliumClient *ciliumClientSet.Clientset
	events       record.EventRecorder
	queue        workqueue.RateLimitingInterface

	crdFactory     externalversions.SharedInformerFactory
	ippoolLister   v1.IPPoolLister
	ippoolInformer cache.SharedIndexInformer

	ciliumFactory      ciliumExternalversions.SharedInformerFactory
	ciliumNodeInformer cache.SharedIndexInformer
	ciliumNodeLister   ciliumListerV2.CiliumNodeLister

	nodeInformer cache.SharedIndexInformer
	nodeLister   listerv1.NodeLister

	syncFuncs []cache.InformerSynced

	balancer *LoadBalancer
}

func New(restConfig *restclient.Config) *Controller {
	kubeClient := kubernetes.NewForConfigOrDie(restConfig)
	crdClient := versioned.NewForConfigOrDie(restConfig)
	ciliumClient := ciliumClientSet.NewForConfigOrDie(restConfig)

	broadcaster := record.NewBroadcaster()
	broadcaster.StartStructuredLogging(0)
	broadcaster.StartRecordingToSink(&typedv1.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	recorder := broadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "ipam-controller"})
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	c := &Controller{
		kubeClient:   kubeClient,
		crdClient:    crdClient,
		ciliumClient: ciliumClient,
		events:       recorder,
		queue:        queue,
	}

	c.crdFactory = externalversions.NewSharedInformerFactory(crdClient, 0)
	c.ippoolInformer = c.crdFactory.Ipam().V1().IPPools().Informer()
	c.ippoolLister = c.crdFactory.Ipam().V1().IPPools().Lister()
	c.ippoolInformer.AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			switch t := obj.(type) {
			case *apiv1.IPPool:
				if len(t.Spec.Cidr) == 0 || len(t.Spec.NodeSelector) == 0 {
					return false
				}
				_, _, err := net.ParseCIDR(t.Spec.Cidr)
				if err != nil {
					klog.Errorf(fmt.Sprintf("ippool:%s cidr is err:%v", t.Name, err))
					return false
				}
				return true
			default:
				runtime.HandleError(fmt.Errorf("object passed to %T that is not expected: %T", c, obj))
				return false
			}
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				key, err := cache.MetaNamespaceKeyFunc(obj)
				if err == nil {
					c.queue.Add(ippoolKey(key))
				}
			},
			UpdateFunc: nil,
			DeleteFunc: func(obj interface{}) {
				key, err := cache.MetaNamespaceKeyFunc(obj)
				if err == nil {
					c.queue.Add(ippoolKey(key))
				}
			},
		},
	})

	c.ciliumFactory = ciliumExternalversions.NewSharedInformerFactory(ciliumClient, 0)
	c.ciliumNodeInformer = c.ciliumFactory.Cilium().V2().CiliumNodes().Informer()
	c.ciliumNodeLister = c.ciliumFactory.Cilium().V2().CiliumNodes().Lister()
	c.ciliumNodeInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: nil,
		DeleteFunc: func(obj interface{}) {
			cn, ok := obj.(*ciliumAPIV2.CiliumNode)
			if !ok {
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					klog.Errorf("unexpected object type: %v", obj)
					return
				}
				if cn, ok = tombstone.Obj.(*ciliumAPIV2.CiliumNode); !ok {
					klog.Errorf("unexpected object type: %v", obj)
					return
				}
			}
			// TODO: release node cidr
			klog.Infof(fmt.Sprintf("CiliumNode %s is delete", cn.Name))
		},
	})

	factory := informers.NewSharedInformerFactory(kubeClient, 0)
	c.nodeInformer = factory.Core().V1().Nodes().Informer()
	c.nodeLister = factory.Core().V1().Nodes().Lister()
	c.nodeInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				c.queue.Add(nodeKey(key))
			}
		},
		DeleteFunc: func(obj interface{}) {
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

			if err := c.balancer.Release(node); err != nil {
				klog.Errorf(fmt.Sprintf("%v", err))
			}
		},
	})

	c.syncFuncs = append(c.syncFuncs, c.nodeInformer.HasSynced, c.ippoolInformer.HasSynced)

	ippools, err := crdClient.IpamV1().IPPools().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Fatal(err)
	}
	c.balancer, err = NewLoadBalancer(ippools.Items)
	if err != nil {
		klog.Fatal(err)
	}

	return c
}

func (c *Controller) Run(ctx context.Context, workers int) {
	defer runtime.HandleCrash()
	defer c.queue.ShutDown()

	klog.Info("Starting service controller")
	defer klog.Info("Shutting down service controller")

	c.crdFactory.Start(ctx.Done())
	go c.nodeInformer.Run(ctx.Done())
	go c.ciliumNodeInformer.Run(ctx.Done())

	if !cache.WaitForNamedCacheSync("ipam-controller", ctx.Done(), c.syncFuncs...) {
		return
	}

	for i := 0; i < workers; i++ {
		go wait.UntilWithContext(ctx, c.worker, time.Second)
	}

	<-ctx.Done()
}

func (c *Controller) worker(ctx context.Context) {
	for c.processNextWorkItem(ctx) {
	}
}

func (c *Controller) processNextWorkItem(ctx context.Context) bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	var err error
	switch t := key.(type) {
	case ippoolKey:
		err = c.syncIPPool(ctx, string(t))
	case nodeKey:
		err = c.syncNode(ctx, string(t))
	}

	runtime.HandleError(fmt.Errorf("error processing %v (will retry): %v", key, err))
	return true
}

func (c *Controller) syncIPPool(ctx context.Context, key string) error {
	ippool, err := c.ippoolLister.Get(key)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// TODO: delete ippool resource
			return nil
		} else {
			return err
		}
	}

	return c.balancer.AddAllocator(key, *ippool)
}

func (c *Controller) syncNode(ctx context.Context, key string) error {
	node, err := c.nodeLister.Get(key)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// TODO: delete ippool resource
			return nil
		} else {
			return err
		}
	}

	// allocate node subnet from specified ippool, and create CiliumNode
	ipnet, err := c.balancer.Allocate(node, key)
	if err != nil {
		return err
	}

	pool := make(ipamTypes.AllocationMap)
	_ = clusterpool.ForEachIP(*ipnet, func(ip string) error {
		pool[ip] = ipamTypes.AllocationIP{}
		return nil
	})

	cn := &ciliumAPIV2.CiliumNode{
		ObjectMeta: metav1.ObjectMeta{
			Name:   node.Name,
			Labels: node.Labels,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: node.APIVersion,
					Kind:       node.Kind,
					Name:       node.Name,
					UID:        node.UID,
				},
			},
		},
		Spec: ciliumAPIV2.NodeSpec{
			IPAM: ipamTypes.IPAMSpec{
				Pool: pool,
				PodCIDRs: []string{
					ipnet.String(),
				},
			},
		},
	}
	cn, err = c.ciliumClient.CiliumV2().CiliumNodes().Create(ctx, cn, metav1.CreateOptions{})
	if err != nil {
		c.balancer.Release(node)
		return err
	}

	c.events.Event(cn, corev1.EventTypeNormal, "AllocateCidr", fmt.Sprintf("allocate cidr %s", ipnet.String()))

	return nil
}
