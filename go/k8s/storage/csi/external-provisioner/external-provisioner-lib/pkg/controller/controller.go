package controller

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/time/rate"

	v1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/uuid"
	utilversion "k8s.io/apimachinery/pkg/util/version"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	corelistersv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

const (
	// DefaultResyncPeriod is used when option function ResyncPeriod is omitted
	DefaultResyncPeriod = 15 * time.Minute
	// DefaultThreadiness is used when option function Threadiness is omitted
	DefaultThreadiness = 4
	// DefaultExponentialBackOffOnError is used when option function ExponentialBackOffOnError is omitted
	DefaultExponentialBackOffOnError = true
	// DefaultCreateProvisionedPVRetryCount is used when option function CreateProvisionedPVRetryCount is omitted
	DefaultCreateProvisionedPVRetryCount = 5
	// DefaultCreateProvisionedPVInterval is used when option function CreateProvisionedPVInterval is omitted
	DefaultCreateProvisionedPVInterval = 10 * time.Second
	// DefaultFailedProvisionThreshold is used when option function FailedProvisionThreshold is omitted
	DefaultFailedProvisionThreshold = 15
	// DefaultFailedDeleteThreshold is used when option function FailedDeleteThreshold is omitted
	DefaultFailedDeleteThreshold = 15
	// DefaultLeaderElection is used when option function LeaderElection is omitted
	DefaultLeaderElection = true
	// DefaultLeaseDuration is used when option function LeaseDuration is omitted
	DefaultLeaseDuration = 15 * time.Second
	// DefaultRenewDeadline is used when option function RenewDeadline is omitted
	DefaultRenewDeadline = 10 * time.Second
	// DefaultRetryPeriod is used when option function RetryPeriod is omitted
	DefaultRetryPeriod = 2 * time.Second
	// DefaultMetricsPort is used when option function MetricsPort is omitted
	DefaultMetricsPort = 0
	// DefaultMetricsAddress is used when option function MetricsAddress is omitted
	DefaultMetricsAddress = "0.0.0.0"
	// DefaultMetricsPath is used when option function MetricsPath is omitted
	DefaultMetricsPath = "/metrics"
	// DefaultAddFinalizer is used when option function AddFinalizer is omitted
	DefaultAddFinalizer = false
)

var errRuntime = fmt.Errorf("cannot call option functions after controller has Run")

// ProvisionController is a controller that provisions PersistentVolumes for
// PersistentVolumeClaims.
type ProvisionController struct {
	client kubernetes.Interface

	// The name of the provisioner for which this controller dynamically
	// provisions volumes. The value of annDynamicallyProvisioned and
	// annStorageProvisioner to set & watch for, respectively
	provisionerName string

	// additional provisioner names (beyond provisionerName) that the
	// provisioner should watch for and handle in annStorageProvisioner
	additionalProvisionerNames []string

	// The provisioner the controller will use to provision and delete volumes.
	// Presumably this implementer of Provisioner carries its own
	// volume-specific options and such that it needs in order to provision
	// volumes.
	provisioner Provisioner

	// Kubernetes cluster server version:
	// * 1.4: storage classes introduced as beta. Technically out-of-tree dynamic
	// provisioning is not officially supported, though it works
	// * 1.5: storage classes stay in beta. Out-of-tree dynamic provisioning is
	// officially supported
	// * 1.6: storage classes enter GA
	kubeVersion *utilversion.Version

	claimInformer  cache.SharedIndexInformer
	claimsIndexer  cache.Indexer
	volumeInformer cache.SharedInformer
	volumes        cache.Store
	classInformer  cache.SharedInformer
	nodeLister     corelistersv1.NodeLister
	classes        cache.Store

	// To determine if the informer is internal or external
	customClaimInformer, customVolumeInformer, customClassInformer bool

	claimQueue  workqueue.RateLimitingInterface
	volumeQueue workqueue.RateLimitingInterface

	// Identity of this controller, generated at creation time and not persisted
	// across restarts. Useful only for debugging, for seeing the source of
	// events. controller.provisioner may have its own, different notion of
	// identity which may/may not persist across restarts
	id            string
	component     string
	eventRecorder record.EventRecorder

	resyncPeriod     time.Duration
	provisionTimeout time.Duration
	deletionTimeout  time.Duration

	rateLimiter               workqueue.RateLimiter
	exponentialBackOffOnError bool
	threadiness               int

	createProvisionedPVBackoff    *wait.Backoff
	createProvisionedPVRetryCount int
	createProvisionedPVInterval   time.Duration
	createProvisionerPVLimiter    workqueue.RateLimiter

	failedProvisionThreshold, failedDeleteThreshold int

	// The metrics collection used by this controller.
	metrics metrics.Metrics
	// The port for metrics server to serve on.
	metricsPort int32
	// The IP address for metrics server to serve on.
	metricsAddress string
	// The path of metrics endpoint path.
	metricsPath string

	// Whether to add a finalizer marking the provisioner as the owner of the PV
	// with clean up duty.
	// TODO: upstream and we may have a race b/w applying reclaim policy and not if pv has protection finalizer
	addFinalizer bool

	// Whether to do kubernetes leader election at all. It should basically
	// always be done when possible to avoid duplicate Provision attempts.
	leaderElection          bool
	leaderElectionNamespace string
	// Parameters of leaderelection.LeaderElectionConfig.
	leaseDuration, renewDeadline, retryPeriod time.Duration

	hasRun     bool
	hasRunLock *sync.Mutex

	// Map UID -> *PVC with all claims that may be provisioned in the background.
	claimsInProgress sync.Map

	volumeStore VolumeStore
}

