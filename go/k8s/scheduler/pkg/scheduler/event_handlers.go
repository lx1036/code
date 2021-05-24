package scheduler

import (
	"fmt"

	internalqueue "k8s-lx1036/k8s/scheduler/pkg/scheduler/internal/queue"

	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

func (scheduler *Scheduler) onCSINodeAdd(obj interface{}) {
	scheduler.PriorityQueue.MoveAllToActiveOrBackoffQueue(internalqueue.CSINodeAdd)
}
func (scheduler *Scheduler) onCSINodeUpdate(oldObj, newObj interface{}) {
	scheduler.PriorityQueue.MoveAllToActiveOrBackoffQueue(internalqueue.CSINodeUpdate)
}
func (scheduler *Scheduler) onPvAdd(obj interface{}) {
	// Pods created when there are no PVs available will be stuck in
	// unschedulable internalqueue. But unbound PVs created for static provisioning and
	// delay binding storage class are skipped in PV controller dynamic
	// provisioning and binding process, will not trigger events to schedule pod
	// again. So we need to move pods to active queue on PV add for this
	// scenario.
	scheduler.PriorityQueue.MoveAllToActiveOrBackoffQueue(internalqueue.PvAdd)
}
func (scheduler *Scheduler) onPvUpdate(old, new interface{}) {
	// Scheduler.bindVolumesWorker may fail to update assumed pod volume
	// bindings due to conflicts if PVs are updated by PV controller or other
	// parties, then scheduler will add pod back to unschedulable internalqueue. We
	// need to move pods to active queue on PV update for this scenario.
	scheduler.PriorityQueue.MoveAllToActiveOrBackoffQueue(internalqueue.PvUpdate)
}
func (scheduler *Scheduler) onPvcAdd(obj interface{}) {
	scheduler.PriorityQueue.MoveAllToActiveOrBackoffQueue(internalqueue.PvcAdd)
}
func (scheduler *Scheduler) onPvcUpdate(old, new interface{}) {
	scheduler.PriorityQueue.MoveAllToActiveOrBackoffQueue(internalqueue.PvcUpdate)
}
func (scheduler *Scheduler) onStorageClassAdd(obj interface{}) {
	sc, ok := obj.(*storagev1.StorageClass)
	if !ok {
		klog.Errorf("cannot convert to *storagev1.StorageClass: %v", obj)
		return
	}

	// CheckVolumeBindingPred fails if pod has unbound immediate PVCs. If these
	// PVCs have specified StorageClass name, creating StorageClass objects
	// with late binding will cause predicates to pass, so we need to move pods
	// to active internalqueue.
	// We don't need to invalidate cached results because results will not be
	// cached for pod that has unbound immediate PVCs.
	if sc.VolumeBindingMode != nil && *sc.VolumeBindingMode == storagev1.VolumeBindingWaitForFirstConsumer {
		scheduler.PriorityQueue.MoveAllToActiveOrBackoffQueue(internalqueue.StorageClassAdd)
	}
}
func (scheduler *Scheduler) onServiceAdd(obj interface{}) {
	scheduler.PriorityQueue.MoveAllToActiveOrBackoffQueue(internalqueue.ServiceAdd)
}
func (scheduler *Scheduler) onServiceUpdate(oldObj interface{}, newObj interface{}) {
	scheduler.PriorityQueue.MoveAllToActiveOrBackoffQueue(internalqueue.ServiceUpdate)
}
func (scheduler *Scheduler) onServiceDelete(obj interface{}) {
	scheduler.PriorityQueue.MoveAllToActiveOrBackoffQueue(internalqueue.ServiceDelete)
}

func addAllEventHandlers(
	scheduler *Scheduler,
	informerFactory informers.SharedInformerFactory,
	podInformer coreinformers.PodInformer,
) {

	// scheduled pod cache
	podInformer.Informer().AddEventHandler(
		cache.FilteringResourceEventHandler{
			FilterFunc: func(obj interface{}) bool { // 只watch已经被scheduled的pod
				switch t := obj.(type) {
				case *v1.Pod:
					return len(t.Spec.NodeName) != 0
				case cache.DeletedFinalStateUnknown:
					if pod, ok := t.Obj.(*v1.Pod); ok {
						return len(pod.Spec.NodeName) != 0
					}
					utilruntime.HandleError(fmt.Errorf("unable to convert object %T to *v1.Pod in %T", obj, scheduler))
					return false
				default:
					utilruntime.HandleError(fmt.Errorf("unable to handle object in %T: %T", scheduler, obj))
					return false
				}
			},
			Handler: cache.ResourceEventHandlerFuncs{
				AddFunc:    scheduler.addPodToCache,
				UpdateFunc: scheduler.updatePodInCache,
				DeleteFunc: scheduler.deletePodFromCache,
			},
		},
	)
	// unscheduled pod queue
	podInformer.Informer().AddEventHandler(
		cache.FilteringResourceEventHandler{
			FilterFunc: func(obj interface{}) bool {
				switch t := obj.(type) {
				case *v1.Pod:
					return !(len(t.Spec.NodeName) != 0) && scheduler.Profiles.HandlesSchedulerName(t.Spec.SchedulerName)
				case cache.DeletedFinalStateUnknown:
					if pod, ok := t.Obj.(*v1.Pod); ok {
						return !(len(pod.Spec.NodeName) != 0) && responsibleForPod(pod, scheduler.Profiles)
					}
					utilruntime.HandleError(fmt.Errorf("unable to convert object %T to *v1.Pod in %T", obj, scheduler))
					return false
				default:
					utilruntime.HandleError(fmt.Errorf("unable to handle object in %T: %T", scheduler, obj))
					return false
				}
			},
			Handler: cache.ResourceEventHandlerFuncs{
				AddFunc:    scheduler.addPodToSchedulingQueue,
				UpdateFunc: scheduler.updatePodInSchedulingQueue,
				DeleteFunc: scheduler.deletePodFromSchedulingQueue,
			},
		},
	)
	informerFactory.Core().V1().Nodes().Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    scheduler.addNodeToCache,
			UpdateFunc: scheduler.updateNodeInCache,
			DeleteFunc: scheduler.deleteNodeFromCache,
		},
	)
	informerFactory.Storage().V1().CSINodes().Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    scheduler.onCSINodeAdd,
			UpdateFunc: scheduler.onCSINodeUpdate,
		},
	)
	// On add and delete of PVs, it will affect equivalence cache items
	// related to persistent volume
	informerFactory.Core().V1().PersistentVolumes().Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			// MaxPDVolumeCountPredicate: since it relies on the counts of PV.
			AddFunc:    scheduler.onPvAdd,
			UpdateFunc: scheduler.onPvUpdate,
		},
	)
	// This is for MaxPDVolumeCountPredicate: add/delete PVC will affect counts of PV when it is bound.
	informerFactory.Core().V1().PersistentVolumeClaims().Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    scheduler.onPvcAdd,
			UpdateFunc: scheduler.onPvcUpdate,
		},
	)
	// This is for ServiceAffinity: affected by the selector of the service is updated.
	// Also, if new service is added, equivalence cache will also become invalid since
	// existing pods may be "captured" by this service and change this predicate result.
	informerFactory.Core().V1().Services().Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    scheduler.onServiceAdd,
			UpdateFunc: scheduler.onServiceUpdate,
			DeleteFunc: scheduler.onServiceDelete,
		},
	)
	informerFactory.Storage().V1().StorageClasses().Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: scheduler.onStorageClassAdd,
		},
	)

}
