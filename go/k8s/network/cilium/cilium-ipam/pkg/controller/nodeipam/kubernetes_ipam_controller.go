package nodeipam

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"reflect"
	"sync"
	"time"

	apiv1 "k8s-lx1036/k8s/network/cilium/cilium-ipam/pkg/apis/ipam.k9s.io/v1"
	"k8s-lx1036/k8s/network/cilium/cilium-ipam/pkg/client/clientset/versioned"
	"k8s-lx1036/k8s/network/cilium/cilium-ipam/pkg/client/informers/externalversions"
	"k8s-lx1036/k8s/network/cilium/cilium-ipam/pkg/client/listers/ipam.k9s.io/v1"
	"k8s-lx1036/k8s/network/cilium/cilium-ipam/pkg/ipam/allocator/clusterpool"

	ipamTypes "github.com/cilium/cilium/pkg/ipam/types"
	ciliumAPIV2 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2"
	ciliumClientSet "github.com/cilium/cilium/pkg/k8s/client/clientset/versioned"
	"github.com/cilium/ipam/cidrset"
	"github.com/projectcalico/calico/libcalico-go/lib/selector/tokenizer"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	listerv1 "k8s.io/client-go/listers/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	clientretry "k8s.io/client-go/util/retry"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

const (
	ipamK8s = "kubernetes"
	ipamCrd = "crd"

	ipThreshold = 5
)

type nodeKey string
type ippoolKey string

type NodeIPAMController struct {
	lock              sync.Mutex
	nodesInProcessing sets.String

	kubeClient   kubernetes.Interface
	crdClient    *versioned.Clientset
	ciliumClient *ciliumClientSet.Clientset
	events       record.EventRecorder
	queue        workqueue.RateLimitingInterface

	crdFactory     externalversions.SharedInformerFactory
	ippoolLister   v1.IPPoolLister
	ippoolInformer cache.SharedIndexInformer

	//ciliumFactory      ciliumExternalversions.SharedInformerFactory
	ciliumNodeInformer cache.SharedIndexInformer
	//ciliumNodeLister   ciliumListerV2.CiliumNodeLister

	nodeInformer cache.SharedIndexInformer
	nodeLister   listerv1.NodeLister

	syncFuncs []cache.InformerSynced

	balancer *LoadBalancer
	ipam     string
}

// INFO:
//  方案一：ipam.mode=kubernetes，且 kube-controller-manager allocate-node-cidrs=false，然后根据 nodeSelector 选择特定的 ippool，
//  再去 annotation node "io.cilium.network.ipv4-pod-cidr: 100.20.30.0/24"，缺点是：每一个 node 只有一个 pod cidr，不能动态扩容；优点是：实现简单
//  方案二：ipam.mode=crd, 且 cilium daemon 需要开启 enable-endpoint-routes: 'true'(这个机制不是最优的)，每一个 pod 一条路由。优点是：可以配置多个 pod cidr，缺点是：实现复杂，不太好弄

func New(restConfig *restclient.Config, ipam string) *NodeIPAMController {
	kubeClient := kubernetes.NewForConfigOrDie(restConfig)
	crdClient := versioned.NewForConfigOrDie(restConfig)
	ciliumClient := ciliumClientSet.NewForConfigOrDie(restConfig)

	broadcaster := record.NewBroadcaster()
	broadcaster.StartRecordingToSink(&typedv1.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	recorder := broadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "node-ipam-controller"})
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	c := &NodeIPAMController{
		kubeClient:        kubeClient,
		crdClient:         crdClient,
		ciliumClient:      ciliumClient,
		events:            recorder,
		queue:             queue,
		nodesInProcessing: sets.NewString(),

		ipam: ipam,
	}

	c.crdFactory = externalversions.NewSharedInformerFactory(crdClient, 0)
	c.ippoolInformer = c.crdFactory.Ipam().V1().IPPools().Informer()
	c.ippoolLister = c.crdFactory.Ipam().V1().IPPools().Lister()
	c.ippoolInformer.AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			switch t := obj.(type) {
			case *apiv1.IPPool:
				if len(t.Spec.Cidr) == 0 || len(t.Spec.NodeSelectors) == 0 {
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
			DeleteFunc: func(obj interface{}) {
				key, err := cache.MetaNamespaceKeyFunc(obj)
				if err == nil {
					c.queue.Add(ippoolKey(key))
				}
			},
		},
	})

	/*c.ciliumFactory = ciliumExternalversions.NewSharedInformerFactory(ciliumClient, 0)
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
	})*/

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
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldNode := oldObj.(*corev1.Node)
			newNode := newObj.(*corev1.Node)
			if newNode.Annotations != nil && len(newNode.Annotations[ipv4PodCidr]) != 0 {
				if !reflect.DeepEqual(oldNode.Labels, newNode.Labels) {
					cn, err := c.ciliumClient.CiliumV2().CiliumNodes().Get(context.TODO(), newNode.Name, metav1.GetOptions{})
					if err != nil {
						klog.Errorf(fmt.Sprintf("get CiliumNode %s err:%v", newNode.Name, err))
						return
					}

					cnCopy := cn.DeepCopy()
					cnCopy.Labels = newNode.Labels
					if err = c.patchCiliumNode(context.TODO(), cn, cnCopy); err != nil {
						klog.Errorf(fmt.Sprintf("patch CiliumNode %s labels err:%v", newNode.Name, err))
						return
					} else {
						klog.Infof(fmt.Sprintf("patch CiliumNode %s labels", newNode.Name))
					}
				}

				return
			}

			key, err := cache.MetaNamespaceKeyFunc(newObj)
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

	/*ippools, err := crdClient.IpamV1().IPPools().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Fatal(err)
	}*/
	var err error
	c.balancer, err = NewLoadBalancer([]apiv1.IPPool{})
	if err != nil {
		klog.Fatal(err)
	}

	return c
}