// HasRun returns whether the controller has Run
func (ctrl *ProvisionController) HasRun() bool {
	ctrl.hasRunLock.Lock()
	defer ctrl.hasRunLock.Unlock()
	return ctrl.hasRun
}

// LeaderElection determines whether to enable leader election or not. Defaults
// to true.
func LeaderElection(leaderElection bool) func(*ProvisionController) error {
	return func(c *ProvisionController) error {
		if c.HasRun() {
			return errRuntime
		}
		c.leaderElection = leaderElection
		return nil
	}
}

// FailedProvisionThreshold is the threshold for max number of retries on
// failures of Provision. Set to 0 to retry indefinitely. Defaults to 15.
func FailedProvisionThreshold(failedProvisionThreshold int) func(*ProvisionController) error {
	return func(c *ProvisionController) error {
		if c.HasRun() {
			return errRuntime
		}
		c.failedProvisionThreshold = failedProvisionThreshold
		return nil
	}
}

// FailedDeleteThreshold is the threshold for max number of retries on failures
// of Delete. Set to 0 to retry indefinitely. Defaults to 15.
func FailedDeleteThreshold(failedDeleteThreshold int) func(*ProvisionController) error {
	return func(c *ProvisionController) error {
		if c.HasRun() {
			return errRuntime
		}
		c.failedDeleteThreshold = failedDeleteThreshold
		return nil
	}
}

// RateLimiter is the workqueue.RateLimiter to use for the provisioning and
// deleting work queues. If set, ExponentialBackOffOnError is ignored.
func RateLimiter(rateLimiter workqueue.RateLimiter) func(*ProvisionController) error {
	return func(c *ProvisionController) error {
		if c.HasRun() {
			return errRuntime
		}
		c.rateLimiter = rateLimiter
		return nil
	}
}

// Threadiness is the number of claim and volume workers each to launch.
// Defaults to 4.
func Threadiness(threadiness int) func(*ProvisionController) error {
	return func(c *ProvisionController) error {
		if c.HasRun() {
			return errRuntime
		}
		c.threadiness = threadiness
		return nil
	}
}

// CreateProvisionedPVRetryCount is the number of retries when we create a PV
// object for a provisioned volume. Defaults to 5.
// If PV is not saved after given number of retries, corresponding storage asset (volume) is deleted!
// Only one of CreateProvisionedPVInterval+CreateProvisionedPVRetryCount or CreateProvisionedPVBackoff or
// CreateProvisionedPVLimiter can be used.
// Deprecated: Use CreateProvisionedPVLimiter instead, it tries indefinitely.
func CreateProvisionedPVRetryCount(createProvisionedPVRetryCount int) func(*ProvisionController) error {
	return func(c *ProvisionController) error {
		if c.HasRun() {
			return errRuntime
		}
		if c.createProvisionedPVBackoff != nil {
			return fmt.Errorf("CreateProvisionedPVBackoff cannot be used together with CreateProvisionedPVRetryCount")
		}
		if c.createProvisionerPVLimiter != nil {
			return fmt.Errorf("CreateProvisionedPVBackoff cannot be used together with CreateProvisionedPVLimiter")
		}
		c.createProvisionedPVRetryCount = createProvisionedPVRetryCount
		return nil
	}
}

