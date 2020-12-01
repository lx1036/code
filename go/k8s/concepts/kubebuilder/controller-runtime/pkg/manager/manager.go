package manager

import (
	"fmt"
	"github.com/go-logr/logr"
	"k8s-lx1036/k8s/concepts/kubebuilder/controller-runtime/pkg/cache"
	"k8s-lx1036/k8s/concepts/kubebuilder/controller-runtime/pkg/client"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
	"net"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/leaderelection"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/recorder"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"time"
)

type Runnable interface {
	Start(<-chan struct{}) error
}

type Manager interface {
	Add(Runnable) error

	Elected() <-chan struct{}

	SetFields(interface{}) error

	AddMetricsExtraHandler(path string, handler http.Handler) error

	AddHealthzCheck(name string, check healthz.Checker) error

	AddReadyzCheck(name string, check healthz.Checker) error

	// Start starts all registered Controllers and blocks until the Stop channel is closed.
	Start(<-chan struct{}) error

	GetConfig() *rest.Config

	GetScheme() *runtime.Scheme

	GetClient() client.Client

	GetFieldIndexer() client.FieldIndexer

	GetCache() cache.Cache

	GetEventRecorderFor(name string) record.EventRecorder

	GetRESTMapper() meta.RESTMapper

	GetAPIReader() client.Reader

	GetWebhookServer() *webhook.Server

	GetLogger() logr.Logger
}

type Options struct {
	// yaml定义和go struct相互映射mapping
	// group,version,kind map to yaml
	Scheme *runtime.Scheme
	// map(go types) -> k8s api
	MapperProvider func(config *rest.Config) (meta.RESTMapper, error)

	NewClient NewClientFunc

	NewCache cache.NewCacheFunc

	SyncPeriod *time.Duration

	// Readiness probe endpoint name, defaults to "readyz"
	ReadinessEndpointName string

	// Liveness probe endpoint name, defaults to "healthz"
	LivenessEndpointName string

	GracefulShutdownTimeout *time.Duration

	Logger logr.Logger

	DryRunClient bool

	LeaderElectionConfig    *rest.Config
	LeaderElection          bool
	LeaderElectionNamespace string
	LeaderElectionID        string

	newResourceLock func(config *rest.Config, recorderProvider recorder.Provider, options leaderelection.Options) (resourcelock.Interface, error)

	newRecorderProvider func(config *rest.Config, scheme *runtime.Scheme, logger logr.Logger, broadcaster record.EventBroadcaster) (recorder.Provider, error)

	MetricsBindAddress string
	newMetricsListener func(addr string) (net.Listener, error)

	HealthProbeBindAddress string
	newHealthProbeListener func(addr string) (net.Listener, error)

	// webhook server
	Port    int
	Host    string
	CertDir string
}

type NewClientFunc func(cache cache.Cache, config *rest.Config, options client.Options) (client.Client, error)

func DefaultNewClient(cache cache.Cache, config *rest.Config, options client.Options) (client.Client, error) {
	c, err := client.New(config, options)
	if err != nil {
		return nil, err
	}

}

func setOptionsDefaults(options Options) Options {
	if options.Scheme == nil {
		options.Scheme = scheme.Scheme
	}

	if options.MapperProvider == nil {
		options.MapperProvider = func(config *rest.Config) (meta.RESTMapper, error) {
			return client.NewDynamicRESTMapper(config)
		}
	}

	// Allow newClient to be mocked
	if options.NewClient == nil {
		options.NewClient = DefaultNewClient
	}

	if options.NewCache == nil {
		options.NewCache = cache.New
	}

	if options.ReadinessEndpointName == "" {
		options.ReadinessEndpointName = defaultReadinessEndpoint
	}

	if options.LivenessEndpointName == "" {
		options.LivenessEndpointName = defaultLivenessEndpoint
	}

	if options.newHealthProbeListener == nil {
		options.newHealthProbeListener = defaultHealthProbeListener
	}

	if options.GracefulShutdownTimeout == nil {
		gracefulShutdownTimeout := defaultGracefulShutdownPeriod
		options.GracefulShutdownTimeout = &gracefulShutdownTimeout
	}

	if options.Logger == nil {
		options.Logger = logf.Log
	}

	return options
}

func New(config *rest.Config, options Options) (Manager, error) {
	if config == nil {
		return nil, fmt.Errorf("must specify Config")
	}

	options = setOptionsDefaults(options)

	// Create the mapper provider
	mapper, err := options.MapperProvider(config)
	if err != nil {
		log.Error(err, "Failed to get API Group-Resources")
		return nil, err
	}

	cacheInformer, err := options.NewCache(config, cache.Options{Scheme: options.Scheme, Mapper: mapper, Resync: options.SyncPeriod, Namespace: options.Namespace})
	if err != nil {
		return nil, err
	}

	apiReader, err := client.New(config, client.Options{Scheme: options.Scheme, Mapper: mapper})
	if err != nil {
		return nil, err
	}

	writeObj, err := options.NewClient(cacheInformer, config, client.Options{Scheme: options.Scheme, Mapper: mapper})
	if err != nil {
		return nil, err
	}

	if options.DryRunClient {
		writeObj = client.NewDryRunClient(writeObj)
	}

	recorderProvider, err := options.newRecorderProvider(config, options.Scheme, log.WithName("events"), options.EventBroadcaster)
	if err != nil {
		return nil, err
	}

	leaderConfig := config
	if options.LeaderElectionConfig != nil {
		leaderConfig = options.LeaderElectionConfig
	}
	resourceLock, err := options.newResourceLock(leaderConfig, recorderProvider, leaderelection.Options{
		LeaderElection:          options.LeaderElection,
		LeaderElectionID:        options.LeaderElectionID,
		LeaderElectionNamespace: options.LeaderElectionNamespace,
	})
	if err != nil {
		return nil, err
	}

	metricsListener, err := options.newMetricsListener(options.MetricsBindAddress)
	if err != nil {
		return nil, err
	}

	metricsExtraHandlers := make(map[string]http.Handler)

	healthProbeListener, err := options.newHealthProbeListener(options.HealthProbeBindAddress)
	if err != nil {
		return nil, err
	}

	stop := make(chan struct{})

	return &controllerManager{
		config:                  config,
		scheme:                  options.Scheme,
		cache:                   cacheInformer,
		fieldIndexes:            cacheInformer,
		client:                  writeObj,
		apiReader:               apiReader,
		recorderProvider:        recorderProvider,
		resourceLock:            resourceLock,
		mapper:                  mapper,
		metricsListener:         metricsListener,
		metricsExtraHandlers:    metricsExtraHandlers,
		logger:                  options.Logger,
		internalStop:            stop,
		internalStopper:         stop,
		elected:                 make(chan struct{}),
		port:                    options.Port,
		host:                    options.Host,
		certDir:                 options.CertDir,
		leaseDuration:           *options.LeaseDuration,
		renewDeadline:           *options.RenewDeadline,
		retryPeriod:             *options.RetryPeriod,
		healthProbeListener:     healthProbeListener,
		readinessEndpointName:   options.ReadinessEndpointName,
		livenessEndpointName:    options.LivenessEndpointName,
		gracefulShutdownTimeout: *options.GracefulShutdownTimeout,
	}, nil
}
