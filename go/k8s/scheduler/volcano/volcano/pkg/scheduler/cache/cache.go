package cache

import (
	"k8s-lx1036/k8s/scheduler/volcano/volcano/pkg/features"
	"k8s.io/client-go/kubernetes"
	"strings"

	v1 "k8s.io/api/core/v1"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/client-go/informers"
	infov1 "k8s.io/client-go/informers/core/v1"
	schedv1 "k8s.io/client-go/informers/scheduling/v1"
	storagev1 "k8s.io/client-go/informers/storage/v1"
	storagev1beta1 "k8s.io/client-go/informers/storage/v1beta1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

type SchedulerCache struct {
	kubeClient                 kubernetes.Interface
	informerFactory            informers.SharedInformerFactory
	podInformer                infov1.PodInformer
	nodeInformer               infov1.NodeInformer
	podGroupInformerV1beta1    vcinformerv1.PodGroupInformer
	queueInformerV1beta1       vcinformerv1.QueueInformer
	pvInformer                 infov1.PersistentVolumeInformer
	pvcInformer                infov1.PersistentVolumeClaimInformer
	scInformer                 storagev1.StorageClassInformer
	pcInformer                 schedv1.PriorityClassInformer
	quotaInformer              infov1.ResourceQuotaInformer
	csiNodeInformer            storagev1.CSINodeInformer
	csiDriverInformer          storagev1.CSIDriverInformer
	csiStorageCapacityInformer storagev1beta1.CSIStorageCapacityInformer

	nodeSelectorLabels map[string]string
}

func (sc *SchedulerCache) Run(stopCh <-chan struct{}) {

}

func (sc *SchedulerCache) addEventHandler() {
	informerFactory := informers.NewSharedInformerFactory(sc.kubeClient, 0)
	sc.informerFactory = informerFactory
	// explicitly register informers to the factory, otherwise resources listers cannot get anything
	// even with no error returned.
	// `Namespace` informer is used by `InterPodAffinity` plugin,
	// `SelectorSpread` and `PodTopologySpread` plugins uses the following four so far.
	informerFactory.Core().V1().Namespaces().Informer()
	informerFactory.Core().V1().Services().Informer()
	if utilfeature.DefaultFeatureGate.Enabled(features.WorkLoadSupport) {
		informerFactory.Core().V1().ReplicationControllers().Informer()
		informerFactory.Apps().V1().ReplicaSets().Informer()
		informerFactory.Apps().V1().StatefulSets().Informer()
	}

	// `PodDisruptionBudgets` informer is used by `Pdb` plugin
	if utilfeature.DefaultFeatureGate.Enabled(features.PodDisruptionBudgetsSupport) {
		informerFactory.Policy().V1().PodDisruptionBudgets().Informer()
	}

	// create informer for node information
	sc.nodeInformer = informerFactory.Core().V1().Nodes()
	sc.nodeInformer.Informer().AddEventHandlerWithResyncPeriod(
		cache.FilteringResourceEventHandler{
			FilterFunc: func(obj interface{}) bool {
				var node *v1.Node
				switch t := obj.(type) {
				case *v1.Node:
					node = t
				case cache.DeletedFinalStateUnknown:
					var ok bool
					node, ok = t.Obj.(*v1.Node)
					if !ok {
						klog.Errorf("Cannot convert to *v1.Node: %v", t.Obj)
						return false
					}
				default:
					return false
				}

				if !responsibleForNode(node.Name, mySchedulerPodName, c) {
					return false
				}
				if len(sc.nodeSelectorLabels) == 0 {
					return true
				}
				for labelName, labelValue := range node.Labels {
					key := labelName + ":" + labelValue
					if _, ok := sc.nodeSelectorLabels[key]; ok {
						return true
					}
				}
				klog.Infof("node %s ignore add/update/delete into schedulerCache", node.Name)
				return false
			},
			Handler: cache.ResourceEventHandlerFuncs{
				AddFunc:    sc.AddNode,
				UpdateFunc: sc.UpdateNode,
				DeleteFunc: sc.DeleteNode,
			},
		},
		0,
	)

	sc.podInformer = informerFactory.Core().V1().Pods()
	sc.pvcInformer = informerFactory.Core().V1().PersistentVolumeClaims()
	sc.pvInformer = informerFactory.Core().V1().PersistentVolumes()
	sc.scInformer = informerFactory.Storage().V1().StorageClasses()
	sc.csiNodeInformer = informerFactory.Storage().V1().CSINodes()
	sc.csiNodeInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    sc.AddOrUpdateCSINode,
			UpdateFunc: sc.UpdateCSINode,
			DeleteFunc: sc.DeleteCSINode,
		},
	)

}

// --node-selector=volcano.sh/role:train --node-selector=volcano.sh/role:serving
func (sc *SchedulerCache) updateNodeSelectors(nodeSelectors []string) {
	for _, nodeSelectorLabel := range nodeSelectors {
		if len(nodeSelectorLabel) == 0 {
			continue
		}

		// check input
		index := strings.Index(nodeSelectorLabel, ":")
		if index < 0 || index >= (len(nodeSelectorLabel)-1) {
			continue
		}

		nodeSelectorLabelName := strings.TrimSpace(nodeSelectorLabel[:index])
		nodeSelectorLabelValue := strings.TrimSpace(nodeSelectorLabel[index+1:])
		key := nodeSelectorLabelName + ":" + nodeSelectorLabelValue
		sc.nodeSelectorLabels[key] = ""
	}
}

func New(config *rest.Config, schedulerNames []string, defaultQueue string, nodeSelectors []string,
	nodeWorkers uint32, ignoredProvisioners []string) Cache {
	return newSchedulerCache(config, schedulerNames, defaultQueue, nodeSelectors, nodeWorkers, ignoredProvisioners)
}

func newSchedulerCache(config *rest.Config, schedulerNames []string, defaultQueue string, nodeSelectors []string,
	nodeWorkers uint32, ignoredProvisioners []string) *SchedulerCache {

	sc := &SchedulerCache{}

	if len(nodeSelectors) > 0 {
		sc.updateNodeSelectors(nodeSelectors)
	}

	sc.addEventHandler()

	return sc
}
