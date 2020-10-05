package client

import (
	"context"
	"k8s.io/apimachinery/pkg/runtime"
)

type typedClient struct {
	cache      *clientCache
	paramCodec runtime.ParameterCodec
}

func (c *typedClient) Get(ctx context.Context, key ObjectKey, obj Object) error {
	return nil
}

func (c *typedClient) List(ctx context.Context, obj runtime.Object, opts ...ListOption) error {
	r, err := c.cache.getResource(obj)
	if err != nil {
		return err
	}

	listOpts := ListOptions{}
	listOpts.ApplyOptions(opts)

	return r.Get().
		NamespaceIfScoped(listOpts.Namespace, r.isNamespaced()).
		Resource(r.resource()).
		VersionedParams(listOpts.AsListOptions(), c.paramCodec).
		Do(ctx).
		Into(obj)
}
