package cache

import (
	"context"
	"fmt"
	"k8s-lx1036/k8s/concepts/kubebuilder/controller-runtime/pkg/client"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"strings"
)

type ErrCacheNotStarted struct{}

func (*ErrCacheNotStarted) Error() string {
	return "the cache is not started, can not read objects"
}

// informerCache is a Kubernetes Object cache populated from InformersMap.  informerCache wraps an InformersMap.
type informerCache struct {
	*InformersMap
}

func (iCache informerCache) Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
	panic("implement me")
}

func (iCache informerCache) List(ctx context.Context, out runtime.Object, opts ...client.ListOption) error {
	gvk, cacheTypeObj, err := iCache.objectTypeForListObject(out)
	if err != nil {
		return err
	}

	started, cache, err := iCache.InformersMap.Get(ctx, *gvk, cacheTypeObj)
	if err != nil {
		return err
	}

	if !started {
		return &ErrCacheNotStarted{}
	}

	return cache.Reader.List(ctx, out, opts...)
}

// objectTypeForListObject tries to find the runtime.Object and associated GVK
// for a single object corresponding to the passed-in list type. We need them
// because they are used as cache map key.
func (iCache *informerCache) objectTypeForListObject(list runtime.Object) (*schema.GroupVersionKind, runtime.Object, error) {
	gvk, err := client.GVKForObject(list, iCache.Scheme)
	if err != nil {
		return nil, nil, err
	}

	if !strings.HasSuffix(gvk.Kind, "List") {
		return nil, nil, fmt.Errorf("non-list type %T (kind %q) passed as output", list, gvk)
	}

}

func (iCache informerCache) GetInformer(ctx context.Context, obj runtime.Object) (interface{}, error) {
	panic("implement me")
}

func (iCache informerCache) GetInformerForKind(ctx context.Context, gvk schema.GroupVersionKind) (interface{}, error) {
	panic("implement me")
}

func (iCache informerCache) Start(stopCh <-chan struct{}) error {
	panic("implement me")
}

func (iCache informerCache) IndexField(ctx context.Context, obj runtime.Object, field string, extractValue client.IndexerFunc) error {
	panic("implement me")
}
