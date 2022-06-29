package watchers

import (
	"github.com/cilium/cilium/pkg/lock"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func (k *K8sWatcher) watchK8sEndpoints(k8sClient kubernetes.Interface) {
	_, endpointController := cache.NewTransformingInformer(
		cache.NewListWatchFromClient(k8sClient.CoreV1().RESTClient(),
			"endpoints", corev1.NamespaceAll, fields.Everything()),
		&corev1.Endpoints{},
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				var valid, equal bool
				defer func() { k.K8sEventReceived(metricEndpoint, metricCreate, valid, equal) }()

				k8sEndpoint, ok := obj.(*corev1.Endpoints)
				if !ok {
					return
				}

				err := k.addK8sEndpoint(k8sEndpoint, swgSvcs)
				k.K8sEventProcessed(metricEndpoint, metricCreate, err == nil)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				var valid, equal bool
				defer func() { k.K8sEventReceived(metricEndpoint, metricUpdate, valid, equal) }()

				oldK8sEndpoint, ok := oldObj.(*corev1.Endpoints)
				if !ok {
					return
				}
				newK8sEndpoint, ok := newObj.(*corev1.Endpoints)
				if !ok {
					return
				}
				if EqualEndpoint(oldK8sEndpoint, newK8sEndpoint) {
					equal = true
					return
				}

				err := k.updateK8sEndpoint(oldK8sEndpoint, newK8sEndpoint, swgSvcs)
				k.K8sEventProcessed(metricEndpoint, metricUpdate, err == nil)
			},
			DeleteFunc: func(obj interface{}) {
				var valid, equal bool
				defer func() { k.K8sEventReceived(metricEndpoint, metricDelete, valid, equal) }()

				k8sEndpoint := ObjToV1Endpoint(obj)
				if k8sEndpoint == nil {
					return
				}

				valid = true
				err := k.deleteK8sEndpoint(k8sEndpoint, swgSvcs)
				k.K8sEventProcessed(metricEndpoint, metricDelete, err == nil)
			},
		},
		nil,
	)

	go endpointController.Run(wait.NeverStop)

}

func (k *K8sWatcher) addK8sEndpoint(ep *corev1.Endpoints, swg *lock.StoppableWaitGroup) error {
	k.K8sSvcCache.UpdateEndpoints(ep, swg)
	return nil
}

func (k *K8sWatcher) updateK8sEndpoint(oldEP, newEP *corev1.Endpoints, swg *lock.StoppableWaitGroup) error {
	return k.addK8sEndpoint(newEP, swg)
}

func (k *K8sWatcher) deleteK8sEndpoint(ep *corev1.Endpoints, swg *lock.StoppableWaitGroup) error {
	k.K8sSvcCache.DeleteEndpoints(ep, swg)
	return nil
}

func ObjToV1Endpoint(obj interface{}) *corev1.Endpoints {
	endpoint, ok := obj.(*corev1.Endpoints)
	if ok {
		return endpoint
	}
	deletedObj, ok := obj.(cache.DeletedFinalStateUnknown)
	if ok {
		// Delete was not observed by the watcher but is
		// removed from kube-apiserver. This is the last
		// known state and the object no longer exists.
		endpoint, ok := deletedObj.Obj.(*corev1.Endpoints)
		if ok {
			return endpoint
		}
	}

	return nil
}
