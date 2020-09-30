package cache

import "k8s.io/client-go/rest"

type NewCacheFunc func(config *rest.Config, opts Options) (Cache, error)