func (c *NodeIPAMController) Run(ctx context.Context, workers int) {
	defer runtime.HandleCrash()
	defer c.queue.ShutDown()

	klog.Info("Starting node ipam controller")
	defer klog.Info("Shutting down node ipam controller")

	c.crdFactory.Start(ctx.Done())
	go c.nodeInformer.Run(ctx.Done())
	//go c.ciliumNodeInformer.Run(ctx.Done())

	if !cache.WaitForNamedCacheSync("node-ipam-controller", ctx.Done(), c.syncFuncs...) {
		return
	}

	for i := 0; i < workers; i++ {
		go wait.UntilWithContext(ctx, c.worker, time.Second)
	}

	if c.ipam == ipamCrd {
		// check if podCIDR is full, 否则添加新的 podCIDR
		go wait.UntilWithContext(ctx, c.checkInsufficient, time.Second*60)
	}

	<-ctx.Done()
}

func (c *NodeIPAMController) checkInsufficient(ctx context.Context) {
	nodes, err := c.ciliumClient.CiliumV2().CiliumNodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Errorf(fmt.Sprintf("list CiliumNodes err:%v", err))
		return
	}

	for _, cn := range nodes.Items {
		used := len(cn.Status.IPAM.Used)
		total := len(cn.Spec.IPAM.Pool)
		if total-used < ipThreshold {
			node, err := c.kubeClient.CoreV1().Nodes().Get(ctx, cn.Name, metav1.GetOptions{})
			if err != nil {
				continue
			}

			alloc, err := c.balancer.getAllocatorByNode(node)
			if err != nil {
				continue
			}

			ipnet, err := alloc.allocator.AllocateNext()
			if err != nil {
				if errors.Is(err, cidrset.ErrCIDRRangeNoCIDRsRemaining) {
					klog.Errorf(fmt.Sprintf("IPPool %s has no remaining cidr", alloc.ippool.Name))
				}
				continue
			}

			cnCopy := cn.DeepCopy()
			cnCopy.Spec.IPAM.PodCIDRs = append(cnCopy.Spec.IPAM.PodCIDRs, ipnet.String())
			allocation := make(ipamTypes.AllocationMap)
			for _, podCidr := range cnCopy.Spec.IPAM.PodCIDRs {
				_, ipnet, _ := net.ParseCIDR(podCidr)
				_ = clusterpool.ForEachIP(*ipnet, func(ip string) error {
					allocation[ip] = ipamTypes.AllocationIP{}
					return nil
				})
			}
			cnCopy.Spec.IPAM.Pool = allocation
			err = c.patchCiliumNode(ctx, &cn, cnCopy)
			if err != nil {
				klog.Errorf(fmt.Sprintf("patch CiliumNode %s ipam podcidr err:%v", cn.Name, err))
				continue
			}

			klog.Infof(fmt.Sprintf("add new podcidr %s for CiliumNode %s", ipnet.String(), cn.Name))
		}
	}
}

func (c *NodeIPAMController) worker(ctx context.Context) {
	for c.processNextWorkItem(ctx) {
	}
}

