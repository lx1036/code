package client

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"strings"
)

type unstructuredClient struct {
	cache      *clientCache
	paramCodec runtime.ParameterCodec
}

func (uc *unstructuredClient) Get(ctx context.Context, key ObjectKey, obj Object) error {
	u, ok := obj.(*unstructured.Unstructured)
	log.WithFields(log.Fields{
		"ok": ok,
	}).Debug("[unstructuredClient List]")
	if !ok {
		return fmt.Errorf("unstructured client did not understand object: %T", obj)
	}

	gvk := u.GroupVersionKind()
	r, err := uc.cache.getResource(obj)
	if err != nil {
		return err
	}

	result := r.Get().
		NamespaceIfScoped(key.Namespace, r.isNamespaced()).
		Resource(r.resource()).
		Name(key.Name).
		Do(ctx).
		Into(obj)

	u.SetGroupVersionKind(gvk)

	return result
}

func (uc *unstructuredClient) List(ctx context.Context, obj runtime.Object, opts ...ListOption) error {
	u, ok := obj.(*unstructured.UnstructuredList)
	log.WithFields(log.Fields{
		"ok": ok,
	}).Debug("[unstructuredClient List]")
	if !ok {
		return fmt.Errorf("unstructured client did not understand object: %T", obj)
	}

	gvk := u.GroupVersionKind()
	if strings.HasSuffix(gvk.Kind, "List") {
		gvk.Kind = gvk.Kind[:len(gvk.Kind)-4]
	}

	r, err := uc.cache.getResource(obj)
	if err != nil {
		return err
	}

	listOpts := ListOptions{}
	listOpts.ApplyOptions(opts)

	return r.Get().
		NamespaceIfScoped(listOpts.Namespace, r.isNamespaced()).
		Resource(r.resource()).
		VersionedParams(listOpts.AsListOptions(), uc.paramCodec).
		Do(ctx).
		Into(obj)
}

func (uc *unstructuredClient) Create(ctx context.Context, obj runtime.Object, opts ...CreateOption) error {
	u, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("unstructured client did not understand object: %T", obj)
	}

	gvk := u.GroupVersionKind()

	objMeta, err := uc.cache.getObjMeta(obj)
	if err != nil {
		return err
	}

	createOpts := &CreateOptions{}
	createOpts.ApplyOptions(opts)
	result := objMeta.Post().
		NamespaceIfScoped(objMeta.GetNamespace(), objMeta.isNamespaced()).
		Resource(objMeta.resource()).
		Body(obj).
		VersionedParams(createOpts.AsCreateOptions(), uc.paramCodec).
		Do(ctx).
		Into(obj)

	u.SetGroupVersionKind(gvk)
	return result
}
func (uc *unstructuredClient) Update(ctx context.Context, obj runtime.Object, opts ...UpdateOption) error {
	u, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("unstructured client did not understand object: %T", obj)
	}

	gvk := u.GroupVersionKind()

	o, err := uc.cache.getObjMeta(obj)
	if err != nil {
		return err
	}

	updateOpts := UpdateOptions{}
	updateOpts.ApplyOptions(opts)
	result := o.Put().
		NamespaceIfScoped(o.GetNamespace(), o.isNamespaced()).
		Resource(o.resource()).
		Name(o.GetName()).
		Body(obj).
		VersionedParams(updateOpts.AsUpdateOptions(), uc.paramCodec).
		Do(ctx).
		Into(obj)

	u.SetGroupVersionKind(gvk)
	return result
}
func (uc *unstructuredClient) UpdateStatus(ctx context.Context, obj runtime.Object, opts ...UpdateOption) error {
	_, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("unstructured client did not understand object: %T", obj)
	}

	o, err := uc.cache.getObjMeta(obj)
	if err != nil {
		return err
	}

	return o.Put().
		NamespaceIfScoped(o.GetNamespace(), o.isNamespaced()).
		Resource(o.resource()).
		Name(o.GetName()).
		SubResource("status").
		Body(obj).
		VersionedParams((&UpdateOptions{}).ApplyOptions(opts).AsUpdateOptions(), uc.paramCodec).
		Do(ctx).
		Into(obj)
}
func (uc *unstructuredClient) Patch(ctx context.Context, obj runtime.Object, patch Patch, opts ...PatchOption) error {
	_, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("unstructured client did not understand object: %T", obj)
	}

	o, err := uc.cache.getObjMeta(obj)
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
		VersionedParams(patchOpts.ApplyOptions(opts).AsPatchOptions(), uc.paramCodec).
		Body(data).
		Do(ctx).
		Into(obj)
}
func (uc *unstructuredClient) PatchStatus(ctx context.Context, obj runtime.Object, patch Patch, opts ...PatchOption) error {
	u, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("unstructured client did not understand object: %T", obj)
	}

	gvk := u.GroupVersionKind()

	o, err := uc.cache.getObjMeta(obj)
	if err != nil {
		return err
	}

	data, err := patch.Data(obj)
	if err != nil {
		return err
	}

	patchOpts := &PatchOptions{}
	result := o.Patch(patch.Type()).
		NamespaceIfScoped(o.GetNamespace(), o.isNamespaced()).
		Resource(o.resource()).
		Name(o.GetName()).
		SubResource("status").
		Body(data).
		VersionedParams(patchOpts.ApplyOptions(opts).AsPatchOptions(), uc.paramCodec).
		Do(ctx).
		Into(u)

	u.SetGroupVersionKind(gvk)
	return result
}
func (uc *unstructuredClient) Delete(ctx context.Context, obj runtime.Object, opts ...DeleteOption) error {
	_, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("unstructured client did not understand object: %T", obj)
	}

	o, err := uc.cache.getObjMeta(obj)
	if err != nil {
		return err
	}

	deleteOpts := DeleteOptions{}
	deleteOpts.ApplyOptions(opts)
	return o.Delete().
		NamespaceIfScoped(o.GetNamespace(), o.isNamespaced()).
		Resource(o.resource()).
		Name(o.GetName()).
		Body(deleteOpts.AsDeleteOptions()).
		Do(ctx).
		Error()
}
func (uc *unstructuredClient) DeleteAllOf(ctx context.Context, obj runtime.Object, opts ...DeleteAllOfOption) error {
	_, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("unstructured client did not understand object: %T", obj)
	}

	o, err := uc.cache.getObjMeta(obj)
	if err != nil {
		return err
	}

	deleteAllOfOpts := DeleteAllOfOptions{}
	deleteAllOfOpts.ApplyOptions(opts)
	return o.Delete().
		NamespaceIfScoped(deleteAllOfOpts.ListOptions.Namespace, o.isNamespaced()).
		Resource(o.resource()).
		VersionedParams(deleteAllOfOpts.AsListOptions(), uc.paramCodec).
		Body(deleteAllOfOpts.AsDeleteOptions()).
		Do(ctx).
		Error()
}
