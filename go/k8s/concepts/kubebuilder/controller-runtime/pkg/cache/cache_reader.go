package cache

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"
)

// CacheReader wraps a cache.Index to implement the client.CacheReader interface for a single type
type CacheReader struct {
	// indexer is the underlying indexer wrapped by this cache.
	indexer cache.Indexer

	// groupVersionKind is the group-version-kind of the resource.
	groupVersionKind schema.GroupVersionKind
}
