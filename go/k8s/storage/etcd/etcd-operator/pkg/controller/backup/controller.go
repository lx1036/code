package backup

import (
	"fmt"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"time"

	"k8s-lx1036/k8s/storage/etcd/etcd-operator/cmd/backup/app/options"
	v1 "k8s-lx1036/k8s/storage/etcd/etcd-operator/pkg/apis/etcd.k9s.io/v1"
	"k8s-lx1036/k8s/storage/etcd/etcd-operator/pkg/client/clientset/versioned"
	"k8s-lx1036/k8s/storage/etcd/etcd-operator/pkg/client/informers/externalversions"
	etcdLister "k8s-lx1036/k8s/storage/etcd/etcd-operator/pkg/client/listers/etcd.k9s.io/v1"
	"k8s-lx1036/k8s/storage/etcd/etcd-operator/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

const (
	Name = "etcd-backup-controller"
)

var (
	keyFunc = cache.DeletionHandlingMetaNamespaceKeyFunc
)

type EtcdBackupController struct {
	kubeClient kubernetes.Interface

	etcdBackupInformer cache.SharedIndexInformer

	queue workqueue.RateLimitingInterface

	recorder record.EventRecorder

	etcdBackupLister etcdLister.EtcdBackupLister
	etcdBackupClient *versioned.Clientset
}

func NewController(option *options.Options) (*EtcdBackupController, error) {
	restConfig, err := utils.NewRestConfig(option.Kubeconfig)
	if err != nil {
		return nil, err
	}

	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to construct lister client: %v", err)
	}

	var etcdBackupFactoryOpts []externalversions.SharedInformerOption
	if option.Namespace != corev1.NamespaceAll {
		etcdBackupFactoryOpts = append(etcdBackupFactoryOpts, externalversions.WithNamespace(option.Namespace))
	}

	etcdBackupClient := versioned.NewForConfigOrDie(restConfig)
	etcdBackupInformerFactory := externalversions.NewSharedInformerFactoryWithOptions(etcdBackupClient, time.Second*30, etcdBackupFactoryOpts...)
	etcdBackupInformer := etcdBackupInformerFactory.Etcd().V1().EtcdBackups().Informer()

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

	controller := &EtcdBackupController{
		etcdBackupInformer: etcdBackupInformer,
		kubeClient:         kubeClient,
		queue:              queue,
		recorder:           eventRecorder,
		etcdBackupClient:   etcdBackupClient,
		etcdBackupLister:   etcdBackupInformerFactory.Etcd().V1().EtcdBackups().Lister(),
	}

	etcdBackupInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.onEtcdBackupAdd,
		UpdateFunc: controller.onEtcdBackupUpdate,
		DeleteFunc: controller.onEtcdBackupDelete,
	})

	return controller, nil
}

func (controller *EtcdBackupController) onEtcdBackupAdd(obj interface{}) {
	etcdBackup := obj.(*v1.EtcdBackup)
	klog.Infof("EtcdBackup %s/%s was added, enqueuing it for submission", etcdBackup.Namespace, etcdBackup.Name)
	controller.enqueue(etcdBackup)
}

func (controller *EtcdBackupController) onEtcdBackupUpdate(oldObj, newObj interface{}) {
	oldEtcdBackup := oldObj.(*v1.EtcdBackup)
	newEtcdBackup := newObj.(*v1.EtcdBackup)

	// The informer will call this function on non-updated resources during resync, avoid
	// enqueuing unchanged applications, unless it has expired or is subject to retry.
	if oldEtcdBackup.ResourceVersion == newEtcdBackup.ResourceVersion {
		//if oldApp.ResourceVersion == newApp.ResourceVersion && !controller.hasApplicationExpired(newApp) && !shouldRetry(newApp) {
		return
	}

	klog.Infof("EtcdBackup %s/%s was updated, enqueuing it", newEtcdBackup.Namespace, newEtcdBackup.Name)
	controller.enqueue(newEtcdBackup)
}

