package client

import (
	"context"
	"fmt"
	"k8s-lx1036/k8s/concepts/kubebuilder/controller-runtime/pkg/client/apiutil"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

type Object interface {
	metav1.Object
	runtime.Object
}
type ObjectList interface {
	metav1.ListInterface
	runtime.Object
}

type Options struct {
	Scheme *runtime.Scheme

	Mapper meta.RESTMapper
}

// k8s client,直接和api-server通信
type client struct {
	typedClient        typedClient
	unstructuredClient unstructuredClient
	scheme             *runtime.Scheme
	mapper             meta.RESTMapper
}

func (c *client) Get(ctx context.Context, key ObjectKey, obj Object) error {
	_, ok := obj.(*unstructured.Unstructured)
	if ok {
		return c.unstructuredClient.Get(ctx, key, obj)
	}

	return c.typedClient.Get(ctx, key, obj)
}

func (c *client) List(ctx context.Context, obj runtime.Object, opts ...ListOption) error {
	_, ok := obj.(*unstructured.UnstructuredList)
	if ok {
		return c.unstructuredClient.List(ctx, obj, opts...)
	}

	return c.typedClient.List(ctx, obj, opts...)
}

func (c *client) Create(ctx context.Context, obj runtime.Object, opts ...CreateOption) error {
	panic("implement me")
}

func (c *client) Delete(ctx context.Context, obj runtime.Object, opts ...DeleteOption) error {
	panic("implement me")
}

func (c *client) Update(ctx context.Context, obj runtime.Object, opts ...UpdateOption) error {
	panic("implement me")
}

func (c *client) Patch(ctx context.Context, obj runtime.Object, patch Patch, opts ...PatchOption) error {
	panic("implement me")
}

func (c *client) DeleteAllOf(ctx context.Context, obj runtime.Object, opts ...DeleteAllOfOption) error {
	panic("implement me")
}

func (c *client) Status() StatusWriter {
	panic("implement me")
}

func (c *client) Scheme() *runtime.Scheme {
	panic("implement me")
}

func (c *client) RESTMapper() meta.RESTMapper {
	panic("implement me")
}

func New(config *rest.Config, options Options) (Client, error) {
	if config == nil {
		return nil, fmt.Errorf("must provide non-nil rest.Config to client.New")
	}
	if options.Scheme == nil {
		options.Scheme = scheme.Scheme
	}

	// Init a Mapper if none provided
	if options.Mapper == nil {
		var err error
		options.Mapper, err = apiutil.NewDynamicRESTMapper(config)
		if err != nil {
			return nil, err
		}
	}

	clientcache := &clientCache{
		config:         config,
		scheme:         options.Scheme,
		mapper:         options.Mapper,
		codecs:         serializer.NewCodecFactory(options.Scheme),
		resourceByType: make(map[schema.GroupVersionKind]*resourceMeta),
	}

	c := &client{
		typedClient: typedClient{
			cache:      clientcache,
			paramCodec: runtime.NewParameterCodec(options.Scheme),
		},
		unstructuredClient: unstructuredClient{
			cache:      clientcache,
			paramCodec: noConversionParamCodec{},
		},
		scheme: options.Scheme,
		mapper: options.Mapper,
	}

	return c, nil
}
