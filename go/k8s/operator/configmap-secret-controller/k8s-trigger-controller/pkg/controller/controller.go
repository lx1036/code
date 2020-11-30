package controller

import (
	"k8s-lx1036/k8s/operator/configmap-secret-controller/k8s-trigger-controller/pkg/metrics"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type Controller struct {
	client     *kubernetes.Clientset
	namespace  string
	collectors metrics.Collectors

	queue workqueue.RateLimitingInterface
}

type Kind string

const (
	ConfigMap Kind = "configMap"
	Secret    Kind = "secret"
)

type item struct {
	object interface{}
	kind   Kind
	key    string
}

var (
	accessor = meta.NewAccessor()
)

func NewController(informerFactory informers.SharedInformerFactory, client *kubernetes.Clientset, collectors metrics.Collectors, namespace string) (*Controller, error) {

	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "configmap-secret")

	controller := &Controller{
		queue: queue,
	}

	configMapInformer := informerFactory.Core().V1().ConfigMaps()
	configMapInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			controller.Enqueue(&item{
				object: obj,
				kind:   ConfigMap,
			})
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			o, _ := accessor.ResourceVersion(oldObj.(runtime.Object))
			n, _ := accessor.ResourceVersion(newObj.(runtime.Object))
			// Only enqueue changes that have a different resource versions to avoid processing resyncs.
			if o != n {
				controller.Enqueue(&item{
					object: newObj,
					kind:   ConfigMap,
				})
			}
		},
	})

	secretInformer := informerFactory.Core().V1().Secrets()
	secretInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			controller.Enqueue(&item{
				object: obj,
				kind:   Secret,
			})
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			o, _ := accessor.ResourceVersion(oldObj.(runtime.Object))
			n, _ := accessor.ResourceVersion(newObj.(runtime.Object))
			// Only enqueue changes that have a different resource versions to avoid processing resyncs.
			if o != n {
				controller.Enqueue(&item{
					object: newObj,
					kind:   Secret,
				})
			}
		},
	})

	return controller, nil
}

func (controller *Controller) Enqueue(item *item) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(item.object)
	if err != nil {
		return
	}
	item.key = key
	controller.queue.Add(item)
}

func Run(threadiness int, stopCh <-chan struct{}) error {

}
