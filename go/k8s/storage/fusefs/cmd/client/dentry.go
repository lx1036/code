package client

import (
	"k8s-lx1036/k8s/storage/fuse/fuseops"
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

func (dentryCache *DentryCache) GetByInode(inodeID fuseops.InodeID) (string, bool) {
	dentryCache.Lock()
	defer dentryCache.Unlock()

	if name, ok := dentryCache.inodeIDCache[uint64(inodeID)]; ok {
		if dentry, exist := dentryCache.cache[name]; exist {
			if dentry.expiration < time.Now().Unix() {
				delete(dentryCache.cache, name)
				delete(dentryCache.inodeIDCache, uint64(inodeID))
				return "", false
			}
			return name, true
		}

		delete(dentryCache.inodeIDCache, uint64(inodeID))
		return "", false
	}

	return "", false
}

// NewDentryCache returns a new dentry cache.
func NewDentryCache() *DentryCache {
	return &DentryCache{
		cache:        make(map[string]*Dentry),
		inodeIDCache: make(map[uint64]string),
	}
}
