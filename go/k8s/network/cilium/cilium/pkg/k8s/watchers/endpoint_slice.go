package watchers

import (
	"github.com/cilium/cilium/pkg/lock"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func (k *K8sWatcher) watchEndpointSlices(k8sClient kubernetes.Interface) {
	_, endpointSliceController := cache.NewTransformingInformer(
		cache.NewListWatchFromClient(k8sClient.DiscoveryV1().RESTClient(),
			"endpointslices", corev1.NamespaceAll, fields.Everything()),
		&discoveryv1.EndpointSlice{},
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				var valid, equal bool
				defer func() { k.K8sEventReceived(metricEndpointSlice, metricCreate, valid, equal) }()

				k8sEndpointSlice, ok := obj.(*discoveryv1.EndpointSlice)
				if !ok {
					return
				}

				err := k.addK8sEndpointSlice(k8sEndpointSlice, swgSvcs)
				k.K8sEventProcessed(metricEndpointSlice, metricCreate, err == nil)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				var valid, equal bool
				defer func() { k.K8sEventReceived(metricEndpointSlice, metricUpdate, valid, equal) }()

				oldK8sEndpointSlice, ok := oldObj.(*discoveryv1.EndpointSlice)
				if !ok {
					return
				}
				newK8sEndpointSlice, ok := newObj.(*discoveryv1.EndpointSlice)
				if !ok {
					return
				}
				if EqualEndpointSlice(oldK8sEndpointSlice, newK8sEndpointSlice) {
					equal = true
					return
				}

				err := k.updateK8sEndpointSlice(oldK8sEndpoint, newK8sEndpoint, swgSvcs)
				k.K8sEventProcessed(metricEndpointSlice, metricUpdate, err == nil)
			},
			DeleteFunc: func(obj interface{}) {
				var valid, equal bool
				defer func() { k.K8sEventReceived(metricEndpointSlice, metricDelete, valid, equal) }()

				newK8sEndpointSlice, ok := obj.(*discoveryv1.EndpointSlice)
				if !ok {
					return
				}

				valid = true
				err := k.deleteK8sEndpointSlice(newK8sEndpointSlice, swgSvcs)
				k.K8sEventProcessed(metricEndpointSlice, metricDelete, err == nil)
			},
		},
		nil,
	)

	go endpointSliceController.Run(wait.NeverStop)

}

func (k *K8sWatcher) addK8sEndpointSlice(epSlice *discoveryv1.EndpointSlice, swg *lock.StoppableWaitGroup) error {
	k.K8sSvcCache.UpdateEndpointSlices(epSlice, swgEps)
	return nil
}

func (k *K8sWatcher) updateK8sEndpointSlice(oldEpSlice, newEpSlice *discoveryv1.EndpointSlice, swg *lock.StoppableWaitGroup) error {
	return k.addK8sEndpointSlice(newEpSlice, swg)
}

func (k *K8sWatcher) deleteK8sEndpointSlice(epSlice *discoveryv1.EndpointSlice, swg *lock.StoppableWaitGroup) error {
	k.K8sSvcCache.DeleteEndpointSlices(epSlice, swgEps)
	return nil
}