func (c *NodeIPAMController) processNextWorkItem(ctx context.Context) bool {
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
		//err = c.syncNode(ctx, string(t))
		err = c.syncK8sNode(ctx, string(t))
		if errors.Is(err, NoIPPoolErr) {
			// wait for added ippool
			// (或者也可以不考虑后加了 ippool 能适配 node labels 这种情况，需要用户重新更新 node labels)
			c.queue.AddAfter(key, time.Second*5)
			return true
		}
	}

	if err != nil {
		runtime.HandleError(fmt.Errorf("error processing %v (will retry): %v", key, err))
	}
	return true
}

func (c *NodeIPAMController) insertNodeToProcessing(nodeName string) bool {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.nodesInProcessing.Has(nodeName) {
		return false
	}
	c.nodesInProcessing.Insert(nodeName)
	return true
}

func (c *NodeIPAMController) removeNodeFromProcessing(nodeName string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.nodesInProcessing.Delete(nodeName)
}

func (c *NodeIPAMController) syncIPPool(ctx context.Context, key string) error {
	ippool, err := c.ippoolLister.Get(key)
	if err != nil {
		if apierrors.IsNotFound(err) {
			c.balancer.DeleteAllocator(key)
			klog.Infof(fmt.Sprintf("delete ippool %s", key))
			return nil
		} else {
			return err
		}
	}

	return c.balancer.AddAllocator(key, *ippool)
}

func (c *NodeIPAMController) syncK8sNode(ctx context.Context, key string) error {
	node, err := c.nodeLister.Get(key)
	if err != nil {
		return err
	}

	if !c.insertNodeToProcessing(node.Name) {
		klog.Infof(fmt.Sprintf("node %s is already in a process of CIDR assignment ", key))
		return nil
	}
	defer c.removeNodeFromProcessing(node.Name)

	newNode, err := c.balancer.Allocate(node, key)
	if err != nil {
		if errors.Is(err, NoIPPoolErr) {
			c.events.Event(node, corev1.EventTypeWarning, "NoIPPoolErr", fmt.Sprintf("%v", NoIPPoolErr))
			klog.Errorf(fmt.Sprintf("choose no ippool base on NodeSelector for node:%s", key))
			return err
		}

		if errors.Is(err, cidrset.ErrCIDRRangeNoCIDRsRemaining) {
			// 如果当前 ippool is full, change other free ippool
			alloc, _ := c.balancer.getAllocatorByNode(node)
			klog.Errorf(fmt.Sprintf("IPPool %s has no remaining cidr", alloc.ippool.Name))
			c.events.Event(node, corev1.EventTypeWarning, "IPPoolFullErr", fmt.Sprintf("%v", cidrset.ErrCIDRRangeNoCIDRsRemaining))
			for ippoolName, allocator := range c.balancer.ListAllocators() {
				if allocator.allocator.IsFull() {
					continue
				}

				label, value := ParseLabelValue(allocator.ippool.Spec.NodeSelectors)
				if len(label) != 0 && len(value) != 0 { // 只考虑类似 "key=='value'" 的 nodeSelector
					c.events.Event(node, corev1.EventTypeWarning, "IPPoolChange", fmt.Sprintf("choose ippool %s instead for node", ippoolName))
					newNode = node.DeepCopy()
					if newNode.Labels == nil {
						newNode.Labels = make(map[string]string)
					}
					newNode.Labels[label] = value
					return patchK8sNode(ctx, c.kubeClient, node, newNode)
				}
			}
		}

		return err
	}

	if reflect.DeepEqual(node, newNode) {
		return nil
	}

	if node.Annotations != nil && node.Annotations[ipv4PodCidr] == newNode.Annotations[ipv4PodCidr] {
		klog.Infof(fmt.Sprintf("node %s annotation %s for %s has no change", key, ipv4PodCidr, newNode.Annotations[ipv4PodCidr]))
		return nil
	}

	if err = patchK8sNode(ctx, c.kubeClient, node, newNode); err == nil {
		klog.Infof(fmt.Sprintf("allocate cidr %s for node %s", newNode.Annotations[ipv4PodCidr], key))
		c.events.Event(node, corev1.EventTypeNormal, "AllocateCidr", fmt.Sprintf("allocate cidr %s", newNode.Annotations[ipv4PodCidr]))

		if c.ipam == ipamCrd {
			// patch CiliumNode pool
			cn, err := c.ciliumClient.CiliumV2().CiliumNodes().Get(ctx, newNode.Name, metav1.GetOptions{})
			if err != nil {
				if apierrors.IsNotFound(err) {
					err = c.createCiliumNode(ctx, newNode)
				}
				return err
			}

			cnCopy := cn.DeepCopy()
			cnCopy.Spec.IPAM.PodCIDRs = append(cnCopy.Spec.IPAM.PodCIDRs, newNode.Annotations[ipv4PodCidr])
			allocation := make(ipamTypes.AllocationMap)
			for _, podCidr := range cnCopy.Spec.IPAM.PodCIDRs {
				_, ipnet, _ := net.ParseCIDR(podCidr)
				_ = clusterpool.ForEachIP(*ipnet, func(ip string) error {
					allocation[ip] = ipamTypes.AllocationIP{}
					return nil
				})
			}
			cnCopy.Spec.IPAM.Pool = allocation
			return c.patchCiliumNode(ctx, cn, cnCopy)
		}

		return err
	}

	return err
}

