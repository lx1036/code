package store

import (
	"context"
	"time"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/kvstore"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/lock"
)

// Configuration is the set of configuration parameters of a shared store.
type Configuration struct {
	// Prefix is the key prefix of the store shared by all keys. The prefix
	// is the unique identification of the store. Multiple collaborators
	// connected to the same kvstore cluster configuring stores with
	// matching prefixes will automatically form a shared store. This
	// parameter is required.
	Prefix string

	// SynchronizationInterval is the interval in which locally owned keys
	// are synchronized with the kvstore. This parameter is optional.
	SynchronizationInterval time.Duration

	// SharedKeyDeleteDelay is the delay before an shared key delete is
	// handled. This parameter is optional
	SharedKeyDeleteDelay time.Duration

	// KeyCreator is called to allocate a Key instance when a new shared
	// key is discovered. This parameter is required.
	KeyCreator KeyCreator

	// Backend is the kvstore to use as a backend. If no backend is
	// specified, kvstore.Client() is being used.
	Backend kvstore.BackendOperations

	// Observer is the observe that will receive events on key mutations
	Observer Observer

	Context context.Context
}

// SharedStore is an instance of a shared store. It is created with
// JoinSharedStore() and released with the SharedStore.Close() function.
type SharedStore struct {
	// conf is a copy of the store configuration. This field is never
	// mutated after JoinSharedStore() so it is safe to access this without
	// a lock.
	conf Configuration

	// name is the name of the shared store. It is derived from the kvstore
	// prefix.
	name string

	// controllerName is the name of the controller used to synchronize
	// with the kvstore. It is derived from the name.
	controllerName string

	// backend is the backend as configured via Configuration
	backend kvstore.BackendOperations

	// mutex protects mutations to localKeys and sharedKeys
	mutex lock.RWMutex

	// localKeys is a map of keys that are owned by the local instance. All
	// local keys are synchronized with the kvstore. This map can be
	// modified with UpdateLocalKey() and DeleteLocalKey().
	localKeys map[string]LocalKey

	// sharedKeys is a map of all keys that either have been discovered
	// from remote collaborators or successfully shared local keys. This
	// map represents the state in the kvstore and is updated based on
	// kvstore events.
	sharedKeys map[string]Key

	kvstoreWatcher *kvstore.Watcher
}
