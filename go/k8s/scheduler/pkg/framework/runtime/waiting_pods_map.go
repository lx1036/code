package runtime

import (
	"sync"
	"time"

	"k8s-lx1036/k8s/scheduler/pkg/framework"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

// WaitingPodsMap a thread-safe map used to maintain pods waiting in the permit phase.
type WaitingPodsMap struct {
	pods map[types.UID]*WaitingPod
	mu   sync.RWMutex
}

// newWaitingPodsMap returns a new waitingPodsMap.
func newWaitingPodsMap() *WaitingPodsMap {
	return &WaitingPodsMap{
		pods: make(map[types.UID]*WaitingPod),
	}
}

func (m *WaitingPodsMap) get(uid types.UID) *WaitingPod {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.pods[uid]
}

// WaitingPod represents a pod waiting in the permit phase.
type WaitingPod struct {
	pod            *v1.Pod
	pendingPlugins map[string]*time.Timer
	s              chan *framework.Status
	mu             sync.RWMutex
}

// Reject declares the waiting pod unschedulable.
func (w *WaitingPod) Reject(pluginName, msg string) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	for _, timer := range w.pendingPlugins {
		timer.Stop()
	}

	// The select clause works as a non-blocking send.
	// If there is no receiver, it's a no-op (default case).
	select {
	case w.s <- framework.NewStatus(framework.Unschedulable, msg).WithFailedPlugin(pluginName):
	default:
	}
}
