package capacity_scheduling

import (
	"context"
	"fmt"
	"sync"

	"k8s-lx1036/k8s/scheduler/pkg/scheduler/apis/scheduling/config"
	"k8s-lx1036/k8s/scheduler/pkg/scheduler/apis/scheduling/v1alpha1"
	"k8s-lx1036/k8s/scheduler/pkg/scheduler/client/clientset/versioned"
	schedinformer "k8s-lx1036/k8s/scheduler/pkg/scheduler/client/informers/externalversions"
	externalv1alpha1 "k8s-lx1036/k8s/scheduler/pkg/scheduler/client/listers/scheduling/v1alpha1"
	//framework "k8s-lx1036/k8s/scheduler/pkg/scheduler/framework"
	framework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	corelisters "k8s.io/client-go/listers/core/v1"
	policylisters "k8s.io/client-go/listers/policy/v1beta1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

// INFO: https://github.com/kubernetes-sigs/scheduler-plugins/blob/v0.19.9/pkg/capacityscheduling/capacity_scheduling.go

const (
	// Name is the name of the plugin used in Registry and configurations.
	Name = "CapacityScheduling"
)

type CapacityScheduling struct {
	sync.RWMutex
	frameworkHandle    framework.FrameworkHandle
	podLister          corelisters.PodLister
	pdbLister          policylisters.PodDisruptionBudgetLister
	elasticQuotaLister externalv1alpha1.ElasticQuotaLister
	elasticQuotaInfos  ElasticQuotaInfos
}

func (c *CapacityScheduling) Name() string {
	return Name
}

func (c *CapacityScheduling) PreFilterExtensions() framework.PreFilterExtensions {
	return nil
}

func (c *CapacityScheduling) PreFilter(ctx context.Context, cycleState *framework.CycleState, pod *v1.Pod) *framework.Status {
	return nil
}

func (c *CapacityScheduling) addElasticQuota(obj interface{}) {
	elasticQuota := obj.(*v1alpha1.ElasticQuota)
	oldElasticQuotaInfo := c.elasticQuotaInfos[elasticQuota.Namespace]
	if oldElasticQuotaInfo != nil {
		return
	}

	elasticQuotaInfo := newElasticQuotaInfo(elasticQuota.Namespace, elasticQuota.Spec.Min, elasticQuota.Spec.Max, nil)

	c.Lock()
	defer c.Unlock()

	klog.Info(fmt.Sprintf("adding ElasticQuota %s/%s in cache", elasticQuota.Namespace, elasticQuota.Name))
	c.elasticQuotaInfos[elasticQuota.Namespace] = elasticQuotaInfo
}

func (c *CapacityScheduling) updateElasticQuota(oldObj, newObj interface{}) {
	oldEQ := oldObj.(*v1alpha1.ElasticQuota)
	newEQ := newObj.(*v1alpha1.ElasticQuota)
	newEQInfo := newElasticQuotaInfo(newEQ.Namespace, newEQ.Spec.Min, newEQ.Spec.Max, nil)

	c.Lock()
	defer c.Unlock()

	oldEQInfo := c.elasticQuotaInfos[oldEQ.Namespace]
	if oldEQInfo != nil {
		newEQInfo.pods = oldEQInfo.pods
		newEQInfo.Used = oldEQInfo.Used
	}

	c.elasticQuotaInfos[newEQ.Namespace] = newEQInfo
}

func (c *CapacityScheduling) deleteElasticQuota(obj interface{}) {
	elasticQuota := obj.(*v1alpha1.ElasticQuota)
	c.Lock()
	defer c.Unlock()

	klog.Info(fmt.Sprintf("deleting ElasticQuota %s/%s in cache", elasticQuota.Namespace, elasticQuota.Name))
	delete(c.elasticQuotaInfos, elasticQuota.Namespace)
}

func (c *CapacityScheduling) addPod(obj interface{}) {
	pod := obj.(*v1.Pod)
	c.Lock()
	defer c.Unlock()

	elasticQuotaInfo := c.elasticQuotaInfos[pod.Namespace]
	// If elasticQuotaInfo is nil, try to list ElasticQuotas through elasticQuotaLister
	if elasticQuotaInfo == nil {
		eqs, err := c.elasticQuotaLister.ElasticQuotas(pod.Namespace).List(labels.NewSelector())
		if err != nil {
			klog.Errorf("Get ElasticQuota %v error %v", pod.Namespace, err)
			return
		}

		// If the length of elasticQuotas is 0, return.
		if len(eqs) == 0 {
			return
		}

		if len(eqs) > 0 {
			// only one elasticquota is supported in each namespace
			eq := eqs[0]
			elasticQuotaInfo = newElasticQuotaInfo(eq.Namespace, eq.Spec.Min, eq.Spec.Max, nil)
			c.elasticQuotaInfos[eq.Namespace] = elasticQuotaInfo
		}
	}

	err := elasticQuotaInfo.addPodIfNotPresent(pod)
	if err != nil {
		klog.Errorf("ElasticQuota addPodIfNotPresent for pod %v/%v error %v", pod.Namespace, pod.Name, err)
	}
}

// INFO: 只考虑 pod 开始创建时的更新那段时间，但是这里逻辑在考虑什么？？？
func (c *CapacityScheduling) updatePod(oldObj, newObj interface{}) {
	oldPod := oldObj.(*v1.Pod)
	newPod := newObj.(*v1.Pod)

	if oldPod.Status.Phase == v1.PodSucceeded || oldPod.Status.Phase == v1.PodFailed {
		return
	}

	if newPod.Status.Phase != v1.PodRunning && newPod.Status.Phase != v1.PodPending {
		c.Lock()
		defer c.Unlock()

		elasticQuotaInfo := c.elasticQuotaInfos[newPod.Namespace]
		if elasticQuotaInfo != nil {
			// INFO: 从 elasticQuotaInfo cache 中删除
			err := elasticQuotaInfo.deletePodIfPresent(newPod)
			if err != nil {
				klog.Errorf("ElasticQuota deletePodIfPresent for pod %v/%v error %v", newPod.Namespace, newPod.Name, err)
			}
		}
	}
}

func (c *CapacityScheduling) deletePod(obj interface{}) {
	pod := obj.(*v1.Pod)
	c.Lock()
	defer c.Unlock()

	elasticQuotaInfo := c.elasticQuotaInfos[pod.Namespace]
	if elasticQuotaInfo != nil {
		err := elasticQuotaInfo.deletePodIfPresent(pod)
		if err != nil {
			klog.Errorf("ElasticQuota deletePodIfPresent for pod %v/%v error %v", pod.Namespace, pod.Name, err)
		}
	}
}

func New(obj runtime.Object, handle framework.FrameworkHandle) (framework.Plugin, error) {
	args, ok := obj.(*config.CapacitySchedulingArgs)
	if !ok {
		return nil, fmt.Errorf("want args to be of type CapacitySchedulingArgs, got %T", obj)
	}
	kubeConfigPath := args.KubeConfigPath

	c := &CapacityScheduling{
		frameworkHandle:   handle,
		elasticQuotaInfos: NewElasticQuotaInfos(),
		//pdbLister:         getPDBLister(handle.SharedInformerFactory()),
	}

	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, err
	}
	elasticQuotaClient, err := versioned.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	schedSharedInformerFactory := schedinformer.NewSharedInformerFactory(elasticQuotaClient, 0)
	c.elasticQuotaLister = schedSharedInformerFactory.Scheduling().V1alpha1().ElasticQuotas().Lister()
	elasticQuotaInformer := schedSharedInformerFactory.Scheduling().V1alpha1().ElasticQuotas().Informer()
	elasticQuotaInformer.AddEventHandler(
		cache.FilteringResourceEventHandler{
			FilterFunc: func(obj interface{}) bool {
				switch t := obj.(type) {
				case *v1alpha1.ElasticQuota:
					return true
				case cache.DeletedFinalStateUnknown:
					if _, ok := t.Obj.(*v1alpha1.ElasticQuota); ok {
						return true
					}
					utilruntime.HandleError(fmt.Errorf("cannot convert to *v1alpha1.ElasticQuota: %v", obj))
					return false
				default:
					utilruntime.HandleError(fmt.Errorf("unable to handle object in %T", obj))
					return false
				}
			},
			Handler: cache.ResourceEventHandlerFuncs{
				AddFunc:    c.addElasticQuota,
				UpdateFunc: c.updateElasticQuota,
				DeleteFunc: c.deleteElasticQuota,
			},
		})

	schedSharedInformerFactory.Start(nil)
	if !cache.WaitForCacheSync(nil, elasticQuotaInformer.HasSynced) {
		return nil, fmt.Errorf("timed out waiting for caches to sync %v", Name)
	}

	podInformer := handle.SharedInformerFactory().Core().V1().Pods().Informer()
	podInformer.AddEventHandler(
		cache.FilteringResourceEventHandler{
			FilterFunc: func(obj interface{}) bool {
				switch t := obj.(type) {
				case *v1.Pod:
					return assignedPod(t)
				case cache.DeletedFinalStateUnknown:
					if pod, ok := t.Obj.(*v1.Pod); ok {
						return assignedPod(pod)
					}
					return false
				default:
					return false
				}
			},
			Handler: cache.ResourceEventHandlerFuncs{
				AddFunc:    c.addPod,
				UpdateFunc: c.updatePod,
				DeleteFunc: c.deletePod,
			},
		},
	)

	klog.Infof("CapacityScheduling start")

	return c, nil
}

// assignedPod selects pods that are assigned (scheduled and running).
func assignedPod(pod *v1.Pod) bool {
	return len(pod.Spec.NodeName) != 0
}
