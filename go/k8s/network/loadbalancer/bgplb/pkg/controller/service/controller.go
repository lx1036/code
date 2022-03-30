package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	v1 "k8s-lx1036/k8s/network/loadbalancer/bgplb/pkg/apis/bgplb.k9s.io/v1"
	"k8s-lx1036/k8s/network/loadbalancer/bgplb/pkg/client/clientset/versioned"
	"k8s-lx1036/k8s/network/loadbalancer/bgplb/pkg/client/informers/externalversions"
	listerv1 "k8s-lx1036/k8s/network/loadbalancer/bgplb/pkg/client/listers/bgplb.k9s.io/v1"

	"github.com/cilium/ipam/service/ipallocator"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
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

type svcKey string
type ippoolKey string

type Controller struct {
	kubeClient kubernetes.Interface
	crdClient  *versioned.Clientset
	events     record.EventRecorder
	queue      workqueue.RateLimitingInterface

	svcIndexer  cache.Indexer
	svcInformer cache.Controller

	crdFactory     externalversions.SharedInformerFactory
	ippoolLister   listerv1.IPPoolLister
	ippoolInformer cache.SharedIndexInformer

	syncFuncs []cache.InformerSynced

	balancer *LoadBalancer
	backoff
}

func New(restConfig *restclient.Config) *Controller {
	kubeClient := kubernetes.NewForConfigOrDie(restConfig)
	crdClient := versioned.NewForConfigOrDie(restConfig)

	broadcaster := record.NewBroadcaster()
	broadcaster.StartStructuredLogging(0)
	broadcaster.StartRecordingToSink(&typedv1.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	recorder := broadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "service-ipam-controller"})
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	c := &Controller{
		kubeClient: kubeClient,
		crdClient:  crdClient,
		events:     recorder,
		queue:      queue,
	}

	c.crdFactory = externalversions.NewSharedInformerFactory(crdClient, 0)
	c.ippoolInformer = c.crdFactory.Bgplb().V1().IPPools().Informer()
	c.ippoolLister = c.crdFactory.Bgplb().V1().IPPools().Lister()
	c.ippoolInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				c.queue.Add(ippoolKey(key))
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldIPPool := oldObj.(*v1.IPPool)
			newIPPool := newObj.(*v1.IPPool)
			if oldIPPool.Spec.Cidr == newIPPool.Spec.Cidr { // only care about cidr change
				return
			}
			key, err := cache.MetaNamespaceKeyFunc(newObj)
			if err == nil {
				c.queue.Add(ippoolKey(key))
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				c.queue.Add(ippoolKey(key))
			}
		},
	})

	svcWatcher := cache.NewListWatchFromClient(kubeClient.CoreV1().RESTClient(), "services", corev1.NamespaceAll, fields.Everything())
	c.svcIndexer, c.svcInformer = cache.NewIndexerInformer(svcWatcher, &corev1.Service{}, 0, cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			switch t := obj.(type) {
			case *corev1.Service:
				return t.Spec.Type == corev1.ServiceTypeLoadBalancer // only watch LoadBalancer service
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

				if err := c.balancer.Release(svc); err != nil {
					klog.Errorf(fmt.Sprintf("%v", err))
				}
			},
		},
	}, cache.Indexers{})

	c.syncFuncs = append(c.syncFuncs, c.svcInformer.HasSynced, c.ippoolInformer.HasSynced)

	ippools, err := crdClient.BgplbV1().IPPools().List(context.TODO(), metav1.ListOptions{})
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
	go c.svcInformer.Run(ctx.Done())

	if !cache.WaitForNamedCacheSync("service", ctx.Done(), c.syncFuncs...) {
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
	case svcKey:
		err = c.syncService(ctx, string(t))
		if err == nil {
			c.queue.Forget(key)
			return true
		} else {
			if errors.Is(err, NoIPPoolErr) {
				c.queue.AddAfter(key, c.Duration())
			}
		}
	case ippoolKey:
		err = c.syncIPPool(ctx, string(t))
		if err == nil {
			c.queue.Forget(key)
			return true
		} else {
			c.queue.AddRateLimited(key)
		}
	}

	runtime.HandleError(fmt.Errorf("error processing %v (will retry): %v", key, err))
	return true
}

func (c *Controller) syncIPPool(ctx context.Context, key string) error {
	ippool, err := c.ippoolLister.Get(key)
	if err != nil {
		if apierrors.IsNotFound(err) {
			c.balancer.DeleteAllocator(key)
		} else {
			return err
		}
	}

	if len(ippool.Spec.Cidr) == 0 {
		return fmt.Errorf("ippool cidr is empty")
	}

	if err := c.balancer.AddAllocator(key, *ippool); err != nil {
		return err
	}

	// INFO: list all unallocated loadbalancer service, and allocate ip for it.
	go func() {
		for _, obj := range c.svcIndexer.List() {
			svc := obj.(*corev1.Service)
			if svc.Spec.Type == corev1.ServiceTypeLoadBalancer && len(svc.Status.LoadBalancer.Ingress) == 0 {
				name, _ := cache.MetaNamespaceKeyFunc(svc)
				if err = c.processServiceCreateOrUpdate(ctx, svc, name); err != nil {
					klog.Errorf(fmt.Sprintf("allocate ip for service:%c err:%v", name, err))
				}
			}
		}
	}()

	// update ippool status metadata
	objCopy := ippool.DeepCopy()
	allocator := c.balancer.GetAllocator(key)
	objCopy.Status.PoolSize = allocator.Free()
	objCopy.Status.Usage = allocator.Free()
	objCopy.Status.FirstIP = allocator.FirstIP().String()
	objCopy.Status.LastIP = allocator.LastIP().String()
	return c.updateIPPoolStatus(ctx, objCopy)
}