// CreateProvisionedPVLimiter is the configuration of rate limiter for queue of unsaved PersistentVolumes.
// If set, PVs that fail to be saved to Kubernetes API server will be re-enqueued to a separate workqueue
// with this limiter and re-tried until they are saved to API server. There is no limit of retries.
// The main difference to other CreateProvisionedPV* option is that the storage asset is never deleted
// and the controller continues saving PV to API server indefinitely.
// This option cannot be used with CreateProvisionedPVBackoff or CreateProvisionedPVInterval
// or CreateProvisionedPVRetryCount.
func CreateProvisionedPVLimiter(limiter workqueue.RateLimiter) func(*ProvisionController) error {
	return func(c *ProvisionController) error {
		if c.HasRun() {
			return errRuntime
		}
		if c.createProvisionedPVRetryCount != 0 {
			return fmt.Errorf("CreateProvisionedPVLimiter cannot be used together with CreateProvisionedPVRetryCount")
		}
		if c.createProvisionedPVInterval != 0 {
			return fmt.Errorf("CreateProvisionedPVLimiter cannot be used together with CreateProvisionedPVInterval")
		}
		if c.createProvisionedPVBackoff != nil {
			return fmt.Errorf("CreateProvisionedPVLimiter cannot be used together with CreateProvisionedPVBackoff")
		}
		c.createProvisionerPVLimiter = limiter
		return nil
	}
}

// ClaimsInformer sets the informer to use for accessing PersistentVolumeClaims.
// Defaults to using a internal informer.
func ClaimsInformer(informer cache.SharedIndexInformer) func(*ProvisionController) error {
	return func(c *ProvisionController) error {
		if c.HasRun() {
			return errRuntime
		}
		c.claimInformer = informer
		c.customClaimInformer = true
		return nil
	}
}

// NodesLister sets the informer to use for accessing Nodes.
// This is needed only for PVCs which have a selected node.
// Defaults to using a GET instead of an informer.
//
// Which approach is better depends on factors like cluster size and
// ratio of PVCs with a selected node.
func NodesLister(nodeLister corelistersv1.NodeLister) func(*ProvisionController) error {
	return func(c *ProvisionController) error {
		if c.HasRun() {
			return errRuntime
		}
		c.nodeLister = nodeLister
		return nil
	}
}

// AdditionalProvisionerNames sets additional names for the provisioner
func AdditionalProvisionerNames(additionalProvisionerNames []string) func(*ProvisionController) error {
	return func(c *ProvisionController) error {
		if c.HasRun() {
			return errRuntime
		}
		c.additionalProvisionerNames = additionalProvisionerNames
		return nil
	}
}

