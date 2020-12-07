package cache

import (
	"context"
	"fmt"
	"k8s-lx1036/k8s/concepts/tools/kubebuilder/controller-runtime/pkg/client"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

var (
	logger = log.Log.WithName("object-cache")
)

type Options struct {
	Scheme *runtime.Scheme

	// Mapper is the RESTMapper to use for mapping GroupVersionKinds to Resources
	Mapper meta.RESTMapper

	// Resync is the base frequency the informers are resynced.
	// Defaults to defaultResyncTime.
	// A 10 percent jitter will be added to the Resync period between informers
	// So that all informers will not send list requests simultaneously.
	Resync *time.Duration

	// Namespace restricts the cache's ListWatch to the desired namespace
	// Default watches all namespaces
	Namespace string
}

type Cache interface {
	client.Reader

	Informers
}

type Informers interface {
	GetInformer(ctx context.Context, obj runtime.Object) (Informer, error)

	GetInformerForKind(ctx context.Context, gvk schema.GroupVersionKind) (Informer, error)

	Start(stopCh <-chan struct{}) error

	// wait for all cache to sync.
	WaitForCacheSync(stop <-chan struct{}) bool

	client.FieldIndexer
}

var defaultResyncTime = 10 * time.Hour

func setOptionsDefaults(config *rest.Config, options Options) (Options, error) {
	if options.Scheme == nil {
		options.Scheme = scheme.Scheme
	}

	if options.Mapper == nil {
		var err error
		options.Mapper, err = client.NewDiscoveryRESTMapper(config)
		if err != nil {
			logger.WithName("setup").Error(err, "Failed to get API Group-Resources")
			return options, fmt.Errorf("could not create RESTMapper from config")
		}
	}

	// Default the resync period to 10 hours if unset
	if options.Resync == nil {
		options.Resync = &defaultResyncTime
	}
	return options, nil
}

func New(config *rest.Config, options Options) (Cache, error) {
	opts, err := setOptionsDefaults(config, options)
	if err != nil {
		return nil, err
	}

	informersMap := NewInformersMap(config, opts.Scheme, opts.Mapper, *opts.Resync, opts.Namespace)
	return &informerCache{InformersMap: informersMap}, nil
}
