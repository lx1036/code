package manager

import (
	"fmt"
	"time"
	"github.com/go-logr/logr"
	"k8s-lx1036/k8s/concepts/kubebuilder/controller-runtime/pkg/cache"
	"k8s-lx1036/k8s/concepts/kubebuilder/controller-runtime/pkg/client"
	"k8s-lx1036/k8s/concepts/kubebuilder/controller-runtime/pkg/client/apiutil"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"net"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

)

type Manager interface {

}

type Options struct {
	// yaml定义和go struct相互映射mapping
	// group,version,kind map to yaml
	Scheme *runtime.Scheme
	// map(go types) -> k8s api
	MapperProvider func(config *rest.Config) (meta.RESTMapper, error)
	
	NewClient NewClientFunc
	
	NewCache cache.NewCacheFunc
	
	// Readiness probe endpoint name, defaults to "readyz"
	ReadinessEndpointName string
	
	// Liveness probe endpoint name, defaults to "healthz"
	LivenessEndpointName string
	
	newMetricsListener     func(addr string) (net.Listener, error)
	newHealthProbeListener func(addr string) (net.Listener, error)
	
	GracefulShutdownTimeout *time.Duration
	
	Logger logr.Logger
}

type NewClientFunc func(cache cache.Cache, config *rest.Config, options client.Options) (client.Client, error)

func DefaultNewClient(cache cache.Cache, config *rest.Config, options client.Options) (client.Client, error) {
	c, err := client.New(config, options)
	if err != nil {
		return nil, err
	}
	
	
}

func setOptionsDefaults(options Options) Options  {
	if options.Scheme == nil {
		options.Scheme = scheme.Scheme
	}
	
	if options.MapperProvider == nil {
		options.MapperProvider = func(config *rest.Config) (meta.RESTMapper, error) {
			return apiutil.NewDynamicRESTMapper(config)
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

func New(config *rest.Config, options Options) (Manager, error)  {
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
	
	
	
	return &controllerManager{
		config:                  config,
		scheme:                  options.Scheme,
		cache:                   cache,
		fieldIndexes:            cache,
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


