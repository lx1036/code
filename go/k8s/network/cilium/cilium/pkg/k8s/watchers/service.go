package watchers

import (
	"github.com/cilium/cilium/pkg/lock"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func (k *K8sWatcher) watchK8sService(k8sClient kubernetes.Interface) {

	_, svcController := cache.NewTransformingInformer(
		cache.NewListWatchFromClient(k8sClient.CoreV1().RESTClient(),
			"services", corev1.NamespaceAll, fields.Everything()),
		&corev1.Service{},
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				var valid, equal bool
				defer func() { k.K8sEventReceived(metricService, metricCreate, valid, equal) }()

				k8sSvc, ok := obj.(*corev1.Service)
				if !ok {
					return
				}

				err := k.addK8sService(k8sSvc, swgSvcs)
				k.K8sEventProcessed(metricService, metricCreate, err == nil)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				var valid, equal bool
				defer func() { k.K8sEventReceived(metricService, metricUpdate, valid, equal) }()

				oldK8sSvc, ok := oldObj.(*corev1.Service)
				if !ok {
					return
				}
				newK8sSvc, ok := newObj.(*corev1.Service)
				if !ok {
					return
				}
				if EqualService(oldK8sSvc, newK8sSvc) {
					equal = true
					return
				}

				err := k.updateK8sService(oldk8sSvc, newk8sSvc, swgSvcs)
				k.K8sEventProcessed(metricService, metricUpdate, err == nil)
			},
			DeleteFunc: func(obj interface{}) {
				var valid, equal bool
				defer func() { k.K8sEventReceived(metricService, metricDelete, valid, equal) }()

				k8sSvc := ObjToV1Services(obj)
				if k8sSvc == nil {
					return
				}

				valid = true
				err := k.deleteK8sService(k8sSvc, swgSvcs)
				k.K8sEventProcessed(metricService, metricDelete, err == nil)
			},
		},
		nil,
	)

	go svcController.Run(wait.NeverStop)
}

func (k *K8sWatcher) addK8sService(svc *corev1.Service, swg *lock.StoppableWaitGroup) error {
	k.K8sSvcCache.UpdateService(svc, swg)
	return nil
}

func (k *K8sWatcher) updateK8sService(oldSvc, newSvc *corev1.Service, swg *lock.StoppableWaitGroup) error {
	return k.addK8sService(newSvc, swg)
}

func (k *K8sWatcher) deleteK8sService(svc *corev1.Service, swg *lock.StoppableWaitGroup) error {
	k.K8sSvcCache.DeleteService(svc, swg)
	return nil
}

// ServiceID identifies the Kubernetes service
type ServiceID struct {
	Name      string `json:"serviceName,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

// Service is an abstraction for a k8s service that is composed by the frontend IP
// address (FEIP) and the map of the frontend ports (Ports).
// +k8s:deepcopy-gen=true
type Service struct {
}

func ParseService(svc *corev1.Service, nodeAddressing datapath.NodeAddressing) (ServiceID, *Service) {

}

func ObjToV1Services(obj interface{}) *corev1.Service {
	svc, ok := obj.(*corev1.Service)
	if ok {
		return svc
	}
	deletedObj, ok := obj.(cache.DeletedFinalStateUnknown)
	if ok {
		// Delete was not observed by the watcher but is
		// removed from kube-apiserver. This is the last
		// known state and the object no longer exists.
		svc, ok := deletedObj.Obj.(*corev1.Service)
		if ok {
			return svc
		}
	}
	return nil
}
