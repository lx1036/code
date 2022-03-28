package client

import (
	"k8s-lx1036/k8s/storage/fuse/fuseops"
	"sync"
	"time"
)

const (
	// the expiration duration of the dentry in the cache (used internally)
	DentryValidDuration = 100 * time.Second
)

// Dentry
// INFO: linux 中 dentry struct 包含 inodeID 信息，见 dentry_inode.png
//  dentry 是 file name 和 inode 的映射，dentry struct 包含 name 和 inode
type Dentry struct {
	inodeID    uint64
	expiration int64 // second
}

// DentryCache INFO: dentry 就是保存 file/dir name 和其 inode 的映射
type DentryCache struct {
	sync.Mutex

	// from name to inodeID
	cache map[string]*Dentry

	//from inodeID to name
	inodeIDCache map[uint64]string
}

func NewDentryCache() *DentryCache {
	return &DentryCache{
		cache:        make(map[string]*Dentry),
		inodeIDCache: make(map[uint64]string),
	}
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

func (dentryCache *DentryCache) Get(name string) (uint64, bool) {
	dentryCache.Lock()
	defer dentryCache.Unlock()

	if dentry, ok := dentryCache.cache[name]; ok {
		if dentry.expiration < time.Now().Unix() {
			delete(dentryCache.cache, name)
			delete(dentryCache.inodeIDCache, dentry.inodeID)
			return 0, false
		}
		return dentry.inodeID, true
	}

	return 0, false
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
