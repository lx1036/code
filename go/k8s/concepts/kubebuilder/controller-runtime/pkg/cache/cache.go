package cache

import (
	"context"
	"k8s.io/client-go/rest"
	"time"

	"k8s-lx1036/k8s/concepts/kubebuilder/controller-runtime/pkg/client"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
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

func setOptionsDefaults(config *rest.Config, options Options) (Options, error) {

}

func New(config *rest.Config, options Options) (Cache, error) {
	opts, err := setOptionsDefaults(config, options)
	if err != nil {
		return nil, err
	}

	im := NewInformersMap(config, opts.Scheme, opts.Mapper, *opts.Resync, opts.Namespace)
	return &informerCache{InformersMap: im}, nil
}
