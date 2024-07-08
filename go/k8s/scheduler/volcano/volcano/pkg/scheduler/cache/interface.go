package cache

import "k8s-lx1036/k8s/scheduler/volcano/volcano/pkg/scheduler/api"

// Cache collects pods/nodes/queues information and provides information snapshot
type Cache interface {

	// Snapshot deep copy overall cache information into snapshot
	Snapshot() *api.ClusterInfo

	// SetMetricsConf set the metrics server related configuration
	SetMetricsConf(conf map[string]string)

	// Run start informer
	Run(stopCh <-chan struct{})

	// WaitForCacheSync waits for all cache synced
	WaitForCacheSync(stopCh <-chan struct{})
}
