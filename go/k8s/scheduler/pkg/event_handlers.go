package pkg

import (
	"fmt"

	"k8s-lx1036/k8s/scheduler/pkg/framework"
	internalqueue "k8s-lx1036/k8s/scheduler/pkg/internal/queue"
	"k8s-lx1036/k8s/scheduler/pkg/profile"

	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

// INFO: MoveAllToActiveOrBackoffQueue() 就是触发把 unschedulePods 再次调度

func (scheduler *Scheduler) onCSINodeAdd(obj interface{}) {
	scheduler.PriorityQueue.MoveAllToActiveOrBackoffQueue(internalqueue.CSINodeAdd, nil)
}
func (scheduler *Scheduler) onCSINodeUpdate(oldObj, newObj interface{}) {
	scheduler.PriorityQueue.MoveAllToActiveOrBackoffQueue(internalqueue.CSINodeUpdate, nil)
}
func (scheduler *Scheduler) onPvAdd(obj interface{}) {
	// Pods created when there are no PVs available will be stuck in
	// unschedulable internalqueue. But unbound PVs created for static provisioning and
	// delay binding storage class are skipped in PV controller dynamic
	// provisioning and binding process, will not trigger events to schedule pod
	// again. So we need to move pods to active queue on PV add for this
	// scenario.
	scheduler.PriorityQueue.MoveAllToActiveOrBackoffQueue(internalqueue.PvAdd, nil)
}
func (scheduler *Scheduler) onPvUpdate(old, new interface{}) {
	// Scheduler.bindVolumesWorker may fail to update assumed pod volume
	// bindings due to conflicts if PVs are updated by PV controller or other
	// parties, then scheduler will add pod back to unschedulable internalqueue. We
	// need to move pods to active queue on PV update for this scenario.
	scheduler.PriorityQueue.MoveAllToActiveOrBackoffQueue(internalqueue.PvUpdate, nil)
}
func (scheduler *Scheduler) onPvcAdd(obj interface{}) {
	scheduler.PriorityQueue.MoveAllToActiveOrBackoffQueue(internalqueue.PvcAdd, nil)
}
func (scheduler *Scheduler) onPvcUpdate(old, new interface{}) {
	scheduler.PriorityQueue.MoveAllToActiveOrBackoffQueue(internalqueue.PvcUpdate, nil)
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
		scheduler.PriorityQueue.MoveAllToActiveOrBackoffQueue(internalqueue.StorageClassAdd, nil)
	}
}
func (scheduler *Scheduler) onServiceAdd(obj interface{}) {
	scheduler.PriorityQueue.MoveAllToActiveOrBackoffQueue(internalqueue.ServiceAdd, nil)
}
func (scheduler *Scheduler) onServiceUpdate(oldObj interface{}, newObj interface{}) {
	scheduler.PriorityQueue.MoveAllToActiveOrBackoffQueue(internalqueue.ServiceUpdate, nil)
}
func (scheduler *Scheduler) onServiceDelete(obj interface{}) {
	scheduler.PriorityQueue.MoveAllToActiveOrBackoffQueue(internalqueue.ServiceDelete, nil)
}

