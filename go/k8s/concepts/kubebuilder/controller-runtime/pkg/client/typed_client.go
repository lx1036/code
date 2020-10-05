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
	r, err := c.cache.getResource(obj)
	if err != nil {
		return err
	}

	return r.Get().
		NamespaceIfScoped(key.Namespace, r.isNamespaced()).
		Resource(r.resource()).
		Name(key.Name).
		Do(ctx).
		Into(obj)
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

func (c *typedClient) Create(ctx context.Context, obj runtime.Object, opts ...CreateOption) error {
	objMeta, err := c.cache.getObjMeta(obj)
	if err != nil {
		return err
	}

	createOpts := &CreateOptions{}
	createOpts.ApplyOptions(opts)
	return objMeta.Post().
		NamespaceIfScoped(objMeta.GetNamespace(), objMeta.isNamespaced()).
		Resource(objMeta.resource()).
		VersionedParams(createOpts.AsCreateOptions(), c.paramCodec).
		Body(obj).
		Do(ctx).
		Into(obj)
}

// Update implements client.Client
func (c *typedClient) Update(ctx context.Context, obj runtime.Object, opts ...UpdateOption) error {
	o, err := c.cache.getObjMeta(obj)
	if err != nil {
		return err
	}

	updateOpts := &UpdateOptions{}
	updateOpts.ApplyOptions(opts)
	return o.Put().
		NamespaceIfScoped(o.GetNamespace(), o.isNamespaced()).
		Resource(o.resource()).
		Name(o.GetName()).
		VersionedParams(updateOpts.AsUpdateOptions(), c.paramCodec).
		Body(obj).
		Do(ctx).
		Into(obj)
}
func (c *typedClient) UpdateStatus(ctx context.Context, obj runtime.Object, opts ...UpdateOption) error {
	o, err := c.cache.getObjMeta(obj)
	if err != nil {
		return err
	}

	updateOpts := &UpdateOptions{}
	updateOpts.ApplyOptions(opts)
	return o.Put().
		NamespaceIfScoped(o.GetNamespace(), o.isNamespaced()).
		Resource(o.resource()).
		Name(o.GetName()).
		SubResource("status").
		VersionedParams(updateOpts.AsUpdateOptions(), c.paramCodec).
		Body(obj).
		Do(ctx).
		Into(obj)
}
func (c *typedClient) Patch(ctx context.Context, obj runtime.Object, patch Patch, opts ...PatchOption) error {
	o, err := c.cache.getObjMeta(obj)
	if err != nil {
		return err
	}

	data, err := patch.Data(obj)
	if err != nil {
		return err
	}

	patchOpts := &PatchOptions{}
	return o.Patch(patch.Type()).
		NamespaceIfScoped(o.GetNamespace(), o.isNamespaced()).
		Resource(o.resource()).
		Name(o.GetName()).
		VersionedParams(patchOpts.ApplyOptions(opts).AsPatchOptions(), c.paramCodec).
		Body(data).
		Do(ctx).
		Into(obj)
}
func (c *typedClient) PatchStatus(ctx context.Context, obj runtime.Object, patch Patch, opts ...PatchOption) error {
	o, err := c.cache.getObjMeta(obj)
	if err != nil {
		return err
	}

	data, err := patch.Data(obj)
	if err != nil {
		return err
	}

	patchOpts := &PatchOptions{}
	return o.Patch(patch.Type()).
		NamespaceIfScoped(o.GetNamespace(), o.isNamespaced()).
		Resource(o.resource()).
		Name(o.GetName()).
		SubResource("status").
		VersionedParams(patchOpts.ApplyOptions(opts).AsPatchOptions(), c.paramCodec).
		Body(data).
		Do(ctx).
		Into(obj)
}
func (c *typedClient) Delete(ctx context.Context, obj runtime.Object, opts ...DeleteOption) error {
	o, err := c.cache.getObjMeta(obj)
	if err != nil {
		return err
	}

	deleteOpts := DeleteOptions{}
	return o.Delete().
		NamespaceIfScoped(o.GetNamespace(), o.isNamespaced()).
		Resource(o.resource()).
		Name(o.GetName()).
		Body(deleteOpts.ApplyOptions(opts).AsDeleteOptions()).
		Do(ctx).
		Error()
}
func (c *typedClient) DeleteAllOf(ctx context.Context, obj runtime.Object, opts ...DeleteAllOfOption) error {
	o, err := c.cache.getObjMeta(obj)
	if err != nil {
		return err
	}

	deleteAllOfOpts := DeleteAllOfOptions{}
	deleteAllOfOpts.ApplyOptions(opts)

	return o.Delete().
		NamespaceIfScoped(deleteAllOfOpts.ListOptions.Namespace, o.isNamespaced()).
		Resource(o.resource()).
		VersionedParams(deleteAllOfOpts.AsListOptions(), c.paramCodec).
		Body(deleteAllOfOpts.AsDeleteOptions()).
		Do(ctx).
		Error()
}
