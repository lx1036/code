package controller

import (
	"fmt"
	"time"

	"k8s-lx1036/k8s/storage/etcd/etcd-operator/cmd/cluster/app/options"
	v1 "k8s-lx1036/k8s/storage/etcd/etcd-operator/pkg/apis/etcd.k9s.io/v1"
	"k8s-lx1036/k8s/storage/etcd/etcd-operator/pkg/client/clientset/versioned"
	"k8s-lx1036/k8s/storage/etcd/etcd-operator/pkg/client/informers/externalversions"
	etcdClusterLister "k8s-lx1036/k8s/storage/etcd/etcd-operator/pkg/client/listers/etcd.k9s.io/v1"
	"k8s-lx1036/k8s/storage/etcd/etcd-operator/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

const (
	Name = "etcd-cluster-controller"
)

var (
	keyFunc = cache.DeletionHandlingMetaNamespaceKeyFunc
)

type EtcdClusterController struct {
	kubeClient kubernetes.Interface

	etcdClusterInformer cache.SharedIndexInformer

	queue workqueue.RateLimitingInterface

	recorder record.EventRecorder

	etcdClusterLister etcdClusterLister.EtcdClusterLister
	etcdClusterClient *versioned.Clientset

	// INFO: 一个 Cluster 表示一个 Etcd Cluster, clusters 表示所有用户的 etcd cluster
	clusters map[string]*Cluster
}

func NewController(option *options.Options) (*EtcdClusterController, error) {
	restConfig, err := utils.NewRestConfig(option.Kubeconfig)
	if err != nil {
		return nil, err
	}

	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to construct lister client: %v", err)
	}

	var etcdClusterFactoryOpts []externalversions.SharedInformerOption
	if option.Namespace != corev1.NamespaceAll {
		etcdClusterFactoryOpts = append(etcdClusterFactoryOpts, externalversions.WithNamespace(option.Namespace))
	}

	etcdClusterClient := versioned.NewForConfigOrDie(restConfig)
	etcdClusterInformerFactory := externalversions.NewSharedInformerFactoryWithOptions(etcdClusterClient, time.Second*30, etcdClusterFactoryOpts...)
	etcdClusterInformer := etcdClusterInformerFactory.Etcd().V1().EtcdClusters().Informer()

	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), Name)

	// INFO: 或者如果注册到 k8s-lx1036/k8s/bigdata/spark-on-k8s/spark-operator/pkg/client/clientset/versioned/scheme::Scheme，就不需要在这里
	//  注册 _ = v1.AddToScheme(scheme.Scheme)，这个逻辑可以参考 volcano queue controller 里的 eventBroadcaster 对象实例化

	// INFO: 这里由于eventBroadcaster.NewRecorder使用的是根scheme.Scheme，所以必须要把sparkoperator.k9s.io/v1注册到这个Scheme里，
	//  否则会在这里报错：https://github.com/kubernetes/kubernetes/blob/v1.19.7/staging/src/k8s.io/client-go/tools/reference/ref.go#L66-L68
	//  说没有注册报错 NewNotRegisteredErrForType(s.schemeName, t)
	_ = v1.AddToScheme(scheme.Scheme)
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{
		Interface: kubeClient.CoreV1().Events(option.Namespace),
	})
	eventRecorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{
		Component: "etcd-cluster",
	})

	controller := &EtcdClusterController{
		etcdClusterInformer: etcdClusterInformer,
		kubeClient:          kubeClient,
		queue:               queue,
		recorder:            eventRecorder,
		etcdClusterClient:   etcdClusterClient,
		etcdClusterLister:   etcdClusterInformerFactory.Etcd().V1().EtcdClusters().Lister(),
		clusters:            make(map[string]*Cluster),
	}

	etcdClusterInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.onEtcdClusterAdd,
		UpdateFunc: controller.onEtcdClusterUpdate,
		DeleteFunc: controller.onEtcdClusterDelete,
	})

	return controller, nil
}

func (controller *EtcdClusterController) onEtcdClusterAdd(obj interface{}) {
	etcdCluster := obj.(*v1.EtcdCluster)
	klog.Infof("EtcdCluster %s/%s was added, enqueuing it for submission", etcdCluster.Namespace, etcdCluster.Name)
	controller.enqueue(etcdCluster)
}

func (controller *EtcdClusterController) onEtcdClusterUpdate(oldObj, newObj interface{}) {
	oldEtcdCluster := oldObj.(*v1.EtcdCluster)
	newEtcdCluster := newObj.(*v1.EtcdCluster)

	// The informer will call this function on non-updated resources during resync, avoid
	// enqueuing unchanged applications, unless it has expired or is subject to retry.
	if oldEtcdCluster.ResourceVersion == newEtcdCluster.ResourceVersion {
		//if oldApp.ResourceVersion == newApp.ResourceVersion && !controller.hasApplicationExpired(newApp) && !shouldRetry(newApp) {
		return
	}

	// INFO: status 被更新了也会触发 update 事件，我们只关注 Spec 的更新
	//  Spec struct 的字段都是标量(如int,string,指针)等等，是可以直接使用操作符 == ，如果有字段类型如 map，则不可以 ==
	if oldEtcdCluster.Spec == newEtcdCluster.Spec {
		return
	}

	klog.Infof("EtcdCluster %s/%s was updated, enqueuing it", newEtcdCluster.Namespace, newEtcdCluster.Name)
	controller.enqueue(newEtcdCluster)
}

