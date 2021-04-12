package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

type resizeController struct {
	name string

	kubeClient kubernetes.Interface
	claimQueue workqueue.RateLimitingInterface

	pvSynced  cache.InformerSynced
	pvcSynced cache.InformerSynced

	eventRecorder record.EventRecorder

	// a cache to store PersistentVolume objects in local
	volumes cache.Store
	// a cache to store PersistentVolumeClaim objects in local
	claims cache.Store
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

// INFO: patch更新pvc.status.conditions, 可以多多参考
func (controller *resizeController) markPVCResizeInProgress(pvc *v1.PersistentVolumeClaim) (*v1.PersistentVolumeClaim, error) {
	// Mark PVC as Resize Started
	progressCondition := v1.PersistentVolumeClaimCondition{
		Type:               v1.PersistentVolumeClaimResizing,
		Status:             v1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
	}
	newPVC := pvc.DeepCopy()
	newPVC.Status.Conditions = MergeResizeConditionsOfPVC(newPVC.Status.Conditions,
		[]v1.PersistentVolumeClaimCondition{progressCondition})

	updatedPVC, err := controller.patchClaim(pvc, newPVC)
	if err != nil {
		return nil, err
	}
	return updatedPVC, nil
}

func GetPatchData(oldObj, newObj interface{}) ([]byte, error) {
	oldData, err := json.Marshal(oldObj)
	if err != nil {
		return nil, fmt.Errorf("marshal old object failed: %v", err)
	}
	newData, err := json.Marshal(newObj)
	if err != nil {
		return nil, fmt.Errorf("marshal new object failed: %v", err)
	}
	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, oldObj)
	if err != nil {
		return nil, fmt.Errorf("CreateTwoWayMergePatch failed: %v", err)
	}
	return patchBytes, nil
}

// 这里注意下addResourceVersion()会设置newPVC.ResourceVersion等于oldPVC.ResourceVersion
func GetPVCPatchData(oldPVC, newPVC *v1.PersistentVolumeClaim) ([]byte, error) {
	patchBytes, err := GetPatchData(oldPVC, newPVC)
	if err != nil {
		return patchBytes, err
	}

	patchBytes, err = addResourceVersion(patchBytes, oldPVC.ResourceVersion)
	if err != nil {
		return nil, fmt.Errorf("apply ResourceVersion to patch data failed: %v", err)
	}
	return patchBytes, nil
}

// 给patch bytes添加resource version
func addResourceVersion(patchBytes []byte, resourceVersion string) ([]byte, error) {
	var patchMap map[string]interface{}
	err := json.Unmarshal(patchBytes, &patchMap)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling patch with %v", err)
	}
	u := unstructured.Unstructured{Object: patchMap}
	accessor, err := meta.Accessor(&u)
	if err != nil {
		return nil, fmt.Errorf("error creating accessor with  %v", err)
	}
	// 设置resourceVersion
	accessor.SetResourceVersion(resourceVersion)
	versionBytes, err := json.Marshal(patchMap)
	if err != nil {
		return nil, fmt.Errorf("error marshalling json patch with %v", err)
	}
	return versionBytes, nil
}

func (controller *resizeController) patchClaim(oldPVC, newPVC *v1.PersistentVolumeClaim) (*v1.PersistentVolumeClaim, error) {
	patchBytes, err := GetPVCPatchData(oldPVC, newPVC)
	if err != nil {
		return nil, fmt.Errorf("can't patch status of PVC %s/%s as generate path data failed: %v", oldPVC.Namespace, oldPVC.Name, err)
	}
	updatedClaim, updateErr := controller.kubeClient.CoreV1().PersistentVolumeClaims(oldPVC.Namespace).
		Patch(context.TODO(), oldPVC.Name, types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{}, "status")
	if updateErr != nil {
		return nil, fmt.Errorf("can't patch status of  PVC %s/%s with %v", oldPVC.Namespace, oldPVC.Name, updateErr)
	}

	// 更新本地缓存
	err = controller.claims.Update(updatedClaim)
	if err != nil {
		return nil, fmt.Errorf("error updating PVC %s/%s in local cache: %v", oldPVC.Namespace, oldPVC.Name, err)
	}

	return updatedClaim, nil
}

// resizePVC will:
// 1. Mark pvc as resizing.
// 2. Resize the volume and the pv object.
// 3. Mark pvc as resizing finished(no error, no need to resize fs), need resizing fs or resize failed.
func (controller *resizeController) resizePVC(pvc *v1.PersistentVolumeClaim, pv *v1.PersistentVolume) error {
	if updatedPVC, err := controller.markPVCResizeInProgress(pvc); err != nil {
		return fmt.Errorf("marking pvc %s/%s as resizing failed: %v", pvc.Namespace, pvc.Name, err)
	} else if updatedPVC != nil {
		pvc = updatedPVC
	}

	// INFO:

	// Record an event to indicate that external resizer is resizing this volume.
	controller.eventRecorder.Event(pvc, v1.EventTypeNormal, string(v1.PersistentVolumeClaimResizing),
		fmt.Sprintf("External resizer is resizing volume %s", pv.Name))

	err := func() error {
		newSize, fsResizeRequired, err := controller.resizeVolume(pvc, pv)
		if err != nil {
			return err
		}

		if fsResizeRequired {
			// Resize volume succeeded and need to resize file system by kubelet, mark it as file system resizing required.
			return controller.markPVCAsFSResizeRequired(pvc)
		}
		// Resize volume succeeded and no need to resize file system by kubelet, mark it as resizing finished.
		return controller.markPVCResizeFinished(pvc, newSize)
	}()

	if err != nil {
		// Record an event to indicate that resize operation is failed.
		controller.eventRecorder.Eventf(pvc, v1.EventTypeWarning, VolumeResizeFailed, err.Error())
	}
	return err
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

func NewResizeController(
	name string,
	resizer resizer.Resizer,
	kubeClient kubernetes.Interface,
	resyncPeriod time.Duration,
	informerFactory informers.SharedInformerFactory,
	pvcRateLimiter workqueue.RateLimiter,
	handleVolumeInUseError bool) *resizeController {

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&corev1.EventSinkImpl{Interface: kubeClient.CoreV1().Events(v1.NamespaceAll)})
	eventRecorder := eventBroadcaster.NewRecorder(scheme.Scheme,
		v1.EventSource{Component: fmt.Sprintf("external-resizer %s", name)})

	pvInformer := informerFactory.Core().V1().PersistentVolumes()
	pvcInformer := informerFactory.Core().V1().PersistentVolumeClaims()

	controller := &resizeController{
		name:          name,
		kubeClient:    kubeClient,
		eventRecorder: eventRecorder,
		claimQueue:    workqueue.NewNamedRateLimitingQueue(workqueue.NewItemExponentialFailureRateLimiter(time.Second*5, time.Minute*5), "claims"),
		pvSynced:      pvInformer.Informer().HasSynced,
		pvcSynced:     pvcInformer.Informer().HasSynced,
		claims:        pvcInformer.Informer().GetStore(),
		volumes:       pvInformer.Informer().GetStore(),
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
