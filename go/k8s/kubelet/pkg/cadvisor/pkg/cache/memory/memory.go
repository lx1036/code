package memory

import (
	"sync"
	"time"

	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/info/v1"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/storage"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/utils"
)

// TODO(vmarmol): See about refactoring this class, we have an unnecessary redirection of containerCache and InMemoryCache.
// containerCache is used to store per-container information
type containerCache struct {
	ref         v1.ContainerReference
	recentStats *utils.TimedStore
	maxAge      time.Duration
	lock        sync.RWMutex
}

type InMemoryCache struct {
	lock              sync.RWMutex
	containerCacheMap map[string]*containerCache
	maxAge            time.Duration
	backend           []storage.StorageDriver
}

func (c *InMemoryCache) RemoveContainer(containerName string) error {
	c.lock.Lock()
	delete(c.containerCacheMap, containerName)
	c.lock.Unlock()

	return nil
}

func New(maxAge time.Duration, backend []storage.StorageDriver) *InMemoryCache {
	ret := &InMemoryCache{
		containerCacheMap: make(map[string]*containerCache, 32),
		maxAge:            maxAge,
		backend:           backend,
	}

	return ret
}
