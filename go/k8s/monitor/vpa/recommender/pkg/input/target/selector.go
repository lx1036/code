package target

import (
	"fmt"
	"time"

	apisv1 "k8s-lx1036/k8s/monitor/vpa/recommender/pkg/apis/autoscaling.k9s.io/v1"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/discovery"
	cacheddiscovery "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/scale"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

const (
	discoveryResetPeriod = 5 * time.Minute
)

type wellKnownController string

const (
	daemonSet             wellKnownController = "DaemonSet"
	deployment            wellKnownController = "Deployment"
	replicaSet            wellKnownController = "ReplicaSet"
	statefulSet           wellKnownController = "StatefulSet"
	replicationController wellKnownController = "ReplicationController"
	job                   wellKnownController = "Job"
	cronJob               wellKnownController = "CronJob"
)

type VpaTargetSelectorFetcher struct {
	scaleNamespacer scale.ScalesGetter
	mapper          meta.RESTMapper
	informersMap    map[wellKnownController]cache.SharedIndexInformer
}

func NewVpaTargetSelectorFetcher(config *rest.Config, kubeClient kubernetes.Interface,
	factory informers.SharedInformerFactory) *VpaTargetSelectorFetcher {

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		klog.Fatalf("Could not create discoveryClient: %v", err)
	}
	cachedDiscoveryClient := cacheddiscovery.NewMemCacheClient(discoveryClient)
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(cachedDiscoveryClient)
	go wait.Until(func() {
		mapper.Reset()
	}, discoveryResetPeriod, make(chan struct{}))

	informersMap := map[wellKnownController]cache.SharedIndexInformer{
		daemonSet:             factory.Apps().V1().DaemonSets().Informer(),
		deployment:            factory.Apps().V1().Deployments().Informer(),
		replicaSet:            factory.Apps().V1().ReplicaSets().Informer(),
		statefulSet:           factory.Apps().V1().StatefulSets().Informer(),
		replicationController: factory.Core().V1().ReplicationControllers().Informer(),
		job:                   factory.Batch().V1().Jobs().Informer(),
		cronJob:               factory.Batch().V1beta1().CronJobs().Informer(),
	}
	for kind, informer := range informersMap {
		stopCh := make(chan struct{})
		go informer.Run(stopCh)
		synced := cache.WaitForCacheSync(stopCh, informer.HasSynced)
		if !synced {
			klog.Fatalf("Could not sync cache for %s: %v", kind, err)
		} else {
			klog.Infof("Initial sync of %s completed", kind)
		}
	}

	return &VpaTargetSelectorFetcher{
		scaleNamespacer: scaleNamespacer,
		mapper:          mapper,
		informersMap:    informersMap,
	}
}

// Fetch
/*
apiVersion: "autoscaling.k9s.io/v1"
kind: VerticalPodAutoscaler
metadata:
  name: hamster-vpa
spec:
  targetRef:
    apiVersion: "apps/v1"
    kind: Deployment
    name: hamster
  resourcePolicy:
    containerPolicies:
      - containerName: '*'
        minAllowed:
          cpu: 100m
          memory: 50Mi
        maxAllowed:
          cpu: 1
          memory: 500Mi
        controlledResources: ["cpu", "memory"]
*/
func (f *VpaTargetSelectorFetcher) Fetch(vpa *apisv1.VerticalPodAutoscaler) (labels.Selector, error) {
	if vpa.Spec.TargetRef == nil {
		return nil, fmt.Errorf("targetRef not defined. If this is a v1beta1 object switch to v1beta2.")
	}
	kind := wellKnownController(vpa.Spec.TargetRef.Kind)
	informer, exists := f.informersMap[kind]
	if exists {
		return getLabelSelector(informer, vpa.Spec.TargetRef.Kind, vpa.Namespace, vpa.Spec.TargetRef.Name)
	}

	groupVersion, err := schema.ParseGroupVersion(vpa.Spec.TargetRef.APIVersion)
	if err != nil {
		return nil, err
	}
	groupKind := schema.GroupKind{
		Group: groupVersion.Group,
		Kind:  vpa.Spec.TargetRef.Kind,
	}

	selector, err := f.getLabelSelectorFromResource(groupKind, vpa.Namespace, vpa.Spec.TargetRef.Name)
	if err != nil {
		return nil, fmt.Errorf("unhandled targetRef %s / %s / %s, last error %v",
			vpa.Spec.TargetRef.APIVersion, vpa.Spec.TargetRef.Kind, vpa.Spec.TargetRef.Name, err)
	}
	return selector, nil
}

func getLabelSelector(informer cache.SharedIndexInformer, kind, namespace, name string) (labels.Selector, error) {
	obj, exists, err := informer.GetStore().GetByKey(namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("%s %s/%s does not exist", kind, namespace, name)
	}

	switch obj.(type) {
	case *appsv1.DaemonSet:
		apiObj, ok := obj.(*appsv1.DaemonSet)
		if !ok {
			return nil, fmt.Errorf("failed to parse %s %s/%s", kind, namespace, name)
		}
		return metav1.LabelSelectorAsSelector(apiObj.Spec.Selector)
	case *appsv1.Deployment:
		apiObj, ok := obj.(*appsv1.Deployment)
		if !ok {
			return nil, fmt.Errorf("failed to parse %s %s/%s", kind, namespace, name)
		}
		return metav1.LabelSelectorAsSelector(apiObj.Spec.Selector)
	case *appsv1.StatefulSet:
		apiObj, ok := obj.(*appsv1.StatefulSet)
		if !ok {
			return nil, fmt.Errorf("failed to parse %s %s/%s", kind, namespace, name)
		}
		return metav1.LabelSelectorAsSelector(apiObj.Spec.Selector)
	case *appsv1.ReplicaSet:
		apiObj, ok := obj.(*appsv1.ReplicaSet)
		if !ok {
			return nil, fmt.Errorf("failed to parse %s %s/%s", kind, namespace, name)
		}
		return metav1.LabelSelectorAsSelector(apiObj.Spec.Selector)
	case *batchv1.Job:
		apiObj, ok := obj.(*batchv1.Job)
		if !ok {
			return nil, fmt.Errorf("failed to parse %s %s/%s", kind, namespace, name)
		}
		return metav1.LabelSelectorAsSelector(apiObj.Spec.Selector)
	case *batchv1beta1.CronJob:
		apiObj, ok := obj.(*batchv1beta1.CronJob)
		if !ok {
			return nil, fmt.Errorf("failed to parse %s %s/%s", kind, namespace, name)
		}
		return metav1.LabelSelectorAsSelector(metav1.SetAsLabelSelector(apiObj.Spec.JobTemplate.Spec.Template.Labels))
	case *corev1.ReplicationController:
		apiObj, ok := obj.(*corev1.ReplicationController)
		if !ok {
			return nil, fmt.Errorf("Failed to parse %s %s/%s", kind, namespace, name)
		}
		return metav1.LabelSelectorAsSelector(metav1.SetAsLabelSelector(apiObj.Spec.Selector))
	}

	return nil, fmt.Errorf("don't know how to read label seletor")
}

func (f *VpaTargetSelectorFetcher) getLabelSelectorFromResource(groupKind schema.GroupKind, namespace, name string) (labels.Selector, error) {
	mappings, err := f.mapper.RESTMappings(groupKind)
	if err != nil {
		return nil, err
	}

}
