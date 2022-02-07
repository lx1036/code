package metadata

import (
	"github.com/containerd/containerd/pkg/timeout"
	"github.com/containerd/containerd/snapshots"
	"os"
	"path/filepath"
	"sync"

	"k8s-lx1036/k8s/kubelet/containerd/containerd/pkg/plugin"
	"k8s-lx1036/k8s/kubelet/containerd/containerd/pkg/services"
	"k8s-lx1036/k8s/kubelet/containerd/containerd/pkg/services/content"

	bolt "go.etcd.io/bbolt"
)

func init() {
	plugin.Register(&plugin.Registration{
		Type:   plugin.MetadataPlugin,
		ID:     services.MetadataService,
		InitFn: initFunc,
	})
}

func initFunc(ic *plugin.InitContext) (interface{}, error) {
	if err := os.MkdirAll(ic.Root, 0711); err != nil {
		return nil, err
	}

	options := *bolt.DefaultOptions
	options.Timeout = timeout.Get(boltOpenTimeout)
	path := filepath.Join(ic.Root, "meta.db")
	db, err := bolt.Open(path, 0644, &options)
	if err != nil {
		return nil, err
	}

	mdb := NewDB(db, cs.(content.Store), snapshotters, dbopts...)
	if err := mdb.Init(ic.Context); err != nil {
		return nil, err
	}
	return mdb, nil
}

// dbOptions configure db options.
type dbOptions struct {
	shared bool
}

// DB represents a metadata database backed by a bolt
// database. The database is fully namespaced and stores
// image, container, namespace, snapshot, and content data
// while proxying data shared across namespaces to backend
// datastores for content and snapshots.
type DB struct {
	// wlock is used to protect access to the data structures during garbage
	// collection. While the wlock is held no writable transactions can be
	// opened, preventing changes from occurring between the mark and
	// sweep phases without preventing read transactions.
	wlock sync.RWMutex

	db *bolt.DB
	ss map[string]*snapshotter
	cs *contentStore

	// dirty flag indicates that references have been removed which require
	// a garbage collection to ensure the database is clean. This tracks
	// the number of dirty operations. This should be updated and read
	// atomically if outside of wlock.Lock.
	dirty uint32
	// dirtySS and dirtyCS flags keeps track of datastores which have had
	// deletions since the last garbage collection. These datastores will
	// be garbage collected during the next garbage collection. These
	// should only be updated inside of a write transaction or wlock.Lock.
	dirtySS map[string]struct{}
	dirtyCS bool

	// mutationCallbacks are called after each mutation with the flag
	// set indicating whether any dirty flags are set
	mutationCallbacks []func(bool)

	dbopts dbOptions
}

// NewDB creates a new metadata database using the provided
// bolt database, content store, and snapshotters.
func NewDB(db *bolt.DB, cs content.Store, ss map[string]snapshots.Snapshotter, opts ...DBOpt) *DB {
	m := &DB{
		db:      db,
		ss:      make(map[string]*snapshotter, len(ss)),
		dirtySS: map[string]struct{}{},
		dbopts: dbOptions{
			shared: true,
		},
	}

	return m
}