// NewProvisionController creates a new provision controller using
// the given configuration parameters and with private (non-shared) informers.
func NewProvisionController(
	client kubernetes.Interface,
	provisionerName string,
	provisioner Provisioner,
	kubeVersion string,
	options ...func(*ProvisionController) error,
) *ProvisionController {
	id, err := os.Hostname()
	if err != nil {
		klog.Fatalf("Error getting hostname: %v", err)
	}
	// add a uniquifier so that two processes on the same host don't accidentally both become active
	id = id + "_" + string(uuid.NewUUID())
	component := provisionerName + "_" + id

	v1.AddToScheme(scheme.Scheme)
	broadcaster := record.NewBroadcaster()
	broadcaster.StartLogging(klog.Infof)
	broadcaster.StartRecordingToSink(&corev1.EventSinkImpl{Interface: client.CoreV1().Events(v1.NamespaceAll)})
	eventRecorder := broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: component})

	controller := &ProvisionController{
		client:                    client,
		provisionerName:           provisionerName,
		provisioner:               provisioner,
		kubeVersion:               utilversion.MustParseSemantic(kubeVersion),
		id:                        id,
		component:                 component,
		eventRecorder:             eventRecorder,
		resyncPeriod:              DefaultResyncPeriod,
		exponentialBackOffOnError: DefaultExponentialBackOffOnError,
		threadiness:               DefaultThreadiness,
		failedProvisionThreshold:  DefaultFailedProvisionThreshold,
		failedDeleteThreshold:     DefaultFailedDeleteThreshold,
		leaderElection:            DefaultLeaderElection,
		leaderElectionNamespace:   getInClusterNamespace(),
		leaseDuration:             DefaultLeaseDuration,
		renewDeadline:             DefaultRenewDeadline,
		retryPeriod:               DefaultRetryPeriod,
		metrics:                   metrics.M,
		metricsPort:               DefaultMetricsPort,
		metricsAddress:            DefaultMetricsAddress,
		metricsPath:               DefaultMetricsPath,
		addFinalizer:              DefaultAddFinalizer,
		hasRun:                    false,
		hasRunLock:                &sync.Mutex{},
	}

	for _, option := range options {
		err := option(controller)
		if err != nil {
			klog.Fatalf("Error processing controller options: %s", err)
		}
	}

	var rateLimiter workqueue.RateLimiter
	if controller.rateLimiter != nil {
		// rateLimiter set via parameter takes precedence
		rateLimiter = controller.rateLimiter
	} else if controller.exponentialBackOffOnError {
		rateLimiter = workqueue.NewMaxOfRateLimiter(
			workqueue.NewItemExponentialFailureRateLimiter(15*time.Second, 1000*time.Second),
			&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(10), 100)},
		)
	} else {
		rateLimiter = workqueue.NewMaxOfRateLimiter(
			workqueue.NewItemExponentialFailureRateLimiter(15*time.Second, 15*time.Second),
			&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(10), 100)},
		)
	}
	controller.claimQueue = workqueue.NewNamedRateLimitingQueue(rateLimiter, "claims")
	controller.volumeQueue = workqueue.NewNamedRateLimitingQueue(rateLimiter, "volumes")

	informer := informers.NewSharedInformerFactory(client, controller.resyncPeriod)

	// ----------------------
	// PersistentVolumeClaims

	claimHandler := cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) { controller.enqueueClaim(obj) },
		UpdateFunc: func(oldObj, newObj interface{}) { controller.enqueueClaim(newObj) },
		DeleteFunc: func(obj interface{}) {
			// NOOP. The claim is either in claimsInProgress and in the queue, so it will be processed as usual
			// or it's not in claimsInProgress and then we don't care
		},
	}

	if controller.claimInformer != nil {
		controller.claimInformer.AddEventHandlerWithResyncPeriod(claimHandler, controller.resyncPeriod)
	} else {
		controller.claimInformer = informer.Core().V1().PersistentVolumeClaims().Informer()
		controller.claimInformer.AddEventHandler(claimHandler)
	}
	err = controller.claimInformer.AddIndexers(cache.Indexers{uidIndex: func(obj interface{}) ([]string, error) {
		uid, err := getObjectUID(obj)
		if err != nil {
			return nil, err
		}
		return []string{uid}, nil
	}})
	if err != nil {
		klog.Fatalf("Error setting indexer %s for pvc informer: %v", uidIndex, err)
	}
	controller.claimsIndexer = controller.claimInformer.GetIndexer()

	// -----------------
	// PersistentVolumes

	volumeHandler := cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) { controller.enqueueVolume(obj) },
		UpdateFunc: func(oldObj, newObj interface{}) { controller.enqueueVolume(newObj) },
		DeleteFunc: func(obj interface{}) { controller.forgetVolume(obj) },
	}

	if controller.volumeInformer != nil {
		controller.volumeInformer.AddEventHandlerWithResyncPeriod(volumeHandler, controller.resyncPeriod)
	} else {
		controller.volumeInformer = informer.Core().V1().PersistentVolumes().Informer()
		controller.volumeInformer.AddEventHandler(volumeHandler)
	}
	controller.volumes = controller.volumeInformer.GetStore()

	// --------------
	// StorageClasses

	// no resource event handler needed for StorageClasses
	if controller.classInformer == nil {
		if controller.kubeVersion.AtLeast(utilversion.MustParseSemantic("v1.6.0")) {
			controller.classInformer = informer.Storage().V1().StorageClasses().Informer()
		} else {
			controller.classInformer = informer.Storage().V1beta1().StorageClasses().Informer()
		}
	}
	controller.classes = controller.classInformer.GetStore()

	if controller.createProvisionerPVLimiter != nil {
		klog.V(2).Infof("Using saving PVs to API server in background")
		controller.volumeStore = NewVolumeStoreQueue(client, controller.createProvisionerPVLimiter, controller.claimsIndexer, controller.eventRecorder)
	} else {
		if controller.createProvisionedPVBackoff == nil {
			// Use linear backoff with createProvisionedPVInterval and createProvisionedPVRetryCount by default.
			if controller.createProvisionedPVInterval == 0 {
				controller.createProvisionedPVInterval = DefaultCreateProvisionedPVInterval
			}
			if controller.createProvisionedPVRetryCount == 0 {
				controller.createProvisionedPVRetryCount = DefaultCreateProvisionedPVRetryCount
			}
			controller.createProvisionedPVBackoff = &wait.Backoff{
				Duration: controller.createProvisionedPVInterval,
				Factor:   1, // linear backoff
				Steps:    controller.createProvisionedPVRetryCount,
				//Cap:      controller.createProvisionedPVInterval,
			}
		}
		klog.V(2).Infof("Using blocking saving PVs to API server")
		controller.volumeStore = NewBackoffStore(client, controller.eventRecorder, controller.createProvisionedPVBackoff, controller)
	}

	return controller
}