func (controller *EtcdClusterController) onEtcdClusterDelete(obj interface{}) {
	etcdCluster, ok := obj.(*v1.EtcdCluster)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			klog.Errorf("Couldn't get object from tombstone %#v", obj)
			return
		}
		etcdCluster, ok = tombstone.Obj.(*v1.EtcdCluster)
		if !ok {
			klog.Errorf("Tombstone contained object that is not expected %#v", obj)
			return
		}
	}

	if etcdCluster != nil {
		controller.handleEtcdClusterDeletion(etcdCluster) // 删除了 etcdCluster
		controller.recorder.Eventf(etcdCluster, corev1.EventTypeNormal, "EtcdClusterDeleted", "EtcdCluster %s was deleted", etcdCluster.Name)
		klog.Infof("EtcdCluster %s/%s was deleted", etcdCluster.Namespace, etcdCluster.Name)
	}
}

func (controller *EtcdClusterController) Start(workers int, stopCh <-chan struct{}) error {
	go controller.etcdClusterInformer.Run(stopCh)

	shutdown := cache.WaitForCacheSync(stopCh, controller.etcdClusterInformer.HasSynced)
	if !shutdown {
		klog.Errorf("can not sync sparkApplication and pods in ")
		return nil
	}

	klog.Info("Starting the workers of the EtcdClusterController")
	for i := 0; i < workers; i++ {
		// runWorker will loop until "something bad" happens. Until will then rekick
		// the worker after one second.
		go wait.Until(controller.runWorker, time.Second, stopCh)
	}

	return nil
}

func (controller *EtcdClusterController) handleEtcdClusterDeletion(etcdCluster *v1.EtcdCluster) {

}

func (controller *EtcdClusterController) enqueue(obj interface{}) {
	key, err := keyFunc(obj)
	if err != nil {
		klog.Errorf("failed to get key for %v: %v", obj, err)
		return
	}

	// INFO: AddRateLimited() 比 Add() 更好在于，AddRateLimited() 有限速器，会在 RateLimiter ok 之后才会 Add()，以后用 AddRateLimited()
	controller.queue.AddRateLimited(key)
}

// runWorker runs a single controller worker.
func (controller *EtcdClusterController) runWorker() {
	defer utilruntime.HandleCrash()
	for controller.processNextItem() {
	}
}

func (controller *EtcdClusterController) processNextItem() bool {
	key, quit := controller.queue.Get()
	if quit {
		return false
	}
	defer controller.queue.Done(key)

	klog.V(2).Infof("Starting processing key: %q", key)
	defer klog.V(2).Infof("Ending processing key: %q", key)
	err := controller.syncEtcdCluster(key.(string))
	if err == nil {
		// Successfully processed the key or the key was not found so tell the queue to stop tracking
		// history for your key. This will reset things like failure counts for per-item rate limiting.
		controller.queue.Forget(key)
		return true
	}

	// There was a failure so be sure to report it. This method allows for pluggable error handling
	// which can be used for things like cluster-monitoring
	utilruntime.HandleError(fmt.Errorf("failed to sync SparkApplication %q: %v", key, err))
	return true
}

func (controller *EtcdClusterController) syncEtcdCluster(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return fmt.Errorf("[syncEtcdCluster]failed to get the namespace and name from key %s: %v", key, err)
	}

	etcdCluster, err := controller.etcdClusterLister.EtcdClusters(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	// INFO: 通过 DeletionTimestamp 判断是否已经删除。Update 事件时可能会出现。可以复用!!!
	if !etcdCluster.DeletionTimestamp.IsZero() {
		controller.handleEtcdClusterDeletion(etcdCluster) // 删除了 EtcdCluster
		return nil
	}

	event := &watch.Event{
		Type:   watch.Added,
		Object: etcdCluster,
	}
	// re-watch or restart could give ADD event.
	// If for an ADD event the cluster spec is invalid then it is not added to the local cache
	// so modifying that cluster will result in another ADD event
	if _, ok := controller.clusters[key]; ok {
		klog.Infof(fmt.Sprintf("[syncEtcdCluster]cluster %s is already existed in clusters, update event type Modified", key))
		event.Type = watch.Modified
	}

	switch event.Type {
	case watch.Added:
		if _, ok := controller.clusters[key]; ok {
			return fmt.Errorf("[syncEtcdCluster]unsafe state. cluster (%s) was created before but we received event (%s)", key, event.Type)
		}

		cluster := NewCluster(&ClusterConfig{
			kubeClient:        controller.kubeClient,
			etcdClusterClient: controller.etcdClusterClient,
		}, etcdCluster)
		controller.clusters[key] = cluster

	case watch.Modified:

	case watch.Deleted:

	}

	return nil
}

func (controller *EtcdClusterController) Stop() {
	klog.Info("Stopping the EtcdClusterController")
	controller.queue.ShutDown()
}
