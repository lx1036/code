package service

import (
	"context"
	"fmt"
	v1 "k8s-lx1036/k8s/network/loadbalancer/bgplb/pkg/client/listers/bgplb.k9s.io/v1"
	"k8s-lx1036/k8s/network/loadbalancer/bgplb/pkg/ipam"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/fields"
	"net"
	"time"

	"k8s-lx1036/k8s/network/loadbalancer/bgplb/pkg/client/clientset/versioned"
	"k8s-lx1036/k8s/network/loadbalancer/bgplb/pkg/client/informers/externalversions"

	corev1 "k8s.io/api/core/v1"
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

type svcKey string
type ippoolKey string

type Controller struct {
	kubeClient kubernetes.Interface
	events     record.EventRecorder
	queue      workqueue.RateLimitingInterface

	svcIndexer  cache.Indexer
	svcInformer cache.Controller

	crdFactory     externalversions.SharedInformerFactory
	ippoolLister   v1.IPPoolLister
	ippoolInformer cache.SharedIndexInformer

	syncFuncs []cache.InformerSynced

	balancer *LoadBalancer
}

func New(restConfig *restclient.Config) *Controller {
	kubeClient := kubernetes.NewForConfigOrDie(restConfig)
	crdClient := versioned.NewForConfigOrDie(restConfig)

	broadcaster := record.NewBroadcaster()
	broadcaster.StartStructuredLogging(0)
	broadcaster.StartRecordingToSink(&typedv1.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	recorder := broadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "service-controller"})
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	c := &Controller{
		kubeClient: kubeClient,
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
	c.svcIndexer, c.svcInformer = cache.NewIndexerInformer(svcWatcher, &corev1.Service{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if !wantsLoadBalancer(obj.(*corev1.Service)) {
				return
			}
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				c.queue.Add(svcKey(key))
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if !wantsLoadBalancer(newObj.(*corev1.Service)) {
				return
			}
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
			if !wantsLoadBalancer(svc) {
				return
			}

			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				c.queue.Add(svcKey(key))
			}
		},
	}, cache.Indexers{})

	c.syncFuncs = append(c.syncFuncs, c.svcInformer.HasSynced, c.ippoolInformer.HasSynced)

	return c
}

func (s *Controller) Run(ctx context.Context, workers int) {
	defer runtime.HandleCrash()
	defer s.queue.ShutDown()

	klog.Info("Starting service controller")
	defer klog.Info("Shutting down service controller")

	s.crdFactory.Start(ctx.Done())
	go s.svcInformer.Run(ctx.Done())

	if !cache.WaitForNamedCacheSync("service", ctx.Done(), s.syncFuncs...) {
		return
	}

	for i := 0; i < workers; i++ {
		go wait.UntilWithContext(ctx, s.worker, time.Second)
	}

	<-ctx.Done()
}

func (s *Controller) worker(ctx context.Context) {
	for s.processNextWorkItem(ctx) {
	}
}

func (s *Controller) processNextWorkItem(ctx context.Context) bool {
	key, quit := s.queue.Get()
	if quit {
		return false
	}
	defer s.queue.Done(key)

	var err error
	switch key.(type) {
	case svcKey:
		err = s.syncService(ctx, key.(string))
		if err == nil {
			s.queue.Forget(key)
			return true
		}
	case ippoolKey:
		err = s.syncIPPool(ctx, key.(string))
		if err == nil {
			s.queue.Forget(key)
			return true
		}
	}

	runtime.HandleError(fmt.Errorf("error processing %v (will retry): %v", key, err))
	s.queue.AddRateLimited(key)
	return true
}

func (s *Controller) syncIPPool(ctx context.Context, key string) error {
	ippool, err := s.ippoolLister.Get(key)
	if err != nil {
		if apierrors.IsNotFound(err) {
			s.balancer.DeleteAllocator(key)
		} else {
			return err
		}
	}

	_, cidr, err := net.ParseCIDR(ippool.Spec.Cidr)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("parse ippool:%s cidr:%s err:%v", key, ippool.Spec.Cidr, err))
	}
	allocator := ipam.NewHostScopeAllocator(cidr)
	s.balancer.AddAllocator(key, allocator)
	return nil
}

func (s *Controller) syncService(ctx context.Context, key string) error {
	startTime := time.Now()
	defer func() {
		klog.V(4).Infof("Finished syncing service %q (%v)", key, time.Since(startTime))
	}()

	// service holds the latest service info from apiserver
	service, exists, err := s.svcIndexer.Get(key)
	switch {
	case !exists:
		// service absence in store means watcher caught the deletion, ensure LB info is cleaned
		err = s.processServiceDeletion(ctx, nil, key)
	case err != nil:
		runtime.HandleError(fmt.Errorf("unable to retrieve service %v from store: %v", key, err))
	default:
		svc := service.(*corev1.Service)
		err = s.processServiceCreateOrUpdate(ctx, svc, key)
	}

	return err
}

func (s *Controller) processServiceDeletion(ctx context.Context, service *corev1.Service, key string) error {
	return s.balancer.EnsureLoadBalancerDeleted()
}

func (s *Controller) processServiceCreateOrUpdate(ctx context.Context, service *corev1.Service, key string) error {
	s.balancer.Allocate()

	s.updateSvcStatus(svc)
}

// @see k8s.io/cloud-provider@v0.23.4/controllers/service/controller.go
func wantsLoadBalancer(service *corev1.Service) bool {
	// if LoadBalancerClass is set, the user does not want the default cloud-provider Load Balancer
	return service.Spec.Type == corev1.ServiceTypeLoadBalancer && service.Spec.LoadBalancerClass == nil
}