// Run starts all of this controller's control loops
func (ctrl *ProvisionController) Run(ctx context.Context) {
	run := func(ctx context.Context) {
		klog.Infof("Starting provisioner controller %s!", ctrl.component)
		defer utilruntime.HandleCrash()
		defer ctrl.claimQueue.ShutDown()
		defer ctrl.volumeQueue.ShutDown()

		ctrl.hasRunLock.Lock()
		ctrl.hasRun = true
		ctrl.hasRunLock.Unlock()
		if ctrl.metricsPort > 0 {
			prometheus.MustRegister([]prometheus.Collector{
				metrics.PersistentVolumeClaimProvisionTotal,
				metrics.PersistentVolumeClaimProvisionFailedTotal,
				metrics.PersistentVolumeClaimProvisionDurationSeconds,
				metrics.PersistentVolumeDeleteTotal,
				metrics.PersistentVolumeDeleteFailedTotal,
				metrics.PersistentVolumeDeleteDurationSeconds,
			}...)
			http.Handle(ctrl.metricsPath, promhttp.Handler())
			address := net.JoinHostPort(ctrl.metricsAddress, strconv.FormatInt(int64(ctrl.metricsPort), 10))
			klog.Infof("Starting metrics server at %s\n", address)
			go wait.Forever(func() {
				err := http.ListenAndServe(address, nil)
				if err != nil {
					klog.Errorf("Failed to listen on %s: %v", address, err)
				}
			}, 5*time.Second)
		}

		// If a external SharedInformer has been passed in, this controller
		// should not call Run again
		if !ctrl.customClaimInformer {
			go ctrl.claimInformer.Run(ctx.Done())
		}
		if !ctrl.customVolumeInformer {
			go ctrl.volumeInformer.Run(ctx.Done())
		}
		if !ctrl.customClassInformer {
			go ctrl.classInformer.Run(ctx.Done())
		}

		if !cache.WaitForCacheSync(ctx.Done(), ctrl.claimInformer.HasSynced, ctrl.volumeInformer.HasSynced, ctrl.classInformer.HasSynced) {
			return
		}

		for i := 0; i < ctrl.threadiness; i++ {
			go wait.Until(func() { ctrl.runClaimWorker(ctx) }, time.Second, ctx.Done())
			go wait.Until(func() { ctrl.runVolumeWorker(ctx) }, time.Second, ctx.Done())
		}

		klog.Infof("Started provisioner controller %s!", ctrl.component)

		select {}
	}

	go ctrl.volumeStore.Run(ctx, DefaultThreadiness)

	if ctrl.leaderElection {
		rl, err := resourcelock.New("endpoints",
			ctrl.leaderElectionNamespace,
			strings.Replace(ctrl.provisionerName, "/", "-", -1),
			ctrl.client.CoreV1(),
			nil,
			resourcelock.ResourceLockConfig{
				Identity:      ctrl.id,
				EventRecorder: ctrl.eventRecorder,
			})
		if err != nil {
			klog.Fatalf("Error creating lock: %v", err)
		}

		leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
			Lock:          rl,
			LeaseDuration: ctrl.leaseDuration,
			RenewDeadline: ctrl.renewDeadline,
			RetryPeriod:   ctrl.retryPeriod,
			Callbacks: leaderelection.LeaderCallbacks{
				OnStartedLeading: run,
				OnStoppedLeading: func() {
					klog.Fatalf("leaderelection lost")
				},
			},
		})
		panic("unreachable")
	} else {
		run(ctx)
	}
}
