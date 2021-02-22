package capacity

import (
	"context"
	"sync"
	"time"

	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	storageinformersv1 "k8s.io/client-go/informers/storage/v1"
	storageinformersv1alpha1 "k8s.io/client-go/informers/storage/v1alpha1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

// Controller creates and updates CSIStorageCapacity objects.  It
// deletes those which are no longer needed because their storage
// class or topology segment are gone. The controller only manages
// those CSIStorageCapacity objects that are owned by a certain
// entity.
//
// The controller maintains a set of topology segments (= NodeSelector
// pointers). Work items are a combination of such a pointer and a
// pointer to a storage class. These keys are mapped to the
// corresponding CSIStorageCapacity object, if one exists.
//
// When processing a work item, the controller first checks whether
// the topology segment and storage class still exist. If not,
// the CSIStorageCapacity object gets deleted. Otherwise, it gets updated
// or created.
//
// New work items are queued for processing when the reconiliation loop
// finds differences, periodically (to refresh existing items) and when
// capacity is expected to have changed.
//
// The work queue is also used to delete duplicate CSIStorageCapacity objects,
// i.e. those that for some reason have the same topology segment
// and storage class name as some other object. That should never happen,
// but the controller is prepared to clean that up, just in case.
type Controller struct {
	csiController    CSICapacityClient
	driverName       string
	client           kubernetes.Interface
	queue            workqueue.RateLimitingInterface
	owner            metav1.OwnerReference
	ownerNamespace   string
	topologyInformer topology.Informer
	scInformer       storageinformersv1.StorageClassInformer
	cInformer        storageinformersv1alpha1.CSIStorageCapacityInformer
	pollPeriod       time.Duration
	immediateBinding bool

	// capacities contains one entry for each object that is supposed
	// to exist.
	capacities     map[workItem]*storagev1alpha1.CSIStorageCapacity
	capacitiesLock sync.Mutex
}

// NewController creates a new controller for CSIStorageCapacity objects.
func NewCentralCapacityController(
	csiController CSICapacityClient,
	driverName string,
	client kubernetes.Interface,
	queue workqueue.RateLimitingInterface,
	owner metav1.OwnerReference,
	ownerNamespace string,
	topologyInformer topology.Informer,
	scInformer storageinformersv1.StorageClassInformer,
	cInformer storageinformersv1alpha1.CSIStorageCapacityInformer,
	pollPeriod time.Duration,
	immediateBinding bool,
) *Controller {
	c := &Controller{
		csiController:    csiController,
		driverName:       driverName,
		client:           client,
		queue:            queue,
		owner:            owner,
		ownerNamespace:   ownerNamespace,
		topologyInformer: topologyInformer,
		scInformer:       scInformer,
		cInformer:        cInformer,
		pollPeriod:       pollPeriod,
		immediateBinding: immediateBinding,
		capacities:       map[workItem]*storagev1alpha1.CSIStorageCapacity{},
	}

	// Now register for changes. Depending on the implementation of the informers,
	// this may already invoke callbacks.
	handler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			sc, ok := obj.(*storagev1.StorageClass)
			if !ok {
				klog.Errorf("added object: expected StorageClass, got %T -> ignoring it", obj)
				return
			}
			c.onSCAddOrUpdate(sc)
		},
		UpdateFunc: func(_ interface{}, newObj interface{}) {
			sc, ok := newObj.(*storagev1.StorageClass)
			if !ok {
				klog.Errorf("updated object: expected StorageClass, got %T -> ignoring it", newObj)
				return
			}
			c.onSCAddOrUpdate(sc)
		},
		DeleteFunc: func(obj interface{}) {
			// Beware of "xxx deleted" events
			if unknown, ok := obj.(cache.DeletedFinalStateUnknown); ok && unknown.Obj != nil {
				obj = unknown.Obj
			}
			sc, ok := obj.(*storagev1.StorageClass)
			if !ok {
				klog.Errorf("deleted object: expected StorageClass, got %T -> ignoring it", obj)
				return
			}
			c.onSCDelete(sc)
		},
	}
	c.scInformer.Informer().AddEventHandler(handler)
	c.topologyInformer.AddCallback(c.onTopologyChanges)

	// We don't want the callbacks yet, but need to ensure that
	// the informer controller is instantiated before the caller
	// starts the factory.
	cInformer.Informer()

	return c
}