func (c *NodeIPAMController) createCiliumNode(ctx context.Context, node *corev1.Node) error {
	_, ipnet, _ := net.ParseCIDR(node.Annotations[ipv4PodCidr])
	allocation := make(ipamTypes.AllocationMap)
	_ = clusterpool.ForEachIP(*ipnet, func(ip string) error {
		allocation[ip] = ipamTypes.AllocationIP{}
		return nil
	})
	cn := &ciliumAPIV2.CiliumNode{
		ObjectMeta: metav1.ObjectMeta{
			Name:   node.Name,
			Labels: node.Labels,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "v1",
					Kind:       "Node",
					Name:       node.Name,
					UID:        node.UID,
				},
			},
		},
		Spec: ciliumAPIV2.NodeSpec{
			IPAM: ipamTypes.IPAMSpec{
				Pool: allocation,
				PodCIDRs: []string{
					ipnet.String(),
				},
			},
		},
	}
	_, err := c.ciliumClient.CiliumV2().CiliumNodes().Create(ctx, cn, metav1.CreateOptions{})
	return err
}

func (c *NodeIPAMController) patchCiliumNode(ctx context.Context, oldCN, newCN *ciliumAPIV2.CiliumNode) error {
	key, _ := cache.MetaNamespaceKeyFunc(oldCN)
	oldData, err := json.Marshal(oldCN)
	if err != nil {
		return fmt.Errorf("failed to marshal the existing CiliumNode %s err: %v", key, err)
	}

	newData, err := json.Marshal(newCN)
	if err != nil {
		return fmt.Errorf("failed to marshal the new CiliumNode %s err: %v", key, err)
	}
	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, &ciliumAPIV2.CiliumNode{})
	if err != nil {
		return fmt.Errorf("failed to create a two-way merge patch: %v", err)
	}

	if _, err := c.ciliumClient.CiliumV2().CiliumNodes().Patch(ctx, oldCN.Name, types.MergePatchType,
		patchBytes, metav1.PatchOptions{}); err != nil {
		return fmt.Errorf("failed to patch the CiliumNode: %v", err)
	}

	return nil
}

func ParseLabelValue(selectors []*metav1.LabelSelector) (label, value string) {
	for _, selector := range selectors {
		sel, err := metav1.LabelSelectorAsSelector(selector)
		if err != nil {
			continue
		}

		requirements, _ := sel.Requirements()
		for _, requirement := range requirements {
			op := requirement.Operator()
			if op == selection.Equals || op == selection.DoubleEquals || op == selection.In {
				if len(requirement.Values().List()) != 0 {
					label = requirement.Key()
					value = requirement.Values().List()[0]
					return
				}
			}
		}
	}

	return
}

func ParseLabelValueWithCalico(input string) (label, value string) {
	tokens, err := tokenizer.Tokenize(input)
	if err != nil {
		return
	}

	for _, token := range tokens {
		if token.Kind == tokenizer.TokLabel {
			label = token.Value.(string)
		}
		if token.Kind == tokenizer.TokStringLiteral {
			value = token.Value.(string)
		}
	}
	return
}

func patchK8sNode(ctx context.Context, kubeClient kubernetes.Interface, oldNode, newNode *corev1.Node) error {
	var err error
	var oldNodeObj, newNodeObj *corev1.Node
	key, _ := cache.MetaNamespaceKeyFunc(oldNode)
	firstTry := true
	return clientretry.RetryOnConflict(clientretry.DefaultRetry, func() error {
		if firstTry {
			oldNodeObj = oldNode
			newNodeObj = newNode
		} else {
			oldNode, err = kubeClient.CoreV1().Nodes().Get(ctx, key, metav1.GetOptions{})
			if err != nil {
				return err
			}
			newestNode := oldNode.DeepCopy()
			if newestNode.Annotations == nil {
				newestNode.Annotations = make(map[string]string)
			}
			newestNode.Annotations[ipv4PodCidr] = newNode.Annotations[ipv4PodCidr]
			oldNodeObj = oldNode
			newNodeObj = newestNode
		}

		oldData, err := json.Marshal(oldNodeObj)
		if err != nil {
			return fmt.Errorf("failed to marshal the existing node %s err: %v", key, err)
		}
		newData, err := json.Marshal(newNodeObj)
		if err != nil {
			return fmt.Errorf("failed to marshal the new node %s err: %v", key, err)
		}
		patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, &corev1.Node{})
		if err != nil {
			return fmt.Errorf("failed to create a two-way merge patch: %v", err)
		}
		if _, err = kubeClient.CoreV1().Nodes().Patch(ctx, oldNode.Name, types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{}); err != nil {
			return fmt.Errorf("failed to patch the node: %v", err)
		}

		return nil
	})
}

