package client

import (
	"k8s-lx1036/k8s-ui/backend/client/api"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
)

type ResourceHandler interface {
	Create(kind string, namespace string, object *runtime.Unknown) (*runtime.Unknown, error)
	Update(kind string, namespace string, name string, object *runtime.Unknown) (*runtime.Unknown, error)
	Get(kind string, namespace string, name string) (runtime.Object, error)
	List(kind string, namespace string, labelSelector string) ([]runtime.Object, error)
	Delete(kind string, namespace string, name string, options *meta_v1.DeleteOptions) error
}

type resourceHandler struct {
	client       *kubernetes.Clientset
	cacheFactory *CacheFactory
}

func (handler *resourceHandler) Create(kind string, namespace string, object *runtime.Unknown) (*runtime.Unknown, error) {
	panic("implement me")
}

func (handler *resourceHandler) Update(kind string, namespace string, name string, object *runtime.Unknown) (*runtime.Unknown, error) {
	panic("implement me")
}

func (handler *resourceHandler) List(kind string, namespace string, labelSelector string) ([]runtime.Object, error) {
	panic("implement me")
}

func (handler *resourceHandler) Delete(kind string, namespace string, name string, options *meta_v1.DeleteOptions) error {
	panic("implement me")
}

func (handler *resourceHandler) Get(kind string, namespace string, name string) (runtime.Object, error) {
	resource, ok := api.KindToResourceMap[kind]
	if !ok {

	}

	genericInformer, err := handler.cacheFactory.sharedInformerFactory.ForResource(resource.GroupVersionResourceKind.GroupVersionResource)
	if err != nil {
		return nil, err
	}

	var result runtime.Object
	lister := genericInformer.Lister()
	if resource.Namespaced {
		result, err = lister.ByNamespace(namespace).Get(name)
		if err != nil {
			return nil, err
		}
	} else {
		result, err = lister.Get(name)
		if err != nil {
			return nil, err
		}
	}

	result.GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{
		Group:   resource.GroupVersionResourceKind.Group,
		Version: resource.GroupVersionResourceKind.Version,
		Kind:    resource.GroupVersionResourceKind.Kind,
	})

	return result, nil
}

func NewResourceHandler(kubeClient *kubernetes.Clientset, cacheFactory *CacheFactory) ResourceHandler {
	return &resourceHandler{
		client:       kubeClient,
		cacheFactory: cacheFactory,
	}
}
