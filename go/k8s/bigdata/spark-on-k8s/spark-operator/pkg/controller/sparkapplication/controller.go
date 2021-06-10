package sparkapplication

import (
	"fmt"
	"github.com/golang/glog"
	"k8s-lx1036/k8s/bigdata/spark-on-k8s/spark-operator/cmd/app/options"
	v1 "k8s-lx1036/k8s/bigdata/spark-on-k8s/spark-operator/pkg/apis/sparkoperator.k9s.io/v1"
	"k8s-lx1036/k8s/bigdata/spark-on-k8s/spark-operator/pkg/client/clientset/versioned"
	sparkApplicationInformer "k8s-lx1036/k8s/bigdata/spark-on-k8s/spark-operator/pkg/client/informers/externalversions"
	sparkApplicationLister "k8s-lx1036/k8s/bigdata/spark-on-k8s/spark-operator/pkg/client/listers/sparkoperator.k9s.io/v1"
	"k8s-lx1036/k8s/bigdata/spark-on-k8s/spark-operator/pkg/config"
	"k8s-lx1036/k8s/bigdata/spark-on-k8s/spark-operator/pkg/utils"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

var (
	keyFunc = cache.DeletionHandlingMetaNamespaceKeyFunc
)

// Controller manages instances of SparkApplication.
type Controller struct {
	crdClient  versioned.Interface
	kubeClient kubernetes.Interface

	podInformer cache.SharedIndexInformer

	sparkAppInformer       cache.SharedIndexInformer
	sparkApplicationLister sparkApplicationLister.SparkApplicationLister
	queue                  workqueue.RateLimitingInterface

	recorder record.EventRecorder
}

func NewController(option options.Options) *Controller {
	restConfig, err := utils.NewRestConfig(option.Kubeconfig)
	if err != nil {
		return nil, err
	}

	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to construct lister client: %v", err)
	}

	var sparkAppFactoryOpts []sparkApplicationInformer.SharedInformerOption
	if option.Namespace != apiv1.NamespaceAll {
		sparkAppFactoryOpts = append(sparkAppFactoryOpts, sparkApplicationInformer.WithNamespace(option.Namespace))
	}
	sparkAppClient := versioned.NewForConfigOrDie(restConfig)
	sparkAppInformerFactory := sparkApplicationInformer.NewSharedInformerFactoryWithOptions(sparkAppClient, time.Second*30, sparkAppFactoryOpts...)
	sparkAppInformer := sparkAppInformerFactory.Sparkoperator().V1().SparkApplications().Informer()

	var podFactoryOpts []informers.SharedInformerOption
	if option.Namespace != apiv1.NamespaceAll {
		podFactoryOpts = append(podFactoryOpts, informers.WithNamespace(option.Namespace))
	}
	tweakListOptionsFunc := func(options *metav1.ListOptions) {
		// INFO: 过滤带有"spark-role,sparkoperator.k8s.io/launched-by-spark-operator" label 的 pods
		// kubectl get pods -l="spark-role,sparkoperator.k8s.io/launched-by-spark-operator" -o wide
		options.LabelSelector = fmt.Sprintf("%s,%s", config.SparkRoleLabel, config.LaunchedBySparkOperatorLabel)
	}
	podFactoryOpts = append(podFactoryOpts, informers.WithTweakListOptions(tweakListOptionsFunc))
	podInformerFactory := informers.NewSharedInformerFactoryWithOptions(kubeClient, time.Second*30, podFactoryOpts...)
	podInformer := podInformerFactory.Core().V1().Pods().Informer()

	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "spark-application-controller")

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{
		Interface: kubeClient.CoreV1().Events(option.Namespace),
	})
	eventRecorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{
		Component: "spark-operator",
	})
	controller := &Controller{
		crdClient:              nil,
		kubeClient:             nil,
		podInformer:            podInformer,
		sparkAppInformer:       sparkAppInformer,
		queue:                  queue,
		recorder:               eventRecorder,
		sparkApplicationLister: sparkAppInformerFactory.Sparkoperator().V1().SparkApplications().Lister(),
	}

	sparkAppInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.onAdd,
		UpdateFunc: controller.onUpdate,
		DeleteFunc: controller.onDelete,
	})

	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.onPodAdd,
		UpdateFunc: controller.onPodUpdate,
		DeleteFunc: controller.onPodDelete,
	})

	return controller
}

