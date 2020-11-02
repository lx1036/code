package client

import (
	"context"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

type ObjectKey = types.NamespacedName

type IndexerFunc func(runtime.Object) []string

// Reader knows how to read and list Kubernetes objects.
type Reader interface {
	Get(ctx context.Context, key ObjectKey, obj runtime.Object) error

	List(ctx context.Context, list runtime.Object, opts ...ListOption) error
}

type FieldIndexer interface {
	IndexField(ctx context.Context, obj runtime.Object, field string, extractValue IndexerFunc) error
}

type Patch interface {
	Type() types.PatchType

	Data(obj runtime.Object) ([]byte, error)
}

type Writer interface {
	Create(ctx context.Context, obj runtime.Object, opts ...CreateOption) error

	Delete(ctx context.Context, obj runtime.Object, opts ...DeleteOption) error

	Update(ctx context.Context, obj runtime.Object, opts ...UpdateOption) error

	Patch(ctx context.Context, obj runtime.Object, patch Patch, opts ...PatchOption) error

	DeleteAllOf(ctx context.Context, obj runtime.Object, opts ...DeleteAllOfOption) error
}

type Client interface {
	Reader

	Writer

	StatusClient

	Scheme() *runtime.Scheme

	RESTMapper() meta.RESTMapper
}
