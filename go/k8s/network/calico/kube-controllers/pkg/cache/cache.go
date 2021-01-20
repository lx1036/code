package cache

import (
	"reflect"
	"sync"
	
	log "github.com/sirupsen/logrus"
	"github.com/patrickmn/go-cache"
	
	"k8s.io/client-go/util/workqueue"
)

// ResourceCache stores resources and queues updates when those resources
// are created, modified, or deleted. It de-duplicates updates by ensuring
// updates are only queued when an object has changed.
type ResourceCache interface {
	// Set sets the key to the provided value, and generates an update
	// on the queue the value has changed.
	Set(key string, value interface{})
	
	// Get gets the value associated with the given key.  Returns nil
	// if the key is not present.
	Get(key string) (interface{}, bool)
	
	// Prime sets the key to the provided value, but does not generate
	// and update on the queue ever.
	Prime(key string, value interface{})
	
	// Delete deletes the value identified by the given key from the cache, and
	// generates an update on the queue if a value was deleted.
	Delete(key string)
	
	// Clean removes the object identified by the given key from the cache.
	// It does not generate an update on the queue.
	Clean(key string)
	
	// ListKeys lists the keys currently in the cache.
	ListKeys() []string
	
	// Run enables the generation of events on the output queue starts
	// cache reconciliation.
	Run(reconcilerPeriod string)
	
	// GetQueue returns the cache's output queue, which emits a stream
	// of any keys which have been created, modified, or deleted.
	GetQueue() workqueue.RateLimitingInterface
}


// calicoCache implements the ResourceCache interface
type calicoCache struct {
	threadSafeCache  *cache.Cache
	workqueue        workqueue.RateLimitingInterface
	ListFunc         func() (map[string]interface{}, error)
	ObjectType       reflect.Type
	log              *log.Entry
	running          bool
	mut              *sync.Mutex
	reconcilerConfig ReconcilerConfig
}



// ResourceCacheArgs struct passed to constructor of ResourceCache.
// Groups togather all the arguments to pass in single struct.
type ResourceCacheArgs struct {
	// ListFunc returns a mapping of keys to objects from the Calico datastore.
	ListFunc func() (map[string]interface{}, error)
	
	// ObjectType is the type of object which is to be stored in this cache.
	ObjectType reflect.Type
	
	// LogTypeDesc (optional) to log the type of object stored in the cache.
	// If not provided it is derived from the ObjectType.
	LogTypeDesc string
	
	ReconcilerConfig ReconcilerConfig
}

// NewResourceCache builds and returns a resource cache using the provided arguments.
func NewResourceCache(args ResourceCacheArgs) ResourceCache {
	// Make sure logging is context aware.
	return &calicoCache{
		threadSafeCache: cache.New(cache.NoExpiration, cache.DefaultExpiration),
		workqueue:       workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		ListFunc:        args.ListFunc,
		ObjectType:      args.ObjectType,
		log: func() *log.Entry {
			if args.LogTypeDesc == "" {
				return log.WithFields(log.Fields{"type": args.ObjectType})
			}
			return log.WithFields(log.Fields{"type": args.LogTypeDesc})
		}(),
		mut:              &sync.Mutex{},
		reconcilerConfig: args.ReconcilerConfig,
	}
}


func (c *calicoCache) Set(key string, value interface{}) {
	panic("implement me")
}

func (c *calicoCache) Get(key string) (interface{}, bool) {
	panic("implement me")
}

func (c *calicoCache) Prime(key string, value interface{}) {
	panic("implement me")
}

func (c *calicoCache) Delete(key string) {
	panic("implement me")
}

func (c *calicoCache) Clean(key string) {
	panic("implement me")
}

func (c *calicoCache) ListKeys() []string {
	panic("implement me")
}

func (c *calicoCache) Run(reconcilerPeriod string) {
	panic("implement me")
}

func (c *calicoCache) GetQueue() workqueue.RateLimitingInterface {
	panic("implement me")
}



