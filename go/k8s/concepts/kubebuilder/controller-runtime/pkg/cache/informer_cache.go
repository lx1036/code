package cache

import (
	"context"
	"k8s-lx1036/k8s/concepts/kubebuilder/controller-runtime/pkg/client"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// informerCache is a Kubernetes Object cache populated from InformersMap.  informerCache wraps an InformersMap.
type informerCache struct {
	*InformersMap
}

func (i informerCache) Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
	panic("implement me")
}

func (i informerCache) List(ctx context.Context, list runtime.Object, opts ...client.ListOption) {
	panic("implement me")
}

func (i informerCache) GetInformer(ctx context.Context, obj runtime.Object) (interface{}, error) {
	panic("implement me")
}

func (i informerCache) GetInformerForKind(ctx context.Context, gvk schema.GroupVersionKind) (interface{}, error) {
	panic("implement me")
}

func (i informerCache) Start(stopCh <-chan struct{}) error {
	panic("implement me")
}

func (i informerCache) IndexField(ctx context.Context, obj runtime.Object, field string, extractValue client.IndexerFunc) error {
	panic("implement me")
}
