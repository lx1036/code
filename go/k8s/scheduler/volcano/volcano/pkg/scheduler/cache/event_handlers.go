package cache

import (
	schedulingapi "k8s-lx1036/k8s/scheduler/volcano/volcano/pkg/scheduler/api"
	"slices"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

func (sc *SchedulerCache) AddNode(obj interface{}) {
	node, ok := obj.(*v1.Node)
	if !ok {
		klog.Errorf("Cannot convert to *v1.Node: %v", obj)
		return
	}
	sc.nodeQueue.Add(node.Name)
}

func (sc *SchedulerCache) DeleteNode(obj interface{}) {
	var node *v1.Node
	switch t := obj.(type) {
	case *v1.Node:
		node = t
	case cache.DeletedFinalStateUnknown:
		var ok bool
		node, ok = t.Obj.(*v1.Node)
		if !ok {
			klog.Errorf("Cannot convert to *v1.Node: %v", t.Obj)
			return
		}
	default:
		klog.Errorf("Cannot convert to *v1.Node: %v", t)
		return
	}
	sc.nodeQueue.Add(node.Name)
}

func (sc *SchedulerCache) UpdateNode(oldObj, newObj interface{}) {
	_, ok := oldObj.(*v1.Node)
	if !ok {
		klog.Errorf("Cannot convert oldObj to *v1.Node: %v", oldObj)
		return
	}
	newNode, ok := newObj.(*v1.Node)
	if !ok {
		klog.Errorf("Cannot convert newObj to *v1.Node: %v", newObj)
		return
	}
	sc.nodeQueue.Add(newNode.Name)
}

func (sc *SchedulerCache) SyncNode(nodeName string) error {
	node, err := sc.nodeInformer.Lister().Get(nodeName)
	if err != nil {
		if errors.IsNotFound(err) {
			deleteErr := sc.RemoveNode(nodeName)
			if deleteErr != nil {
				klog.Errorf("Failed to delete node <%s> and remove from cache: %s", nodeName, deleteErr.Error())
				return deleteErr
			}

			klog.V(3).Infof("Node <%s> was deleted, removed from cache.", nodeName)
			return nil
		}
		klog.Errorf("Failed to get node %s, error: %v", nodeName, err)
		return err
	}

	csiNode, err := sc.csiNodeInformer.Lister().Get(nodeName)
	if err == nil {
		sc.setCSIResourceOnNode(csiNode, node)
	} else if !errors.IsNotFound(err) {
		return err
	}

	return sc.AddOrUpdateNode(node)
}

// AddOrUpdateNode adds or updates node info in cache.
func (sc *SchedulerCache) AddOrUpdateNode(node *v1.Node) error {
	sc.Mutex.Lock()
	defer sc.Mutex.Unlock()

	if sc.Nodes[node.Name] != nil {
		sc.Nodes[node.Name].SetNode(node)
		sc.removeNodeImageStates(node.Name)
	} else {
		sc.Nodes[node.Name] = schedulingapi.NewNodeInfo(node)
	}
	sc.addNodeImageStates(node, sc.Nodes[node.Name])

	var nodeExisted bool
	for _, name := range sc.NodeList {
		if name == node.Name {
			nodeExisted = true
			break
		}
	}
	if !nodeExisted {
		sc.NodeList = append(sc.NodeList, node.Name)
	}
	return nil
}

// AddPod add pod to scheduler cache
func (sc *SchedulerCache) AddPod(obj interface{}) {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		klog.Errorf("Cannot convert to *v1.Pod: %v", obj)
		return
	}

	sc.Mutex.Lock()
	defer sc.Mutex.Unlock()

	err := sc.addPod(pod)
	if err != nil {
		klog.Errorf("Failed to add pod <%s/%s> into cache: %v",
			pod.Namespace, pod.Name, err)
		return
	}
	klog.V(3).Infof("Added pod <%s/%v> into cache.", pod.Namespace, pod.Name)
}

