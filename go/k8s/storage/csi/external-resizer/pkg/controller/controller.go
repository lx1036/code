package controller

import (
	"context"
	"fmt"
	"k8s-lx1036/k8s/storage/csi/external-resizer/pkg/util"
	v1 "k8s.io/api/core/v1"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

type resizeController struct {
	name string

	claimQueue workqueue.RateLimitingInterface

	pvSynced  cache.InformerSynced
	pvcSynced cache.InformerSynced

	// a cache to store PersistentVolume objects in local
	volumes cache.Store
	// a cache to store PersistentVolumeClaim objects in local
	claims cache.Store
}

// Run starts the controller.
func (controller *resizeController) Run(workers int, ctx context.Context) {
	defer controller.claimQueue.ShutDown()

	klog.Infof("Starting external resizer %s", controller.name)
	defer klog.Infof("Shutting down external resizer %s", controller.name)

	stopCh := ctx.Done()
	informersSyncd := []cache.InformerSynced{controller.pvSynced, controller.pvcSynced}
	/*if controller.handleVolumeInUseError {
		informersSyncd = append(informersSyncd, controller.podListerSynced)
	}*/

	if !cache.WaitForCacheSync(stopCh, informersSyncd...) {
		klog.Errorf("Cannot sync pod, pv or pvc caches")
		return
	}

	for i := 0; i < workers; i++ {
		go wait.Until(controller.syncPVCs, 0, stopCh)
	}

	<-stopCh
}

func (controller *resizeController) syncPVCs() {
	key, quit := controller.claimQueue.Get()
	if quit {
		return
	}
	defer controller.claimQueue.Done(key)

	if err := controller.syncPVC(key.(string)); err != nil {
		// Put PVC back to the queue so that we can retry later.
		klog.Errorf("Error syncing PVC: %v", err)
		controller.claimQueue.AddRateLimited(key)
	} else {
		controller.claimQueue.Forget(key)
	}
}

func (controller *resizeController) addPVC(obj interface{}) {
	objKey, err := getObjectKey(obj)
	if err != nil {
		return
	}
	controller.claimQueue.Add(objKey)
}

// updatePVC这里考虑"newPVC.ResourceVersion == oldPVC.ResourceVersion"，还是有问题？？？
func (controller *resizeController) updatePVC(oldObj, newObj interface{}) {
	oldPVC, ok := oldObj.(*v1.PersistentVolumeClaim)
	if !ok || oldPVC == nil {
		return
	}

	newPVC, ok := newObj.(*v1.PersistentVolumeClaim)
	if !ok || newPVC == nil {
		return
	}

	newSize := newPVC.Spec.Resources.Requests[v1.ResourceStorage]
	oldSize := oldPVC.Spec.Resources.Requests[v1.ResourceStorage]

	newResizerName := newPVC.Annotations[util.VolumeResizerKey]
	oldResizerName := oldPVC.Annotations[util.VolumeResizerKey]

	// We perform additional checks to avoid double processing of PVCs, as we will also receive Update event when:
	// 1. Administrator or users may introduce other changes(such as add labels, modify annotations, etc.)
	//    unrelated to volume resize.
	// 2. Informer will resync and send Update event periodically without any changes.
	//
	// We add the PVC into work queue when the new size is larger then the old size
	// or when the resizer name changes. This is needed for CSI migration for the follow two cases:
	//
	// 1. First time a migrated PVC is expanded:
	// It does not yet have the annotation because annotation is only added by in-tree resizer when it receives a volume
	// expansion request. So first update event that will be received by external-resizer will be ignored because it won't
	// know how to support resizing of a "un-annotated" in-tree PVC. When in-tree resizer does add the annotation, a second
	// update even will be received and we add the pvc to workqueue. If annotation matches the registered driver name in
	// csi_resizer object, we proceeds with expansion internally or we discard the PVC.
	// 2. An already expanded in-tree PVC:
	// An in-tree PVC is resized with in-tree resizer. And later, CSI migration is turned on and resizer name is updated from
	// in-tree resizer name to CSI driver name.
	if newSize.Cmp(oldSize) > 0 || newResizerName != oldResizerName {
		controller.addPVC(newObj)
	} else {
		// PVC's size not changed, so this Update event maybe caused by:
		//
		// 1. Administrators or users introduce other changes(such as add labels, modify annotations, etc.)
		//    unrelated to volume resize.
		// 2. Informer resynced the PVC and send this Update event without any changes.
		//
		// If it is case 1, we can just discard this event. If case 2, we need to put it into the queue to
		// perform a resync operation.
		if newPVC.ResourceVersion == oldPVC.ResourceVersion {
			// This is case 2.
			controller.addPVC(newObj)
		}
	}
}

func (controller *resizeController) deletePVC(obj interface{}) {
	objKey, err := getObjectKey(obj)
	if err != nil {
		return
	}

	controller.claimQueue.Forget(objKey)
}

func getObjectKey(obj interface{}) (string, error) {
	if unknown, ok := obj.(cache.DeletedFinalStateUnknown); ok && unknown.Obj != nil {
		obj = unknown.Obj
	}
	objKey, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		klog.Errorf("Failed to get key from object: %v", err)
		return "", err
	}

	return objKey, nil
}

