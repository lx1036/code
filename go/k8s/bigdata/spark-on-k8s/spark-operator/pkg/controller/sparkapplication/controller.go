package sparkapplication

import (
	"context"
	"fmt"
	"time"

	"k8s-lx1036/k8s/bigdata/spark-on-k8s/spark-operator/cmd/app/options"
	v1 "k8s-lx1036/k8s/bigdata/spark-on-k8s/spark-operator/pkg/apis/sparkoperator.k9s.io/v1"
	"k8s-lx1036/k8s/bigdata/spark-on-k8s/spark-operator/pkg/batchscheduler"
	"k8s-lx1036/k8s/bigdata/spark-on-k8s/spark-operator/pkg/batchscheduler/schedulerinterface"
	"k8s-lx1036/k8s/bigdata/spark-on-k8s/spark-operator/pkg/client/clientset/versioned"
	sparkApplicationInformer "k8s-lx1036/k8s/bigdata/spark-on-k8s/spark-operator/pkg/client/informers/externalversions"
	sparkApplicationLister "k8s-lx1036/k8s/bigdata/spark-on-k8s/spark-operator/pkg/client/listers/sparkoperator.k9s.io/v1"
	"k8s-lx1036/k8s/bigdata/spark-on-k8s/spark-operator/pkg/config"
	"k8s-lx1036/k8s/bigdata/spark-on-k8s/spark-operator/pkg/utils"

	"github.com/google/uuid"
	apiv1 "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

const (
	Name = "spark-application-controller"
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

	batchSchedulerMgr *batchscheduler.SchedulerManager

	metrics *sparkAppMetrics
}

func NewController(option *options.Options) (*Controller, error) {
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

	// INFO: 只会 watch driver pod
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

	batchSchedulerMgr := batchscheduler.NewSchedulerManager(restConfig)

	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), Name)

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
		batchSchedulerMgr:      batchSchedulerMgr,
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

	return controller, nil
}

func (controller *Controller) Start(workers int, stopCh <-chan struct{}) error {
	go controller.podInformer.Run(stopCh)
	go controller.sparkAppInformer.Run(stopCh)

	shutdown := cache.WaitForCacheSync(stopCh, controller.podInformer.HasSynced, controller.sparkAppInformer.HasSynced)
	if !shutdown {
		klog.Errorf("can not sync sparkApplication and pods in ")
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

func (controller *Controller) Stop() {
	klog.Info("Stopping the SparkApplication controller")
	controller.queue.ShutDown()
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

	klog.V(2).Infof("Starting processing key: %q", key)
	defer klog.V(2).Infof("Ending processing key: %q", key)
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
	if oldApp.ResourceVersion == newApp.ResourceVersion {
		//if oldApp.ResourceVersion == newApp.ResourceVersion && !controller.hasApplicationExpired(newApp) && !shouldRetry(newApp) {
		return
	}

	// INFO: 如果 sparkApplication spec 发生了变化，比如 apply 了一个 test1 SparkApplication，然后修改其 spec 内容但是name没变重新 apply
	// 则需要 InvalidatingState
	if !equality.Semantic.DeepEqual(oldApp.Spec, newApp.Spec) {
		_, err := controller.updateApplicationStatusWithRetries(newApp, func(status *v1.SparkApplicationStatus) {
			status.AppState.State = v1.InvalidatingState
		})
		if err != nil {
			controller.recorder.Eventf(newApp, apiv1.EventTypeWarning, "SparkApplicationSpecUpdateFailed", "failed to process spec update for SparkApplication %s: %v", newApp.Name, err)
			return
		}

		// INFO: 生成出一个新 event 对象，而且是 SparkApplication 对象的 event，在 kubectl describe SparkApplication 时可以看到相关事件，非常便于 debug
		controller.recorder.Eventf(newApp, apiv1.EventTypeNormal, "SparkApplicationSpecUpdateProcessed", "Successfully processed spec update for SparkApplication %s", newApp.Name)
	}

	klog.V(2).Infof("SparkApplication %s/%s was updated, enqueuing it", newApp.Namespace, newApp.Name)
	controller.enqueue(newApp)
}

func (controller *Controller) onDelete(obj interface{}) {
	sparkApplication, ok := obj.(*v1.SparkApplication)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			klog.Errorf("Couldn't get object from tombstone %#v", obj)
			return
		}
		sparkApplication, ok = tombstone.Obj.(*v1.SparkApplication)
		if !ok {
			klog.Errorf("Tombstone contained object that is not expected %#v", obj)
			return
		}
	}

	if sparkApplication != nil {
		controller.handleSparkApplicationDeletion(sparkApplication) // 删除了 SparkApplication
		controller.recorder.Eventf(sparkApplication, apiv1.EventTypeNormal, "SparkApplicationDeleted", "SparkApplication %s was deleted", sparkApplication.Name)
		klog.V(2).Infof("SparkApplication %s/%s was deleted", sparkApplication.Namespace, sparkApplication.Name)
	}
}

