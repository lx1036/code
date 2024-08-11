package cache

import (
	"k8s-lx1036/k8s/scheduler/volcano/volcano/pkg/scheduler/api"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

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

// Binder interface for binding task and hostname
type Binder interface {
	Bind(kubeClient kubernetes.Interface, tasks []*api.TaskInfo) ([]*api.TaskInfo, []error)
}

// StatusUpdater updates pod with given PodCondition
type StatusUpdater interface {
	UpdatePodStatus(pod *v1.Pod) (*v1.Pod, error)
	UpdatePodGroup(pg *api.PodGroup) (*api.PodGroup, error)
	UpdateQueueStatus(queue *api.QueueInfo) error
}

// BatchBinder updates podgroup or job information
type BatchBinder interface {
	Bind(job *api.JobInfo, cluster string) (*api.JobInfo, error)
}