func (c *Controller) syncService(ctx context.Context, key string) error {
	startTime := time.Now()
	defer func() {
		klog.V(4).Infof("Finished syncing service %q (%v)", key, time.Since(startTime))
	}()

	// service holds the latest service info from apiserver
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

func (c *Controller) processServiceCreateOrUpdate(ctx context.Context, service *corev1.Service, key string) error {
	svc, err := c.balancer.Allocate(service, key)
	if err != nil {
		if errors.Is(err, NoIPPoolErr) {
			c.events.Event(service, corev1.EventTypeWarning, "NoIPPoolErr", fmt.Sprintf("%v", NoIPPoolErr))
		}

		if errors.Is(err, ipallocator.ErrFull) { // change to choose free ippool
			c.events.Event(service, corev1.EventTypeWarning, "IPPoolFullErr", fmt.Sprintf("%v", ipallocator.ErrFull))

			for ippoolName, allocator := range c.balancer.ListAllocators() {
				if allocator.IsFull() {
					continue
				}

				c.events.Event(service, corev1.EventTypeWarning, "IPPoolChange", fmt.Sprintf("choose ippool %c instead for service", ippoolName))
				newSvc := service.DeepCopy()
				newSvc.Annotations[svcIPPoolAnnotation] = ippoolName
				return c.patchService(service, newSvc)
			}
		}

		return err
	}

	if reflect.DeepEqual(service.Status, svc.Status) {
		klog.Infof(fmt.Sprintf("service status %c/%c no change", svc.Namespace, svc.Name))
		return nil
	}

	err = c.updateSvcStatus(ctx, svc)
	if err == nil {
		klog.Infof(fmt.Sprintf("allocate ip:%c for service:%c", svc.Status.LoadBalancer.Ingress[0].IP, key))
		c.events.Event(service, corev1.EventTypeNormal, "AllocateIP", fmt.Sprintf("allocate ip %c", svc.Status.LoadBalancer.Ingress[0].IP))
	}

	// update ippool.status.used
	go func() {
		ippoolName := c.balancer.getIPPoolNameByService(svc)
		ippool, err := c.crdClient.BgplbV1().IPPools().Get(ctx, ippoolName, metav1.GetOptions{})
		if err != nil {
			klog.Errorf(fmt.Sprintf("%v", err))
			return
		}
		ippool.Status.Usage = ippool.Status.Usage - 1
		if ippool.Status.Usage == 0 {
			klog.Warningf(fmt.Sprintf("ippool:%c is full at %c", ippool.Name, time.Now().String()))
		}
		if ippool.Status.Used == nil {
			ippool.Status.Used = make(map[string]string)
		}
		ippool.Status.Used[svc.Status.LoadBalancer.Ingress[0].IP] = key
		if err = c.updateIPPoolStatus(ctx, ippool); err != nil {
			klog.Errorf(fmt.Sprintf("%v", err))
			return
		}
	}()

	return err
}

func (c *Controller) updateSvcStatus(ctx context.Context, service *corev1.Service) error {
	_, err := c.kubeClient.CoreV1().Services(service.Namespace).UpdateStatus(ctx, service, metav1.UpdateOptions{})
	return err
}

func (c *Controller) updateIPPoolStatus(ctx context.Context, ippool *v1.IPPool) error {
	_, err := c.crdClient.BgplbV1().IPPools().UpdateStatus(ctx, ippool, metav1.UpdateOptions{})
	return err
}

func (c *Controller) patchService(oldSvc, newSvc *corev1.Service) error {
	key, _ := cache.MetaNamespaceKeyFunc(oldSvc)
	oldData, err := json.Marshal(oldSvc)
	if err != nil {
		return fmt.Errorf("failed to marshal the existing service %s err: %v", key, err)
	}

	newData, err := json.Marshal(newSvc)
	if err != nil {
		return fmt.Errorf("failed to marshal the new service %s err: %v", key, err)
	}
	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, &corev1.Service{})
	if err != nil {
		return fmt.Errorf("failed to create a two-way merge patch: %v", err)
	}
	if _, err := c.kubeClient.CoreV1().Services(oldSvc.Namespace).Patch(context.TODO(), oldSvc.Name, types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{}); err != nil {
		return fmt.Errorf("failed to patch the service: %v", err)
	}

	return nil
}

// @see k8s.io/cloud-provider@v0.23.4/controllers/service/controller.go
func wantsLoadBalancer(service *corev1.Service) bool {
	// if LoadBalancerClass is set, the user does not want the default cloud-provider Load Balancer
	return service.Spec.Type == corev1.ServiceTypeLoadBalancer && service.Spec.LoadBalancerClass == nil
}
