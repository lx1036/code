package controller

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"k8s-lx1036/k8s/storage/log/filebeat/daemonset-operator/pkg/metrics"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const Name = "filebeat"

type LogController struct {
	InformerResources []schema.GroupVersionResource
	ApiServerClient   kubernetes.Interface
	PodStore          cache.Store
	PodInformer       cache.Controller
	NodeStore         cache.Store
	NodeInformer      cache.Controller

	TaskQueue *TaskQueue

	stopCh chan struct{}

	NodeName         string
	ResyncPeriod     time.Duration
	TaskHandlePeriod time.Duration
}

var (
	accessor = meta.NewAccessor()
)

const defaultNode = "localhost"

func InClusterNamespace() (string, error) {
	data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func DiscoverKubernetesNode(host string, client kubernetes.Interface) string {
	if len(host) != 0 {
		log.Infof("Using node %s provided by env", host)
		return host
	}

	// node discover by pod
	ns, err := InClusterNamespace()
	if err != nil {
		log.Errorf("Can't get namespace in cluster with error: %v", err)
		return defaultNode
	}
	podName, err := os.Hostname()
	if err != nil {
		log.Errorf("Can't get hostname as pod name in cluster with error: %v", err)
		return defaultNode
	}

	pod, err := client.CoreV1().Pods(ns).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		log.Errorf("Can't query pod in cluster with error: %v", err)
		return defaultNode
	}

	log.Infof("Using node %s discovered by pod in cluster", pod.Spec.NodeName)

	return pod.Spec.NodeName
}

type Controller struct {
	queue           workqueue.RateLimitingInterface
	informerFactory informers.SharedInformerFactory
	client          *kubernetes.Clientset

	podInformer  cache.SharedIndexInformer
	nodeInformer cache.SharedIndexInformer
}

// Resource data
type Resource = runtime.Object
type WatchOptions struct {
	Namespace string
	Node      string

	// SyncTimeout is timeout for listing historical resources
	SyncTimeout time.Duration
}

func nodeSelector(options *metav1.ListOptions, opt WatchOptions) {
	if len(opt.Node) != 0 {
		options.FieldSelector = fmt.Sprintf("spec.nodeName=%s", opt.Node)
	}
}
func nameSelector(options *metav1.ListOptions, opt WatchOptions) {
	if len(opt.Node) != 0 {
		options.FieldSelector = fmt.Sprintf("metadata.name=%s", opt.Node)
	}
}
func NewInformer(client kubernetes.Interface, resource Resource, opts WatchOptions, indexers cache.Indexers) cache.SharedIndexInformer {
	ctx := context.TODO()
	var listWatch *cache.ListWatch
	switch resource.(type) {
	case *coreV1.Pod:
		pod := client.CoreV1().Pods(opts.Namespace)
		listWatch = &cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				nodeSelector(&options, opts)
				return pod.List(ctx, options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				nodeSelector(&options, opts)
				return pod.Watch(ctx, options)
			},
		}
	case *coreV1.Node:
		node := client.CoreV1().Nodes()
		listWatch = &cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				nameSelector(&options, opts)
				return node.List(ctx, options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				nameSelector(&options, opts)
				return node.Watch(ctx, options)
			},
		}
	}

	return cache.NewSharedIndexInformer(listWatch, resource, opts.SyncTimeout, indexers)
}

func NewController(informerFactory informers.SharedInformerFactory, client *kubernetes.Clientset, collectors metrics.Collectors) (*Controller, error) {
	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), Name)
	controller := &Controller{
		queue:           queue,
		informerFactory: informerFactory,
		client:          client,
	}

	podInformer := NewInformer(client, &coreV1.Pod{}, WatchOptions{
		Namespace:   viper.GetString("namespace"),
		Node:        DiscoverKubernetesNode(viper.GetString("node"), client),
		SyncTimeout: viper.GetDuration("sync-period"),
	}, cache.Indexers{})
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.AddPod,
		UpdateFunc: controller.UpdatePod,
		DeleteFunc: nil,
	})
	controller.podInformer = podInformer

	nodeInformer := NewInformer(client, &coreV1.Node{}, WatchOptions{
		Node:        DiscoverKubernetesNode(viper.GetString("node"), client),
		SyncTimeout: viper.GetDuration("sync-period"),
	}, cache.Indexers{})

	controller.nodeInformer = nodeInformer

	return controller, nil
}

type item struct {
	object interface{}
	key    string
}

func (controller *Controller) UpdatePod(oldObj, newObj interface{}) {
	o, _ := accessor.ResourceVersion(oldObj.(runtime.Object))
	n, _ := accessor.ResourceVersion(newObj.(runtime.Object))
	// 只有resource version不同才是新对象
	if o != n {
		controller.Enqueue(&item{
			object: n,
		})
	}
}

func (controller *Controller) AddPod(obj interface{}) {
	controller.Enqueue(&item{
		object: obj,
	})
}

