package client

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
	"strings"
	"sync"
)

// cache rest client 和 meta
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

// getResource returns the resource meta information for the given type of object.
// If the object is a list, the resource represents the item's type instead.
func (c *clientCache) getResource(obj runtime.Object) (*resourceMeta, error) {
	gvk, err := GVKForObject(obj, c.scheme)
	if err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"GVK": fmt.Sprintf("%s/%s/%s", gvk.Group, gvk.Version, gvk.Kind),
	}).Debug("[GVK]")

	// It's better to do creation work twice than to not let multiple
	// people make requests at once
	c.mu.RLock()
	r, known := c.resourceByType[gvk]
	c.mu.RUnlock()

	if known {
		return r, nil
	}

	// Initialize a new Client
	c.mu.Lock()
	defer c.mu.Unlock()
	r, err = c.newResource(gvk, meta.IsListType(obj))
	if err != nil {
		return nil, err
	}
	c.resourceByType[gvk] = r
	return r, err
}

// objMeta stores type and object information about a Kubernetes type
type objMeta struct {
	// resourceMeta contains type information for the object
	*resourceMeta

	// Object contains meta data for the object instance
	metav1.Object
}

// getObjMeta returns objMeta containing both type and object metadata and state
func (c *clientCache) getObjMeta(obj runtime.Object) (*objMeta, error) {
	r, err := c.getResource(obj)
	if err != nil {
		return nil, err
	}

	m, err := meta.Accessor(obj)
	if err != nil {
		return nil, err
	}
	return &objMeta{resourceMeta: r, Object: m}, err
}

// newResource maps obj to a Kubernetes Resource and constructs a client for that Resource.
// If the object is a list, the resource represents the item's type instead.
func (c *clientCache) newResource(gvk schema.GroupVersionKind, isList bool) (*resourceMeta, error) {
	if strings.HasSuffix(gvk.Kind, "List") && isList {
		// if this was a list, treat it as a request for the item's resource
		gvk.Kind = gvk.Kind[:len(gvk.Kind)-4]
	}

	client, err := RESTClientForGVK(gvk, c.config, c.codecs)
	if err != nil {
		return nil, err
	}
	mapping, err := c.mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, err
	}
	return &resourceMeta{Interface: client, mapping: mapping, gvk: gvk}, nil
}

type resourceMeta struct {
	rest.Interface

	gvk schema.GroupVersionKind

	// mapping is the rest mapping
	mapping *meta.RESTMapping
}

// isNamespaced returns true if the type is namespaced
func (r *resourceMeta) isNamespaced() bool {
	return r.mapping.Scope.Name() != meta.RESTScopeNameRoot
}

// resource returns the resource name of the type
func (r *resourceMeta) resource() string {
	return r.mapping.Resource.Resource
}