package manager

import (
	"context"
	"github.com/go-logr/logr"
	"k8s-lx1036/k8s/concepts/tools/kubebuilder/controller-runtime/pkg/cache"
	"k8s-lx1036/k8s/concepts/tools/kubebuilder/controller-runtime/pkg/client"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
	"net"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sync"
	"time"
)

const (
	// Values taken from: https://github.com/kubernetes/apiserver/blob/master/pkg/apis/config/v1alpha1/defaults.go
	defaultLeaseDuration          = 15 * time.Second
	defaultRenewDeadline          = 10 * time.Second
	defaultRetryPeriod            = 2 * time.Second
	defaultGracefulShutdownPeriod = 30 * time.Second

	defaultReadinessEndpoint = "/readyz/"
	defaultLivenessEndpoint  = "/healthz/"
	defaultMetricsEndpoint   = "/metrics"
)

var log = logf.RuntimeLog.WithName("manager")

type controllerManager struct {
	config *rest.Config

	scheme *runtime.Scheme

	// leaderElectionRunnables is the set of Controllers that the controllerManager injects deps into and Starts.
	// These Runnables are managed by lead election.
	leaderElectionRunnables []Runnable
	// nonLeaderElectionRunnables is the set of webhook servers that the controllerManager injects deps into and Starts.
	// These Runnables will not be blocked by lead election.
	nonLeaderElectionRunnables []Runnable

	cache cache.Cache

	client client.Client

	apiReader client.Reader

	fieldIndexes client.FieldIndexer

	recorderProvider *intrec.Provider

	// resourceLock forms the basis for leader election
	resourceLock resourcelock.Interface

	// leaderElectionReleaseOnCancel defines if the manager should step back from the leader lease
	// on shutdown
	leaderElectionReleaseOnCancel bool

	// mapper is used to map resources to kind, and map kind and version.
	mapper meta.RESTMapper

	// metricsListener is used to serve prometheus metrics
	metricsListener net.Listener

	// metricsExtraHandlers contains extra handlers to register on http server that serves metrics.
	metricsExtraHandlers map[string]http.Handler

	// healthProbeListener is used to serve liveness probe
	healthProbeListener net.Listener

	// Readiness probe endpoint name
	readinessEndpointName string

	// Liveness probe endpoint name
	livenessEndpointName string

	// Readyz probe handler
	readyzHandler *healthz.Handler

	// Healthz probe handler
	healthzHandler *healthz.Handler

	mu             sync.Mutex
	started        bool
	startedLeader  bool
	healthzStarted bool
	errChan        chan error

	// internalStop is the stop channel *actually* used by everything involved
	// with the manager as a stop channel, so that we can pass a stop channel
	// to things that need it off the bat (like the Channel source).  It can
	// be closed via `internalStopper` (by being the same underlying channel).
	internalStop <-chan struct{}

	// internalStopper is the write side of the internal stop channel, allowing us to close it.
	// It and `internalStop` should point to the same channel.
	internalStopper chan<- struct{}

	// Logger is the logger that should be used by this manager.
	// If none is set, it defaults to log.Log global logger.
	logger logr.Logger

	// leaderElectionCancel is used to cancel the leader election. It is distinct from internalStopper,
	// because for safety reasons we need to os.Exit() when we lose the leader election, meaning that
	// it must be deferred until after gracefulShutdown is done.
	leaderElectionCancel context.CancelFunc

	// stop procedure engaged. In other words, we should not add anything else to the manager
	stopProcedureEngaged bool

	// elected is closed when this manager becomes the leader of a group of
	// managers, either because it won a leader election or because no leader
	// election was configured.
	elected chan struct{}

	startCache func(stop <-chan struct{}) error

	// port is the port that the webhook server serves at.
	port int
	// host is the hostname that the webhook server binds to.
	host string
	// CertDir is the directory that contains the server key and certificate.
	// if not set, webhook server would look up the server key and certificate in
	// {TempDir}/k8s-webhook-server/serving-certs
	certDir string

	webhookServer *webhook.Server

	// leaseDuration is the duration that non-leader candidates will
	// wait to force acquire leadership.
	leaseDuration time.Duration
	// renewDeadline is the duration that the acting controlplane will retry
	// refreshing leadership before giving up.
	renewDeadline time.Duration
	// retryPeriod is the duration the LeaderElector clients should wait
	// between tries of actions.
	retryPeriod time.Duration

	// waitForRunnable is holding the number of runnables currently running so that
	// we can wait for them to exit before quitting the manager
	waitForRunnable sync.WaitGroup

	// gracefulShutdownTimeout is the duration given to runnable to stop
	// before the manager actually returns on stop.
	gracefulShutdownTimeout time.Duration

	// onStoppedLeading is callled when the leader election lease is lost.
	// It can be overridden for tests.
	onStoppedLeading func()

	// shutdownCtx is the context that can be used during shutdown. It will be cancelled
	// after the gracefulShutdownTimeout ended. It must not be accessed before internalStop
	// is closed because it will be nil.
	shutdownCtx context.Context
}

func (cm *controllerManager) Add(runnable Runnable) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Add the runnable to the leader election or the non-leaderelection list

}

func (cm *controllerManager) Elected() <-chan struct{} {
	panic("implement me")
}

func (cm *controllerManager) SetFields(i interface{}) error {

	return nil
}

func (cm *controllerManager) AddMetricsExtraHandler(path string, handler http.Handler) error {
	panic("implement me")
}

func (cm *controllerManager) AddHealthzCheck(name string, check Checker) error {
	panic("implement me")
}

func (cm *controllerManager) AddReadyzCheck(name string, check Checker) error {
	panic("implement me")
}

func (cm *controllerManager) GetConfig() *rest.Config {
	panic("implement me")
}

func (cm *controllerManager) GetScheme() *runtime.Scheme {
	panic("implement me")
}

func (cm *controllerManager) GetClient() client.Client {
	panic("implement me")
}

func (cm *controllerManager) GetFieldIndexer() client.FieldIndexer {
	panic("implement me")
}

func (cm *controllerManager) GetCache() cache.Cache {
	panic("implement me")
}

func (cm *controllerManager) GetEventRecorderFor(name string) record.EventRecorder {
	panic("implement me")
}

func (cm *controllerManager) GetRESTMapper() meta.RESTMapper {
	panic("implement me")
}

func (cm *controllerManager) GetAPIReader() client.Reader {
	panic("implement me")
}

func (cm *controllerManager) GetWebhookServer() *Server {
	panic("implement me")
}

func (cm *controllerManager) GetLogger() logr.Logger {
	return cm.logger
}

func (cm *controllerManager) waitForCache() {

	cm.cache.WaitForCacheSync(cm.internalStop)

}
func (cm *controllerManager) startNonLeaderElectionRunnables() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.waitForCache()

	for _, c := range cm.nonLeaderElectionRunnables {
		cm.startRunnable(c)
	}
}
func (cm *controllerManager) startRunnable(r Runnable) {
	cm.waitForRunnable.Add(1)
	go func() {
		defer cm.waitForRunnable.Done()
		if err := r.Start(cm.internalStop); err != nil {
			cm.errChan <- err
		}
	}()
}

func (cm *controllerManager) Start(stop <-chan struct{}) error {
	stopComplete := make(chan struct{})
	defer close(stopComplete)

	go cm.startNonLeaderElectionRunnables()

	select {
	case <-stop:
		// We are done
		return nil
	case err := <-cm.errChan:
		// Error starting or running a runnable
		return err
	}
}
