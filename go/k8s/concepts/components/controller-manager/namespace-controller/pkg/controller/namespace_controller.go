package controller

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/metadata"
	"time"

	"k8s-lx1036/k8s/concepts/components/controller-manager/namespace-controller/pkg/kube"

	log "github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	corev1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type NamespaceController struct {
	queue                      workqueue.RateLimitingInterface
	namespacedResourcesDeleter NamespacedResourcesDeleterInterface

	namespaceLister       corev1.NamespaceLister
	namespaceListerSynced cache.InformerSynced
}

func NewNamespaceController() *NamespaceController {
	clientset := kube.GetClientset()
	informerFactory := informers.NewSharedInformerFactory(clientset, time.Minute*10)
	namespaceInformer := informerFactory.Core().V1().Namespaces()
	discoverResourcesFn := clientset.Discovery().ServerPreferredNamespacedResources
	metadataClient, err := metadata.NewForConfig(kube.GetConfig())
	if err != nil {
		panic(err)
	}

	namespaceController := &NamespaceController{
		queue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "namespace"),
		namespacedResourcesDeleter: NewNamespacedResourcesDeleter(discoverResourcesFn,
			clientset.CoreV1().Namespaces(),
			v1.FinalizerKubernetes,
			metadataClient,
		),
		namespaceLister:       namespaceInformer.Lister(),
		namespaceListerSynced: namespaceInformer.Informer().HasSynced,
	}

	return namespaceController
}

func (controller *NamespaceController) Run(workers int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer controller.queue.ShutDown()

	log.Infof("Starting namespace controller")
	defer log.Infof("Shutting down namespace controller")

	if !cache.WaitForNamedCacheSync("namespace", stopCh, controller.namespaceListerSynced) {
		return fmt.Errorf("kubernetes informer is unable to sync cache")
	}

	log.Info("Starting workers of namespace controller")
	for i := 0; i < workers; i++ {
		go wait.Until(func() {
			for controller.process() {
			}
		}, time.Second, stopCh)
	}

	return nil
}

func (controller *NamespaceController) process() bool {
	key, quit := controller.queue.Get()
	if quit {
		return false
	}
	defer controller.queue.Done(key)

	err := controller.syncNamespace(key.(string))
	if err == nil {
		// no error, forget this entry and return
		controller.queue.Forget(key)
		return true
	}

	if estimate, ok := err.(*ResourcesRemainingError); ok {
		t := estimate.Estimate/2 + 1
		log.Infof("Content remaining in namespace %s, waiting %d seconds", key, t)
		controller.queue.AddAfter(key, time.Duration(t)*time.Second)
	} else {
		// rather than wait for a full resync, re-add the namespace to the queue to be processed
		controller.queue.AddRateLimited(key)
		utilruntime.HandleError(fmt.Errorf("deletion of namespace %v failed: %v", key, err))
	}

	return true
}

func (controller *NamespaceController) syncNamespace(key string) error {
	// 记录延迟
	startTime := time.Now()
	defer func() {
		log.Infof("Finished syncing namespace %q (%v)", key, time.Since(startTime))
	}()

	// 从本地缓存根据key取出namespace对象，因为namespace对象可能在周期性同步中已经被删除了，就不需要继续执行后续业务逻辑
	namespace, err := controller.namespaceLister.Get(key)
	if errors.IsNotFound(err) {
		log.Infof("Namespace has been deleted %v", key)
		return nil
	}
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("unable to retrieve namespace %v from store: %v", key, err))
		return err
	}

	return controller.namespacedResourcesDeleter.Delete(namespace.Name)
}
