package cache

import "k8s.io/client-go/rest"

// Cache collects pods/nodes/queues information
// and provides information snapshot
type Cache interface {
}

func New(config *rest.Config, schedulerName string, defaultQueue string) Cache {
	return newSchedulerCache(config, schedulerName, defaultQueue)
}