func (scheduler *Scheduler) addPodToCache(obj interface{}) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		klog.Errorf("cannot convert to *corev1.Pod: %v", obj)
		return
	}
	klog.Infof("add event for scheduled pod %s/%s ", pod.Namespace, pod.Name)

	// 存入scheduler的cache
	if err := scheduler.SchedulerCache.AddPod(pod); err != nil {
		klog.Errorf("scheduler cache AddPod failed: %v", err)
	}

	// 存入PriorityQueue
	scheduler.PriorityQueue.AssignedPodAdded(pod)
}
func (scheduler *Scheduler) updatePodInCache(oldObj, newObj interface{}) {
	oldPod, ok := oldObj.(*corev1.Pod)
	if !ok {
		klog.ErrorS(nil, "Cannot convert oldObj to *corev1.Pod", "oldObj", oldObj)
		return
	}
	newPod, ok := newObj.(*corev1.Pod)
	if !ok {
		klog.ErrorS(nil, "Cannot convert newObj to *corev1.Pod", "newObj", newObj)
		return
	}
	klog.V(4).InfoS("Update event for scheduled pod", "pod", klog.KObj(oldPod))

	if err := scheduler.SchedulerCache.UpdatePod(oldPod, newPod); err != nil {
		klog.ErrorS(err, "Scheduler cache UpdatePod failed", "pod", klog.KObj(oldPod))
	}

	scheduler.SchedulingQueue.AssignedPodUpdated(newPod)
}
func (scheduler *Scheduler) deletePodFromCache(obj interface{}) {
	var pod *corev1.Pod
	switch t := obj.(type) {
	case *corev1.Pod:
		pod = t
	case cache.DeletedFinalStateUnknown:
		var ok bool
		pod, ok = t.Obj.(*corev1.Pod)
		if !ok {
			klog.Errorf("cannot convert to *corev1.Pod: %v", t.Obj)
			return
		}
	default:
		klog.Errorf("cannot convert to *corev1.Pod: %v", t)
		return
	}
	klog.Infof("delete event for scheduled pod %s/%s ", pod.Namespace, pod.Name)
	// NOTE: Updates must be written to scheduler cache before invalidating
	// equivalence cache, because we could snapshot equivalence cache after the
	// invalidation and then snapshot the cache itself. If the cache is
	// snapshotted before updates are written, we would update equivalence
	// cache with stale information which is based on snapshot of old cache.
	if err := scheduler.SchedulerCache.RemovePod(pod); err != nil {
		klog.Errorf("scheduler cache RemovePod failed: %v", err)
	}

	scheduler.PriorityQueue.MoveAllToActiveOrBackoffQueue(internalqueue.AssignedPodDelete, nil)
}
func (scheduler *Scheduler) addPodToSchedulingQueue(obj interface{}) {
	pod := obj.(*corev1.Pod)
	klog.V(3).Infof("add event for unscheduled pod %s/%s", pod.Namespace, pod.Name)
	if err := scheduler.PriorityQueue.Add(pod); err != nil {
		utilruntime.HandleError(fmt.Errorf("unable to queue %T: %v", obj, err))
	}
}
func (scheduler *Scheduler) updatePodInSchedulingQueue(oldObj, newObj interface{}) {
	oldPod, newPod := oldObj.(*corev1.Pod), newObj.(*corev1.Pod)
	// Bypass update event that carries identical objects; otherwise, a duplicated
	// Pod may go through scheduling and cause unexpected behavior (see #96071).
	if oldPod.ResourceVersion == newPod.ResourceVersion {
		return
	}
	isAssumed, err := scheduler.SchedulerCache.IsAssumedPod(newPod)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("failed to check whether pod %s/%s is assumed: %v", newPod.Namespace, newPod.Name, err))
	}
	if isAssumed {
		return
	}

	if err := scheduler.PriorityQueue.Update(oldPod, newPod); err != nil {
		utilruntime.HandleError(fmt.Errorf("unable to update %T: %v", newObj, err))
	}
}
func (scheduler *Scheduler) deletePodFromSchedulingQueue(obj interface{}) {
	var pod *corev1.Pod
	switch t := obj.(type) {
	case *corev1.Pod:
		pod = obj.(*corev1.Pod)
	case cache.DeletedFinalStateUnknown:
		var ok bool
		pod, ok = t.Obj.(*corev1.Pod)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("unable to convert object %T to *corev1.Pod in %T", obj, scheduler))
			return
		}
	default:
		utilruntime.HandleError(fmt.Errorf("unable to handle object in %T: %T", scheduler, obj))
		return
	}
	klog.V(3).Infof("delete event for unscheduled pod %s/%s", pod.Namespace, pod.Name)
	if err := scheduler.PriorityQueue.Delete(pod); err != nil {
		utilruntime.HandleError(fmt.Errorf("unable to dequeue %T: %v", obj, err))
	}
	prof, err := scheduler.profileForPod(pod)
	if err != nil {
		// This shouldn't happen, because we only accept for scheduling the pods
		// which specify a scheduler name that matches one of the profiles.
		klog.Error(err)
		return
	}
	prof.Framework.RejectWaitingPod(pod.UID)
}
func (scheduler *Scheduler) addNodeToCache(obj interface{}) {
	node, ok := obj.(*corev1.Node)
	if !ok {
		klog.Errorf("cannot convert to *corev1.Node: %v", obj)
		return
	}

	nodeInfo := scheduler.SchedulerCache.AddNode(node)
	klog.V(3).Infof("add event for node %q", node.Name)
	scheduler.PriorityQueue.MoveAllToActiveOrBackoffQueue(internalqueue.NodeAdd, preCheckForNode(nodeInfo))
}
func (scheduler *Scheduler) updateNodeInCache(oldObj, newObj interface{}) {
	oldNode, ok := oldObj.(*corev1.Node)
	if !ok {
		klog.Errorf("cannot convert oldObj to *corev1.Node: %v", oldObj)
		return
	}
	newNode, ok := newObj.(*corev1.Node)
	if !ok {
		klog.Errorf("cannot convert newObj to *corev1.Node: %v", newObj)
		return
	}

	nodeInfo := scheduler.SchedulerCache.UpdateNode(oldNode, newNode)
	// Only requeue unschedulable pods if the node became more schedulable.
	if event := nodeSchedulingPropertiesChange(newNode, oldNode); event != nil {
		scheduler.PriorityQueue.MoveAllToActiveOrBackoffQueue(*event, preCheckForNode(nodeInfo))
	}
}
func (scheduler *Scheduler) deleteNodeFromCache(obj interface{}) {
	var node *corev1.Node
	switch t := obj.(type) {
	case *corev1.Node:
		node = t
	case cache.DeletedFinalStateUnknown:
		var ok bool
		node, ok = t.Obj.(*corev1.Node)
		if !ok {
			klog.Errorf("cannot convert to *corev1.Node: %v", t.Obj)
			return
		}
	default:
		klog.Errorf("cannot convert to *corev1.Node: %v", t)
		return
	}
	klog.V(3).Infof("delete event for node %q", node.Name)
	// NOTE: Updates must be written to scheduler cache before invalidating
	// equivalence cache, because we could snapshot equivalence cache after the
	// invalidation and then snapshot the cache itself. If the cache is
	// snapshotted before updates are written, we would update equivalence
	// cache with stale information which is based on snapshot of old cache.
	if err := scheduler.SchedulerCache.RemoveNode(node); err != nil {
		klog.Errorf("scheduler cache RemoveNode failed: %v", err)
	}
}