func (controller *Controller) handleSparkApplicationDeletion(app *v1.SparkApplication) {
	// SparkApplication deletion requested, lets delete driver pod.
	if err := controller.deleteSparkResources(app); err != nil {
		klog.Errorf("failed to delete resources associated with deleted SparkApplication %s/%s: %v", app.Namespace, app.Name, err)
	}
}

// Delete the driver pod and optional UI resources (Service/Ingress) created for the application.
func (controller *Controller) deleteSparkResources(app *v1.SparkApplication) error {

	return nil
}

func (controller *Controller) onPodAdd(obj interface{}) {

}

func (controller *Controller) onPodUpdate(oldObj, newObj interface{}) {

}

func (controller *Controller) onPodDelete(obj interface{}) {

}

// INFO: 更新 SparkApplication status 子对象，如果失败，则尝试4次。这个函数逻辑可以直接复用!!!
func (controller *Controller) updateApplicationStatusWithRetries(original *v1.SparkApplication,
	updateFunc func(status *v1.SparkApplicationStatus)) (*v1.SparkApplication, error) {
	toUpdate := original.DeepCopy()
	updateErr := wait.ExponentialBackoff(retry.DefaultBackoff, func() (ok bool, err error) {
		updateFunc(&toUpdate.Status) // 更新 status 字段值
		if equality.Semantic.DeepEqual(original.Status, toUpdate.Status) {
			return true, nil
		}
		// 开始更新 SparkApplication status 子对象
		toUpdate, err = controller.crdClient.SparkoperatorV1().SparkApplications(original.Namespace).UpdateStatus(context.TODO(), toUpdate, metav1.UpdateOptions{})
		if err == nil {
			return true, nil
		}
		if !errors.IsConflict(err) {
			return false, err
		}

		// INFO: 更新时发生 conflict 错误，这是因为不是 latest resource version，所以需要重新 fetch 下
		toUpdate, err = controller.crdClient.SparkoperatorV1().SparkApplications(original.Namespace).Get(context.TODO(), original.Name, metav1.GetOptions{})
		if err != nil {

			return false, err
		}

		// Retry with the latest version. 使用最新的 toUpdate 继续重试
		return false, nil
	})

	if updateErr != nil {
		klog.Errorf("failed to update SparkApplication %s/%s: %v", original.Namespace, original.Name, updateErr)
		return nil, updateErr
	}

	return toUpdate, nil
}

func (controller *Controller) enqueue(obj interface{}) {
	key, err := keyFunc(obj)
	if err != nil {
		klog.Errorf("failed to get key for %v: %v", obj, err)
		return
	}

	// INFO: AddRateLimited() 比 Add() 更好在于，AddRateLimited() 有限速器，会在 RateLimiter ok 之后才会 Add()，以后用 AddRateLimited()
	controller.queue.AddRateLimited(key)
}