func (controller *Controller) Start(workers int, stopCh <-chan struct{}) error {
	go controller.podInformer.Run(stopCh)
	go controller.sparkAppInformer.Run(stopCh)

	shutdown := cache.WaitForCacheSync(stopCh, controller.podInformer.HasSynced, controller.sparkAppInformer.HasSynced)
	if !shutdown {
		klog.Errorf("can not sync pods in node %s", server.option.Nodename)
		return nil
	}

	klog.Info("Starting the workers of the SparkApplication controller")
	for i := 0; i < workers; i++ {
		// runWorker will loop until "something bad" happens. Until will then rekick
		// the worker after one second.
		go wait.Until(controller.runWorker, time.Second, stopCh)
	}

	return nil
}

// runWorker runs a single controller worker.
func (controller *Controller) runWorker() {
	defer utilruntime.HandleCrash()
	for controller.processNextItem() {
	}
}

func (controller *Controller) processNextItem() bool {
	key, quit := controller.queue.Get()

	if quit {
		return false
	}
	defer controller.queue.Done(key)

	glog.V(2).Infof("Starting processing key: %q", key)
	defer glog.V(2).Infof("Ending processing key: %q", key)
	err := controller.syncSparkApplication(key.(string))
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

func (controller *Controller) onAdd(obj interface{}) {
	app := obj.(*v1.SparkApplication)
	klog.Infof("SparkApplication %s/%s was added, enqueuing it for submission", app.Namespace, app.Name)
	controller.enqueue(app)
}

func (controller *Controller) onUpdate(oldObj, newObj interface{}) {
	oldApp := oldObj.(*v1.SparkApplication)
	newApp := newObj.(*v1.SparkApplication)

	// The informer will call this function on non-updated resources during resync, avoid
	// enqueuing unchanged applications, unless it has expired or is subject to retry.
	if oldApp.ResourceVersion == newApp.ResourceVersion && !controller.hasApplicationExpired(newApp) && !shouldRetry(newApp) {
		return
	}

	// INFO: 如果 sparkApplication spec 发生了变化，比如 apply 了一个 test1 SparkApplication，然后修改其 spec 内容但是name没变重新 apply
	// 则去
	if !equality.Semantic.DeepEqual(oldApp.Spec, newApp.Spec) {
		updatedApp := newApp.DeepCopy()
		updatedApp.Status.AppState.State = v1.InvalidatingState
		err := controller.updateApplicationStatusWithRetries(newApp, updatedApp)
		if err != nil {
			controller.recorder.Eventf(
				newApp,
				apiv1.EventTypeWarning,
				"SparkApplicationSpecUpdateFailed",
				"failed to process spec update for SparkApplication %s: %v",
				newApp.Name,
				err)
			return
		}

		// INFO: 生成出一个新 event 对象，而且是 SparkApplication 对象的 event，在 kubectl describe SparkApplication 时可以看到相关事件，非常便于 debug
		controller.recorder.Eventf(
			newApp,
			apiv1.EventTypeNormal,
			"SparkApplicationSpecUpdateProcessed",
			"Successfully processed spec update for SparkApplication %s",
			newApp.Name)
	}

	klog.V(2).Infof("SparkApplication %s/%s was updated, enqueuing it", newApp.Namespace, newApp.Name)
	controller.enqueue(newApp)
}

// INFO: 更新 SparkApplication status 子对象，如果失败，则尝试几次
func (c *Controller) updateApplicationStatusWithRetries() error {

}

func (c *Controller) enqueue(obj interface{}) {
	key, err := keyFunc(obj)
	if err != nil {
		glog.Errorf("failed to get key for %v: %v", obj, err)
		return
	}

	// INFO: AddRateLimited() 比 Add() 更好在于，AddRateLimited() 有限速器，会在 RateLimiter ok 之后才会 Add()，以后用 AddRateLimited()
	c.queue.AddRateLimited(key)
}

func (c *Controller) getSparkApplication(namespace string, name string) (*v1.SparkApplication, error) {
	app, err := c.sparkApplicationLister.SparkApplications(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return app, nil
}

// State Machine for SparkApplication:
//+--------------------------------------------------------------------------------------------------------------------+
//|        +---------------------------------------------------------------------------------------------+             |
//|        |       +----------+                                                                          |             |
//|        |       |          |                                                                          |             |
//|        |       |          |                                                                          |             |
//|        |       |Submission|                                                                          |             |
//|        |  +---->  Failed  +----+------------------------------------------------------------------+  |             |
//|        |  |    |          |    |                                                                  |  |             |
//|        |  |    |          |    |                                                                  |  |             |
//|        |  |    +----^-----+    |  +-----------------------------------------+                     |  |             |
//|        |  |         |          |  |                                         |                     |  |             |
//|        |  |         |          |  |                                         |                     |  |             |
//|      +-+--+----+    |    +-----v--+-+          +----------+           +-----v-----+          +----v--v--+          |
//|      |         |    |    |          |          |          |           |           |          |          |          |
//|      |         |    |    |          |          |          |           |           |          |          |          |
//|      |   New   +---------> Submitted+----------> Running  +----------->  Failing  +---------->  Failed  |          |
//|      |         |    |    |          |          |          |           |           |          |          |          |
//|      |         |    |    |          |          |          |           |           |          |          |          |
//|      |         |    |    |          |          |          |           |           |          |          |          |
//|      +---------+    |    +----^-----+          +-----+----+           +-----+-----+          +----------+          |
//|                     |         |                      |                      |                                      |
//|                     |         |                      |                      |                                      |
//|    +------------+   |         |             +-------------------------------+                                      |
//|    |            |   |   +-----+-----+       |        |                +-----------+          +----------+          |
//|    |            |   |   |  Pending  |       |        |                |           |          |          |          |
//|    |            |   +---+   Rerun   <-------+        +---------------->Succeeding +---------->Completed |          |
//|    |Invalidating|       |           <-------+                         |           |          |          |          |
//|    |            +------->           |       |                         |           |          |          |          |
//|    |            |       |           |       |                         |           |          |          |          |
//|    |            |       +-----------+       |                         +-----+-----+          +----------+          |
//|    +------------+                           |                               |                                      |
//|                                             |                               |                                      |
//|                                             +-------------------------------+                                      |
//|                                                                                                                    |
//+--------------------------------------------------------------------------------------------------------------------+
func (c *Controller) syncSparkApplication(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return fmt.Errorf("failed to get the namespace and name from key %s: %v", key, err)
	}
	app, err := c.getSparkApplication(namespace, name)
	if err != nil {
		return err
	}
	if app == nil {
		// INFO: SparkApplication not found, 不用管
		return nil
	}

	if !app.DeletionTimestamp.IsZero() {
		c.handleSparkApplicationDeletion(app) // 删除了 SparkApplication
		return nil
	}

	appCopy := app.DeepCopy()
	// Apply the default values to the copy. Note that the default values applied
	// won't be sent to the API server as we only update the /status subresource.
	v1.SetSparkApplicationDefaults(appCopy)

	// Take action based on application state.
	switch appCopy.Status.AppState.State {
	case v1.NewState:

	case v1.SucceedingState:

	case v1.FailingState:

	case v1.FailedSubmissionState:

	case v1.InvalidatingState:

	case v1.PendingRerunState:

	case v1.SubmittedState, v1.RunningState, v1.UnknownState:

	case v1.CompletedState, v1.FailedState:

	}

	if appCopy != nil {

	}

	return nil
}
