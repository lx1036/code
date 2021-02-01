package controller

import (
	"context"
	"fmt"
	"k8s-lx1036/k8s/storage/csi/external-provisioner/external-provisioner-lib/pkg/controller"
	v1 "k8s.io/api/core/v1"
	"time"

	"k8s-lx1036/k8s/storage/csi/csi-lib-utils/rpc"

	"github.com/container-storage-interface/spec/lib/go/csi"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

//
// This package introduce a way to handle finalizers, related to in-progress PVC cloning. This is a two step approach:
//
// 1) PVC referenced as a data source is now updated with a finalizer `provisioner.storage.kubernetes.io/cloning-protection` during a ProvisionExt method call.
// The detection of cloning in-progress is based on the assumption that a PVC with `spec.DataSource` pointing on a another PVC will go into `Pending` state.
// The downside of this, is that fact that any other reason causing PVC to stay in the `Pending` state also blocks resource from deletion it from deletion
//
// 2) When cloning is finished for each PVC referencing the one as a data source,
// this PVC will go from `Pending` to `Bound` state. That allows remove the finalizer.
//

// CloningProtectionController is storing all related interfaces
// to handle cloning protection finalizer removal after CSI cloning is finished
type CloningProtectionController struct {
	client        kubernetes.Interface
	claimLister   corelisters.PersistentVolumeClaimLister
	claimInformer cache.SharedInformer
	claimQueue    workqueue.RateLimitingInterface
}

// NewCloningProtectionController creates new controller for additional CSI claim protection capabilities
func NewCloningProtectionController(
	client kubernetes.Interface,
	claimLister corelisters.PersistentVolumeClaimLister,
	claimInformer cache.SharedInformer,
	claimQueue workqueue.RateLimitingInterface,
	controllerCapabilities rpc.ControllerCapabilitySet,
) *CloningProtectionController {
	if !controllerCapabilities[csi.ControllerServiceCapability_RPC_CLONE_VOLUME] {
		return nil
	}
	controller := &CloningProtectionController{
		client:        client,
		claimLister:   claimLister,
		claimInformer: claimInformer,
		claimQueue:    claimQueue,
	}
	return controller
}

// Run is a main CloningProtectionController handler
func (p *CloningProtectionController) Run(ctx context.Context, threadiness int) {
	klog.Info("Starting CloningProtection controller")
	defer utilruntime.HandleCrash()
	defer p.claimQueue.ShutDown()

	claimHandler := cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) { p.enqueueClaimUpdate(ctx, obj) },
		UpdateFunc: func(_ interface{}, newObj interface{}) { p.enqueueClaimUpdate(ctx, newObj) },
	}
	p.claimInformer.AddEventHandlerWithResyncPeriod(claimHandler, controller.DefaultResyncPeriod)

	for i := 0; i < threadiness; i++ {
		go wait.Until(func() {
			p.runClaimWorker(ctx)
		}, time.Second, ctx.Done())
	}

	go p.claimInformer.Run(ctx.Done())

	klog.Infof("Started CloningProtection controller")
	<-ctx.Done()
	klog.Info("Shutting down CloningProtection controller")
}

// enqueueClaimUpdate takes a PVC obj and stores it into the claim work queue.
func (p *CloningProtectionController) enqueueClaimUpdate(ctx context.Context, obj interface{}) {
	pvc, ok := obj.(*v1.PersistentVolumeClaim)
	if !ok {
		utilruntime.HandleError(fmt.Errorf("expected claim but got %+v", pvc))
		return
	}

	// Timestamp didn't appear
	if pvc.DeletionTimestamp == nil {
		return
	}

	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(err)
		return
	}

	p.claimQueue.Add(key)
}