func (c *Controller) prepare(ctx context.Context) {
	// Wait for topology and storage class informer sync. Once we have that,
	// we know which CSIStorageCapacity objects we need.
	if !cache.WaitForCacheSync(ctx.Done(), c.topologyInformer.HasSynced, c.scInformer.Informer().HasSynced, c.cInformer.Informer().HasSynced) {
		return
	}

	// The caches are fully populated now, but the event handlers
	// may or may not have been invoked yet. To be sure that we
	// have all data, we need to list all resources. Here we list
	// topology segments, onTopologyChanges lists the classes.
	c.onTopologyChanges(c.topologyInformer.List(), nil)

	if klog.V(3).Enabled() {
		scs, err := c.scInformer.Lister().List(labels.Everything())
		if err != nil {
			// Shouldn't happen.
			utilruntime.HandleError(err)
		}
		klog.V(3).Infof("Initial number of topology segments %d, storage classes %d, potential CSIStorageCapacity objects %d",
			len(c.topologyInformer.List()),
			len(scs),
			len(c.capacities))
	}

	// Now that we know what we need, we can check what we have.
	// We do that both via callbacks *and* by iterating over all
	// objects: callbacks handle future updates and iterating
	// avoids the assumumption that the callback will be invoked
	// for all objects immediately when adding it.
	klog.V(3).Info("Checking for existing CSIStorageCapacity objects")
	handler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			csc, ok := obj.(*storagev1alpha1.CSIStorageCapacity)
			if !ok {
				klog.Errorf("added object: expected CSIStorageCapacity, got %T -> ignoring it", obj)
				return
			}
			c.onCAddOrUpdate(ctx, csc)
		},
		UpdateFunc: func(_ interface{}, newObj interface{}) {
			csc, ok := newObj.(*storagev1alpha1.CSIStorageCapacity)
			if !ok {
				klog.Errorf("updated object: expected CSIStorageCapacity, got %T -> ignoring it", newObj)
				return
			}
			c.onCAddOrUpdate(ctx, csc)
		},
		DeleteFunc: func(obj interface{}) {
			// Beware of "xxx deleted" events
			if unknown, ok := obj.(cache.DeletedFinalStateUnknown); ok && unknown.Obj != nil {
				obj = unknown.Obj
			}
			csc, ok := obj.(*storagev1alpha1.CSIStorageCapacity)
			if !ok {
				klog.Errorf("deleted object: expected CSIStorageCapacity, got %T -> ignoring it", obj)
				return
			}
			c.onCDelete(ctx, csc)
		},
	}
	c.cInformer.Informer().AddEventHandler(handler)
	capacities, err := c.cInformer.Lister().List(labels.Everything())
	if err != nil {
		// This shouldn't happen.
		utilruntime.HandleError(err)
		return
	}
	for _, capacity := range capacities {
		c.onCAddOrUpdate(ctx, capacity)
	}

	// Now that we have seen all existing objects, we are done
	// with the preparation and can let our caller start
	// processing work items.
}

// Run is a main Controller handler
func (c *Controller) Run(ctx context.Context, threadiness int) {
	klog.Info("Starting Capacity Controller")
	defer c.queue.ShutDown()
	go c.scInformer.Informer().Run(ctx.Done())
	go c.topologyInformer.Run(ctx)

	c.prepare(ctx)
	for i := 0; i < threadiness; i++ {
		go wait.UntilWithContext(ctx, func(ctx context.Context) {
			c.runWorker(ctx)
		}, time.Second)
	}

	go wait.UntilWithContext(ctx, func(ctx context.Context) { c.pollCapacities() }, c.pollPeriod)

	klog.Info("Started Capacity Controller")
	<-ctx.Done()
	klog.Info("Shutting down Capacity Controller")
}
