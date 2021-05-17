package cache

import (
	expirationcache "k8s.io/client-go/tools/cache"
)

// ObjectCache is a simple wrapper of expiration cache that
// 1. use string type key
// 2. has an updater to get value directly if it is expired
// 3. then update the cache
type ObjectCache struct {
	cache   expirationcache.Store
	updater func() (interface{}, error)
}