// UpdatePod update pod to scheduler cache
func (sc *SchedulerCache) UpdatePod(oldObj, newObj interface{}) {
	oldPod, ok := oldObj.(*v1.Pod)
	if !ok {
		klog.Errorf("Cannot convert oldObj to *v1.Pod: %v", oldObj)
		return
	}
	newPod, ok := newObj.(*v1.Pod)
	if !ok {
		klog.Errorf("Cannot convert newObj to *v1.Pod: %v", newObj)
		return
	}

	sc.Mutex.Lock()
	defer sc.Mutex.Unlock()

	err := sc.updatePod(oldPod, newPod)
	if err != nil {
		klog.Errorf("Failed to update pod %v in cache: %v", oldPod.Name, err)
		return
	}

	klog.V(4).Infof("Updated pod <%s/%v> in cache.", oldPod.Namespace, oldPod.Name)
}

// DeletePod delete pod from scheduler cache
func (sc *SchedulerCache) DeletePod(obj interface{}) {
	var pod *v1.Pod
	switch t := obj.(type) {
	case *v1.Pod:
		pod = t
	case cache.DeletedFinalStateUnknown:
		var ok bool
		pod, ok = t.Obj.(*v1.Pod)
		if !ok {
			klog.Errorf("Cannot convert to *v1.Pod: %v", t.Obj)
			return
		}
	default:
		klog.Errorf("Cannot convert to *v1.Pod: %v", t)
		return
	}

	sc.Mutex.Lock()
	defer sc.Mutex.Unlock()

	err := sc.deletePod(pod)
	if err != nil {
		klog.Errorf("Failed to delete pod %v from cache: %v", pod.Name, err)
		return
	}

	klog.V(3).Infof("Deleted pod <%s/%v> from cache.", pod.Namespace, pod.Name)
}

// Assumes that lock is already acquired.
func (sc *SchedulerCache) addPod(pod *v1.Pod) error {
	pi, err := sc.NewTaskInfo(pod)
	if err != nil {
		klog.Errorf("generate taskInfo for pod(%s) failed: %v", pod.Name, err)
		sc.resyncTask(pi)
	}

	return sc.addTask(pi)
}

func (sc *SchedulerCache) NewTaskInfo(pod *v1.Pod) (*schedulingapi.TaskInfo, error) {
	taskInfo := schedulingapi.NewTaskInfo(pod)
	//if err := sc.addPodCSIVolumesToTask(taskInfo); err != nil {
	//	return taskInfo, err
	//}
	// Update BestEffort because the InitResreq maybe changes
	taskInfo.BestEffort = taskInfo.InitResreq.IsEmpty()
	return taskInfo, nil
}

func (sc *SchedulerCache) addTask(taskInfo *schedulingapi.TaskInfo) error {
	if len(taskInfo.NodeName) != 0 {
		// TODO
	}

	job := sc.getOrCreateJob(taskInfo)
	if job != nil {
		job.AddTaskInfo(taskInfo)
	}

	return nil
}

// getOrCreateJob will return corresponding Job for pi if it exists, or it will create a Job and return it if
// pi.Pod.Spec.SchedulerName is same as volcano scheduler's name, otherwise it will return nil.
func (sc *SchedulerCache) getOrCreateJob(pi *schedulingapi.TaskInfo) *schedulingapi.JobInfo {
	if len(pi.Job) == 0 {
		if !slices.Contains(sc.schedulerNames, pi.Pod.Spec.SchedulerName) {
			klog.V(4).Infof("Pod %s/%s will not scheduled by %#v, skip creating PodGroup and Job for it",
				pi.Pod.Namespace, pi.Pod.Name, sc.schedulerNames)
		}
		return nil
	}

	if _, found := sc.Jobs[pi.Job]; !found {
		sc.Jobs[pi.Job] = schedulingapi.NewJobInfo(pi.Job)
	}

	return sc.Jobs[pi.Job]
}