func (controller *EtcdBackupController) onEtcdBackupDelete(obj interface{}) {
	etcdBackup := obj.(*v1.EtcdBackup)
	klog.Infof("EtcdBackup %s/%s was deleted, enqueuing it for submission", etcdBackup.Namespace, etcdBackup.Name)
	controller.enqueue(etcdBackup)
}

func (controller *EtcdBackupController) enqueue(obj interface{}) {
	key, err := keyFunc(obj)
	if err != nil {
		klog.Errorf("failed to get key for %v: %v", obj, err)
		return
	}

	// INFO: AddRateLimited() 比 Add() 更好在于，AddRateLimited() 有限速器，会在 RateLimiter ok 之后才会 Add()，以后用 AddRateLimited()
	controller.queue.AddRateLimited(key)
}

func (controller *EtcdBackupController) Start(workers int, stopCh <-chan struct{}) error {
	go controller.etcdBackupInformer.Run(stopCh)

	shutdown := cache.WaitForCacheSync(stopCh, controller.etcdBackupInformer.HasSynced)
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

// runWorker runs a single controller worker.
func (controller *EtcdBackupController) runWorker() {
	defer utilruntime.HandleCrash()
	for controller.processNextItem() {
	}
}

func (controller *EtcdBackupController) processNextItem() bool {
	key, quit := controller.queue.Get()
	if quit {
		return false
	}
	defer controller.queue.Done(key)

	klog.V(2).Infof("Starting processing key: %q", key)
	defer klog.V(2).Infof("Ending processing key: %q", key)
	err := controller.syncEtcdBackup(key.(string))
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

func (controller *EtcdBackupController) syncEtcdBackup(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}
	etcdBackup, err := controller.etcdBackupLister.EtcdBackups(namespace).Get(name)
	if err != nil {
		if apierrors.IsNotFound(err) { // deleted etcdBackup
			// TODO:
			return nil
		}

		return err
	}

	if etcdBackup.DeletionTimestamp != nil { // deleting etcdBackup
		// TODO:
		return nil
	}

	isPeriodic := isPeriodicBackup(etcdBackup)
	// don't process the CR if it has a status since
	// having a status means that the backup is either made or failed.
	if !isPeriodic && (etcdBackup.Status.Succeeded || len(etcdBackup.Status.Reason) != 0) {
		return nil
	}

	if !isPeriodic {
		err = controller.handleBackup(etcdBackup)
	} else if isPeriodic && controller.isChanged(etcdBackup) {

	}

	return err
}

func (controller *EtcdBackupController) handleBackup(etcdBackup *v1.EtcdBackup) error {
	err := validate(etcdBackup)
	if err != nil {
		return err
	}

	switch etcdBackup.Spec.StorageType {
	case v1.BackupStorageTypeS3:
		handleS3Backup(etcdBackup, controller.kubeClient)
	}

}

func isPeriodicBackup(etcdBackup *v1.EtcdBackup) bool {
	if etcdBackup.Spec.BackupPolicy != nil {
		return etcdBackup.Spec.BackupPolicy.BackupIntervalInSecond != 0
	}

	return false
}

// TODO: 这个应该放在 api 定义处的 validation
func validate(etcdBackup *v1.EtcdBackup) error {
	if len(etcdBackup.Spec.EtcdEndpoints) == 0 {
		return fmt.Errorf("spec.etcdEndpoints should not be empty")
	}
	if etcdBackup.Spec.BackupPolicy != nil {
		if etcdBackup.Spec.BackupPolicy.BackupIntervalInSecond < 0 {
			return fmt.Errorf("spec.BackupPolicy.BackupIntervalInSecond should not be lower than 0")
		}
		if etcdBackup.Spec.BackupPolicy.MaxBackups < 0 {
			return fmt.Errorf("spec.BackupPolicy.MaxBackups should not be lower than 0")
		}
	}

	return nil
}
