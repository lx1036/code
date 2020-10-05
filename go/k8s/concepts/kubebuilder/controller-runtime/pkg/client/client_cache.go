package client

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
	"sync"
)

type clientCache struct {
	// config is the rest.Config to talk to an apiserver
	config *rest.Config

	// scheme maps go structs to GroupVersionKinds
	scheme *runtime.Scheme

	// mapper maps GroupVersionKinds to Resources
	mapper meta.RESTMapper

	// codecs are used to create a REST client for a gvk
	codecs serializer.CodecFactory

	// resourceByType caches type metadata
	resourceByType map[schema.GroupVersionKind]*resourceMeta
	mu             sync.RWMutex
}

type resourceMeta struct {
	rest.Interface

	gvk schema.GroupVersionKind

	mapping *meta.RESTMapping
}
