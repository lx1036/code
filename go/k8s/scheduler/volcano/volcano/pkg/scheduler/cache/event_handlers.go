package cache

import (
	schedulingapi "k8s-lx1036/k8s/scheduler/volcano/volcano/pkg/scheduler/api"

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

}