// syncPVC checks if a pvc requests resizing, and execute the resize operation if requested.
func (controller *resizeController) syncPVC(key string) error {
	klog.Infof("Started PVC processing %s", key)

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return fmt.Errorf("getting namespace and name from key %s failed: %v", key, err)
	}

	pvcObject, exists, err := controller.claims.GetByKey(key)
	if err != nil {
		return fmt.Errorf("getting PVC %s/%s failed: %v", namespace, name, err)
	}

	if !exists {
		klog.V(3).Infof("PVC %s/%s is deleted or does not exist", namespace, name)
		return nil
	}

	pvc, ok := pvcObject.(*v1.PersistentVolumeClaim)
	if !ok {
		return fmt.Errorf("expected PVC got: %v", pvcObject)
	}

	if !controller.pvcNeedResize(pvc) {
		klog.V(4).Infof("No need to resize PVC %s/%s", pvc.Namespace, pvc.Name)
		return nil
	}

	volumeObj, exists, err := controller.volumes.GetByKey(pvc.Spec.VolumeName)
	if err != nil {
		return fmt.Errorf("get PV %s of pvc %s/%s failed: %v", pvc.Spec.VolumeName, pvc.Namespace, pvc.Name, err)
	}

	if !exists {
		klog.Warningf("PV %q bound to PVC %s/%s not found", pvc.Spec.VolumeName, pvc.Namespace, pvc.Name)
		return nil
	}

	pv, ok := volumeObj.(*v1.PersistentVolume)
	if !ok {
		return fmt.Errorf("expected volume but got %+v", volumeObj)
	}

	if !controller.pvNeedResize(pvc, pv) {
		klog.V(4).Infof("No need to resize PV %q", pv.Name)
		return nil
	}

	return controller.resizePVC(pvc, pv)

}

func NewResizeController(
	name string,
	resizer resizer.Resizer,
	kubeClient kubernetes.Interface,
	resyncPeriod time.Duration,
	informerFactory informers.SharedInformerFactory,
	pvcRateLimiter workqueue.RateLimiter,
	handleVolumeInUseError bool) *resizeController {

	pvInformer := informerFactory.Core().V1().PersistentVolumes()
	pvcInformer := informerFactory.Core().V1().PersistentVolumeClaims()

	controller := &resizeController{
		name:       name,
		claimQueue: workqueue.NewNamedRateLimitingQueue(workqueue.NewItemExponentialFailureRateLimiter(time.Second*5, time.Minute*5), "claims"),
		pvSynced:   pvInformer.Informer().HasSynced,
		pvcSynced:  pvcInformer.Informer().HasSynced,
		claims:     pvcInformer.Informer().GetStore(),
		volumes:    pvInformer.Informer().GetStore(),
	}

	// Add a resync period as the PVC's request size can be resized again when we handling
	// a previous resizing request of the same PVC.
	pvcInformer.Informer().AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.addPVC,
		UpdateFunc: controller.updatePVC,
		DeleteFunc: controller.deletePVC,
	}, resyncPeriod)

	return controller
}
