package runtime

import (
	"sync"
	"time"

	"k8s-lx1036/k8s/scheduler/pkg/framework"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

// waitingPodsMap a thread-safe map used to maintain pods waiting in the permit phase.
type waitingPodsMap struct {
	pods map[types.UID]*waitingPod
	mu   sync.RWMutex
}

// newWaitingPodsMap returns a new waitingPodsMap.
func newWaitingPodsMap() *waitingPodsMap {
	return &waitingPodsMap{
		pods: make(map[types.UID]*waitingPod),
	}
}

// waitingPod represents a pod waiting in the permit phase.
type waitingPod struct {
	pod            *v1.Pod
	pendingPlugins map[string]*time.Timer
	s              chan *v1alpha1.Status
	mu             sync.RWMutex
}