func assignedPod(pod *corev1.Pod) bool {
	return len(pod.Spec.NodeName) != 0
}

// responsibleForPod returns true if the pod has asked to be scheduled by the given scheduler.
func responsibleForPod(pod *corev1.Pod, profiles profile.Map) bool {
	return profiles.HandlesSchedulerName(pod.Spec.SchedulerName)
}

func addAllEventHandlers(
	scheduler *Scheduler,
	informerFactory informers.SharedInformerFactory,
	dynInformerFactory dynamicinformer.DynamicSharedInformerFactory,
	gvkMap map[framework.GVK]framework.ActionType,
) {
	// pod 已经调度了，放到 pod cache 和调度队列中
	informerFactory.Core().V1().Pods().Informer().AddEventHandler(
		cache.FilteringResourceEventHandler{
			FilterFunc: func(obj interface{}) bool { // 只watch已经被scheduled的pod
				switch t := obj.(type) {
				case *corev1.Pod:
					return assignedPod(t)
				case cache.DeletedFinalStateUnknown:
					if pod, ok := t.Obj.(*corev1.Pod); ok {
						return len(pod.Spec.NodeName) != 0
					}
					utilruntime.HandleError(fmt.Errorf("unable to convert object %T to *corev1.Pod in %T", obj, scheduler))
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
	// 还没有被调度的 pod，且 scheduler profile 里包含 pod.Spec.SchedulerName，放到调度队列里
	informerFactory.Core().V1().Pods().Informer().AddEventHandler(
		cache.FilteringResourceEventHandler{
			FilterFunc: func(obj interface{}) bool {
				switch t := obj.(type) {
				case *corev1.Pod:
					return !assignedPod(t) && responsibleForPod(t, scheduler.Profiles)
				case cache.DeletedFinalStateUnknown:
					if pod, ok := t.Obj.(*corev1.Pod); ok {
						return !assignedPod(pod) && responsibleForPod(pod, scheduler.Profiles)
					}
					utilruntime.HandleError(fmt.Errorf("unable to convert object %T to *corev1.Pod in %T", obj, scheduler))
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

	informerFactory.Storage().V1().StorageClasses().Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: scheduler.onStorageClassAdd,
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
}