//func (c *NodeIPAMController) syncNode(ctx context.Context, key string) error {
//	node, err := c.nodeLister.Get(key)
//	if err != nil {
//		return err
//	}
//
//	cn, err := c.ciliumClient.CiliumV2().CiliumNodes().Get(ctx, node.Name, metav1.GetOptions{})
//	if err != nil {
//		if apierrors.IsNotFound(err) {
//			return c.AllocateNode(ctx, node, key)
//		} else {
//			return err
//		}
//	}
//
//	pool, err := c.balancer.getAllocatorByNode(node)
//	if err != nil {
//		return err
//	}
//
//	var (
//		inRange    []string
//		notInRange []string
//	)
//	for _, podCidr := range cn.Spec.IPAM.PodCIDRs {
//		_, ipnet, _ := net.ParseCIDR(podCidr)
//		if pool.allocator.InRange(ipnet) {
//			inRange = append(inRange, podCidr)
//		} else {
//			notInRange = append(notInRange, podCidr)
//		}
//	}
//
//	switch {
//	case len(notInRange) == 0:
//		return nil
//
//	case len(notInRange) != 0 && len(inRange) == 0:
//		ipnet, err := c.balancer.Allocate(node, key)
//		if err != nil {
//			return err
//		}
//		inRange = []string{ipnet.String()}
//	}
//
//	cnCopy := cn.DeepCopy()
//	cnCopy.Spec.IPAM.PodCIDRs = inRange
//	inRangePool := make(ipamTypes.AllocationMap)
//	for _, in := range inRange {
//		_, ipnet, _ := net.ParseCIDR(in)
//		_ = clusterpool.ForEachIP(*ipnet, func(ip string) error {
//			inRangePool[ip] = ipamTypes.AllocationIP{}
//			return nil
//		})
//	}
//	cnCopy.Spec.IPAM.Pool = inRangePool
//	if err = c.patchCiliumNode(cn, cnCopy); err != nil {
//		return err
//	}
//
//	klog.Infof(fmt.Sprintf("allocate pod cidr %s for CiliumNode:%s", strings.Join(cnCopy.Spec.IPAM.PodCIDRs, ","), key))
//	//c.events.Event(cn, corev1.EventTypeNormal, "AllocateCidr", fmt.Sprintf("allocate cidr %s", strings.Join(cnCopy.Spec.IPAM.PodCIDRs, ",")))
//	return nil
//}

//func (c *NodeIPAMController) AllocateNode(ctx context.Context, node *corev1.Node, key string) error {
//	// allocate node subnet from specified ippool, and create CiliumNode
//	ipnet, err := c.balancer.Allocate(node, key)
//	if err != nil {
//		return err
//	}
//
//	pool := make(ipamTypes.AllocationMap)
//	/*_ = clusterpool.ForEachIP(*ipnet, func(ip string) error {
//		pool[ip] = ipamTypes.AllocationIP{}
//		return nil
//	})*/
//
//	cn := &ciliumAPIV2.CiliumNode{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:   node.Name,
//			Labels: node.Labels,
//			OwnerReferences: []metav1.OwnerReference{
//				{
//					APIVersion: "v1",
//					Kind:       "Node",
//					Name:       node.Name,
//					UID:        node.UID,
//				},
//			},
//		},
//		Spec: ciliumAPIV2.NodeSpec{
//			IPAM: ipamTypes.IPAMSpec{
//				Pool: pool,
//				PodCIDRs: []string{
//					ipnet.String(),
//				},
//			},
//		},
//	}
//	cn, err = c.ciliumClient.CiliumV2().CiliumNodes().Create(ctx, cn, metav1.CreateOptions{})
//	if err != nil {
//		c.balancer.Release(node)
//		return err
//	}
//
//	//c.events.Event(cn, corev1.EventTypeNormal, "AllocateCidr", fmt.Sprintf("allocate cidr %s", ipnet.String()))
//	return nil
//}
