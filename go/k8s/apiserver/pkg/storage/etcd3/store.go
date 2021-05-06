package etcd3

import (
	"context"
	"path"

	"k8s-lx1036/k8s/apiserver/pkg/storage"
	"k8s-lx1036/k8s/apiserver/pkg/storage/value"

	"go.etcd.io/etcd/clientv3"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
)

type store struct {
	client        *clientv3.Client
	codec         runtime.Codec
	versioner     storage.Versioner
	transformer   value.Transformer
	pathPrefix    string
	watcher       *watcher
	pagingEnabled bool
	leaseManager  *leaseManager
}

func (s *store) Create(ctx context.Context, key string, obj, out runtime.Object, ttl uint64) error {
	panic("implement me")
}

func (s *store) Delete(ctx context.Context, key string, out runtime.Object, preconditions *storage.Preconditions, validateDeletion storage.ValidateObjectFunc) error {
	panic("implement me")
}

func (s *store) Watch(ctx context.Context, key string, opts storage.ListOptions) (watch.Interface, error) {
	panic("implement me")
}

func (s *store) WatchList(ctx context.Context, key string, opts storage.ListOptions) (watch.Interface, error) {
	panic("implement me")
}

func (s *store) Get(ctx context.Context, key string, opts storage.GetOptions, objPtr runtime.Object) error {
	panic("implement me")
}

func (s *store) List(ctx context.Context, key string, opts storage.ListOptions, listObj runtime.Object) error {
	panic("implement me")
}

// New returns an etcd3 implementation of storage.Interface.
func New(c *clientv3.Client, codec runtime.Codec, prefix string, transformer value.Transformer, pagingEnabled bool,
	leaseManagerConfig LeaseManagerConfig) storage.Interface {
	return newStore(c, pagingEnabled, codec, prefix, transformer, leaseManagerConfig)
}

func newStore(c *clientv3.Client, pagingEnabled bool, codec runtime.Codec, prefix string,
	transformer value.Transformer, leaseManagerConfig LeaseManagerConfig) *store {
	versioner := APIObjectVersioner{}
	result := &store{
		client:        c,
		codec:         codec,
		versioner:     versioner,
		transformer:   transformer,
		pagingEnabled: pagingEnabled,
		// for compatibility with etcd2 impl.
		// no-op for default prefix of '/registry'.
		// keeps compatibility with etcd2 impl for custom prefixes that don't start with '/'
		pathPrefix:   path.Join("/", prefix),
		watcher:      newWatcher(c, codec, versioner, transformer),
		leaseManager: newDefaultLeaseManager(c, leaseManagerConfig),
	}

	return result
}
