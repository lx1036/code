package cache

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"time"
	"context"
)

type InformersMap struct {
	structured   *specificInformersMap
	unstructured *specificInformersMap

	// Scheme maps runtime.Objects to GroupVersionKinds
	Scheme *runtime.Scheme
}

func NewInformersMap(
	config *rest.Config,
	scheme *runtime.Scheme,
	mapper meta.RESTMapper,
	resync time.Duration,
	namespace string) *InformersMap {
	return &InformersMap{
		structured:   newStructuredInformersMap(config, scheme, mapper, resync, namespace),
		unstructured: newUnstructuredInformersMap(config, scheme, mapper, resync, namespace),
		Scheme: scheme,
	}
}

func (m *InformersMap) WaitForCacheSync(stop <-chan struct{}) bool {

	return cache.WaitForCacheSync(stop, syncedFuncs...)
}
func (m *InformersMap) Get(ctx context.Context, gvk schema.GroupVersionKind, obj runtime.Object) (bool, *MapEntry, error) {

}


// newStructuredInformersMap creates a new InformersMap for structured objects.
func newStructuredInformersMap(config *rest.Config, scheme *runtime.Scheme, mapper meta.RESTMapper, resync time.Duration, namespace string) *specificInformersMap {
	return newSpecificInformersMap(config, scheme, mapper, resync, namespace, createStructuredListWatch)
}

// newUnstructuredInformersMap creates a new InformersMap for unstructured objects.
func newUnstructuredInformersMap(config *rest.Config, scheme *runtime.Scheme, mapper meta.RESTMapper, resync time.Duration, namespace string) *specificInformersMap {
	return newSpecificInformersMap(config, scheme, mapper, resync, namespace, createUnstructuredListWatch)
}
