package client

import (
	"sync"
	"time"
)

const (
	// the expiration duration of the dentry in the cache (used internally)
	DentryValidDuration = 10 * time.Second
)

// INFO: linux 中 dentry struct 包含 inodeID 信息，见 dentry_inode.png
type Dentry struct {
	inodeID    uint64
	expiration int64 // second
}

type DentryCache struct {
	sync.Mutex

	// from name to inodeID
	cache map[string]*Dentry

	//from inodeID to name
	inodeIDCache map[uint64]string
}

func (dentryCache *DentryCache) Put(name string, inodeID uint64) {
	dentryCache.Lock()
	defer dentryCache.Unlock()

	dentry := &Dentry{
		inodeID:    inodeID,
		expiration: time.Now().Add(DentryValidDuration).Unix(), // second
	}
	dentryCache.cache[name] = dentry
	dentryCache.inodeIDCache[inodeID] = name
}

// NewDentryCache returns a new dentry cache.
func NewDentryCache() *DentryCache {
	return &DentryCache{
		cache:        make(map[string]*Dentry),
		inodeIDCache: make(map[uint64]string),
	}
}
