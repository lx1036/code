package cache

import "k8s-lx1036/k8s/scheduler/volcano/volcano/pkg/scheduler/api"

// Cache collects pods/nodes/queues information and provides information snapshot
type Cache interface {

	// Snapshot deep copy overall cache information into snapshot
	Snapshot() *api.ClusterInfo
}