func (controller *Controller) Enqueue(item *item) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(item.object)
	if err != nil {
		return
	}

	item.key = key
	controller.queue.Add(item)
}

func (controller *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	go controller.podInformer.Run(stopCh)
	go controller.nodeInformer.Run(stopCh)

	if !cache.WaitForNamedCacheSync(Name, stopCh,
		controller.podInformer.HasSynced,
		controller.nodeInformer.HasSynced) {
		return fmt.Errorf("kubernetes informer is unable to sync cache")
	}

	for i := 0; i < threadiness; i++ {
		// Wrap the process function with wait.Until so that if the controller crashes, it starts up again after a second.
		go wait.Until(func() {
			for controller.process() {
			}
		}, time.Second*1, stopCh)
	}

	return nil
}

func (controller *Controller) process() bool {
	keyObj, quit := controller.queue.Get()
	if quit {
		return false
	}

	err := func(obj interface{}) error {
		defer controller.queue.Done(obj)

		var entry *item
		var ok bool
		if entry, ok = obj.(*item); !ok {
			controller.queue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected *item in workqueue but got %#v", obj))
			return nil
		}

		obj, exists, err := controller.podInformer.GetStore().GetByKey(entry.key)
		if err != nil {
			return err
		}
		if !exists {
			log.Infof("object %+v was not found in the store", entry.key)
			return nil
		}

		var pod *coreV1.Pod
		if pod, ok = obj.(*coreV1.Pod); !ok {
			controller.queue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected *coreV1.Pod but got %#v", obj))
			return nil
		}

		log.Infof("updating cache with pod %s/%s", pod.Namespace, pod.Name)

		return nil
	}(keyObj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

/*func New(options common.Options) *LogController {

	client, err := common.GetKubernetesClient(options.KubeConfig)
	if err != nil {

	}

	podWatcher, err := k8s.NewWatcher(client, &k8s.Pod{}, k8s.WatchOptions{
		SyncTimeout: time.Minute * 10,
		Node:        DiscoverKubernetesNode(options.Host, client),
		Namespace:   "",
		IsUpdated:   nil,
	}, nil)

	ctr := &LogController{
		InformerResources: []schema.GroupVersionResource{
			{
				Group:    coreV1.GroupName,
				Version:  coreV1.SchemeGroupVersion.Version,
				Resource: "pods",
			},
			{
				Group: appsv1.GroupName,
				Version:  coreV1.SchemeGroupVersion.Version,
				Resource: "nodes",
			},
		},
		ApiServerClient:   apiServerClient,
		stopCh:            nil,
	}

	ctr := &LogController{
		InformerResources: nil,
		ApiServerClient:   nil,
		stopCh:            nil,
		NodeName:          "",
		ResyncPeriod:      0,
		TaskHandlePeriod:  0,
	}

	ctr.TaskQueue = NewTaskQueue(ctr.syncTask)

	podListWatch := cache.NewListWatchFromClient(ctr.ApiServerClient.CoreV1().RESTClient(), "pods", coreV1.NamespaceAll, fields.OneTermEqualSelector("spec.nodeName", ctr.NodeName))
	ctr.PodStore, ctr.PodInformer = cache.NewInformer(podListWatch, &coreV1.Pod{}, ctr.ResyncPeriod, cache.ResourceEventHandlerFuncs{
		AddFunc:    ctr.AddPod,
		UpdateFunc: ctr.UpdatePod,
		DeleteFunc: ctr.DeletePod,
	})

	nodeListWatch := cache.NewListWatchFromClient(ctr.ApiServerClient.CoreV1().RESTClient(), "nodes", coreV1.NamespaceAll, fields.OneTermEqualSelector("metadata.name", ctr.NodeName))
	ctr.NodeStore, ctr.NodeInformer = cache.NewInformer(nodeListWatch, &coreV1.Node{}, ctr.ResyncPeriod, cache.ResourceEventHandlerFuncs{})

}

func (ctr *LogController) AddPod(obj interface{}) {
	pod := obj.(*coreV1.Pod)
	ctr.TaskQueue.Enqueue(pod)
}
func (ctr *LogController) UpdatePod(oldObj, newObj interface{}) {}
func (ctr *LogController) DeletePod(obj interface{}) {
	pod := obj.(*coreV1.Pod)
	ctr.TaskQueue.Enqueue(pod)
}

// 批量处理pod事件，更新filebeat input.yml
func (ctr *LogController) syncTask(tasks []interface{}) {

}

func (ctr *LogController) Run() {

	go ctr.PodInformer.Run(ctr.stopCh)
	go ctr.NodeInformer.Run(ctr.stopCh)

	go ctr.TaskQueue.Run(ctr.TaskHandlePeriod, ctr.stopCh)

	<-ctr.stopCh
	//k8s.Run(ctr.ApiServerClient, ctr.InformerResources, ctr.stopCh)
}*/
