package controller

import (
	"sync"
	"time"

	utilversion "k8s.io/apimachinery/pkg/util/version"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	corelistersv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

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