func (controller *Controller) getSparkApplication(namespace string, name string) (*v1.SparkApplication, error) {
	app, err := controller.sparkApplicationLister.SparkApplications(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return app, nil
}

func (controller *Controller) recordSparkApplicationEvent(app *v1.SparkApplication) {
	switch app.Status.AppState.State {
	case v1.NewState:
		controller.recorder.Eventf(app, apiv1.EventTypeNormal, "SparkApplicationAdded", "SparkApplication %s was added, enqueuing it for submission", app.Name)

	case v1.FailedSubmissionState:
		controller.recorder.Eventf(app, apiv1.EventTypeWarning, "SparkApplicationSubmissionFailed", "failed to submit SparkApplication %s: %s", app.Name, app.Status.AppState.ErrorMessage)

	case v1.SubmittedState:
		controller.recorder.Eventf(app, apiv1.EventTypeNormal, "SparkApplicationSubmitted", "SparkApplication %s was submitted successfully", app.Name)
	}

}

// Clean up when the spark application is terminated.
func (controller *Controller) cleanUpOnTermination(oldApp, newApp *v1.SparkApplication) error {

	return nil
}

// INFO: NodeSelector 与 DriverNodeSelector / ExecutorNodeSelector 互斥的
func (controller *Controller) validateSparkApplication(app *v1.SparkApplication) error {
	appSpec := app.Spec
	if appSpec.NodeSelector != nil && (appSpec.Driver.NodeSelector != nil || appSpec.Executor.NodeSelector != nil) {
		return fmt.Errorf("NodeSelector property can be defined at SparkApplication or at any of Driver,Executor")
	}

	return nil
}

func (controller *Controller) shouldDoBatchScheduling(app *v1.SparkApplication) (schedulerinterface.BatchScheduler, bool) {
	if controller.batchSchedulerMgr == nil || app.Spec.BatchScheduler == nil || *app.Spec.BatchScheduler == "" {
		return nil, false
	}

	scheduler, err := controller.batchSchedulerMgr.GetScheduler(*app.Spec.BatchScheduler)
	if err != nil {
		klog.Errorf("failed to get batch scheduler for name %s, %v", *app.Spec.BatchScheduler, err)
		return nil, false
	}

	return scheduler, scheduler.ShouldSchedule(app)
}

// INFO: 使用 `spark-submit` 来提交 SparkApplication 中定义的
func (controller *Controller) submitSparkApplication(app *v1.SparkApplication) *v1.SparkApplication {
	// INFO: DoBatchSchedulingOnSubmission 做了两个逻辑：1. 创建或更新 podgroup 对象；2. 更新 SparkApplication driver/executor annotation 值
	if scheduler, needScheduling := controller.shouldDoBatchScheduling(app); needScheduling {
		err := scheduler.DoBatchSchedulingOnSubmission(app)
		if err != nil {
			klog.Errorf("failed to process batch scheduler BeforeSubmitSparkApplication with error %v", err)
			return app
		}
	}

	driverPodName := getDriverPodName(app)
	submissionID := uuid.New().String()
	submissionCmdArgs, err := buildSubmissionCommandArgs(app, driverPodName, submissionID)
	if err != nil {
		app.Status = v1.SparkApplicationStatus{
			AppState: v1.ApplicationState{
				State:        v1.FailedSubmissionState,
				ErrorMessage: err.Error(),
			},
			SubmissionAttempts:        app.Status.SubmissionAttempts + 1,
			LastSubmissionAttemptTime: metav1.Now(),
		}
		return app
	}

	// Try submitting the application by running spark-submit.
	submitted, err := runSparkSubmit(newSubmission(submissionCmdArgs, app))
	if err != nil {
		app.Status = v1.SparkApplicationStatus{
			AppState: v1.ApplicationState{
				State:        v1.FailedSubmissionState,
				ErrorMessage: err.Error(),
			},
			SubmissionAttempts:        app.Status.SubmissionAttempts + 1,
			LastSubmissionAttemptTime: metav1.Now(),
		}
		controller.recordSparkApplicationEvent(app)
		klog.Errorf("failed to run spark-submit for SparkApplication %s/%s: %v", app.Namespace, app.Name, err)
		return app
	}
	if !submitted {
		// The application may not have been submitted even if err == nil, e.g., when some
		// state update caused an attempt to re-submit the application, in which case no
		// error gets returned from runSparkSubmit. If this is the case, we simply return.
		return app
	}

	klog.Infof("SparkApplication %s/%s has been submitted", app.Namespace, app.Name)
	app.Status = v1.SparkApplicationStatus{
		SubmissionID: submissionID,
		AppState: v1.ApplicationState{
			State: v1.SubmittedState,
		},
		DriverInfo: v1.DriverInfo{
			PodName: driverPodName,
		},
		SubmissionAttempts:        app.Status.SubmissionAttempts + 1,
		ExecutionAttempts:         app.Status.ExecutionAttempts + 1,
		LastSubmissionAttemptTime: metav1.Now(),
	}
	controller.recordSparkApplicationEvent(app)

	return app
}

func (controller *Controller) updateStatusAndExportMetrics(oldApp, newApp *v1.SparkApplication) error {
	// Skip update if nothing changed.
	if equality.Semantic.DeepEqual(oldApp.Status, newApp.Status) {
		return nil
	}

	// INFO: 这个函数可以复用
	oldStatusJSON, err := printStatus(&oldApp.Status)
	if err != nil {
		return err
	}
	newStatusJSON, err := printStatus(&newApp.Status)
	if err != nil {
		return err
	}

	klog.V(2).Infof("Update the status of SparkApplication %s/%s from:\n%s\nto:\n%s", newApp.Namespace, newApp.Name, oldStatusJSON, newStatusJSON)
	updatedApp, err := controller.updateApplicationStatusWithRetries(oldApp, func(status *v1.SparkApplicationStatus) {
		*status = newApp.Status
	})
	if err != nil {
		return err
	}

	// Export metrics if the update was successful.
	if controller.metrics != nil {
		controller.metrics.exportMetrics(oldApp, updatedApp)
	}

	return nil
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
func (controller *Controller) syncSparkApplication(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return fmt.Errorf("failed to get the namespace and name from key %s: %v", key, err)
	}
	app, err := controller.getSparkApplication(namespace, name)
	if err != nil {
		return err
	}
	if app == nil {
		// INFO: SparkApplication not found, 不用管
		return nil
	}

	if !app.DeletionTimestamp.IsZero() {
		controller.handleSparkApplicationDeletion(app) // 删除了 SparkApplication
		return nil
	}

	appCopy := app.DeepCopy()
	// Apply the default values to the copy. Note that the default values applied
	// won't be sent to the API server as we only update the /status subresource.
	v1.SetSparkApplicationDefaults(appCopy)

	// INFO: 可以参考 **[Running Spark on Kubernetes](https://spark.apache.org/docs/latest/running-on-kubernetes.html)** 看看 `spark-submit` 提交spark作业流程

	// Take action based on application state.
	switch appCopy.Status.AppState.State {
	case v1.NewState:
		// INFO: (1) `spark-submit --conf ...` 提交作业，--conf 参数是由 SparkApplication 对象的字段值拼接起来的；同时还会创建 podgroup
		// v1.NewState -> v1.SubmittedState/v1.FailedSubmissionState
		controller.recordSparkApplicationEvent(appCopy)
		if err := controller.validateSparkApplication(appCopy); err != nil {
			appCopy.Status.AppState.State = v1.FailedState
			appCopy.Status.AppState.ErrorMessage = err.Error()
		} else {
			appCopy = controller.submitSparkApplication(appCopy)
		}
	case v1.SucceedingState:

	case v1.FailingState:

	case v1.FailedSubmissionState:

	case v1.InvalidatingState:

	case v1.PendingRerunState:

	case v1.SubmittedState, v1.RunningState, v1.UnknownState:
		// INFO: podgroup 已经创建，进入提交成功状态

	case v1.CompletedState, v1.FailedState:

	}

	// INFO: 更新 SparkApplication status
	if appCopy != nil {
		err = controller.updateStatusAndExportMetrics(app, appCopy)
		if err != nil {
			klog.Errorf("failed to update SparkApplication %s/%s: %v", app.Namespace, app.Name, err)
			return err
		}

		if state := appCopy.Status.AppState.State; state == v1.CompletedState || state == v1.FailedState {
			if err := controller.cleanUpOnTermination(app, appCopy); err != nil {
				klog.Errorf("failed to clean up resources for SparkApplication %s/%s: %v", app.Namespace, app.Name, err)
				return err
			}
		}
	}

	return nil
}
