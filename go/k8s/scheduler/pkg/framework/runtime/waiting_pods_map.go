package runtime

import (
	"fmt"
	"sync"
	"time"

	"k8s-lx1036/k8s/scheduler/pkg/framework"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

// WaitingPodsMap a thread-safe map used to maintain pods waiting in the permit phase.
type WaitingPodsMap struct {
	mu   sync.RWMutex
	pods map[types.UID]*WaitingPod
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
func (m *WaitingPodsMap) add(wp *WaitingPod) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pods[wp.GetPod().UID] = wp
}
func (m *WaitingPodsMap) remove(uid types.UID) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.pods, uid)
}

// WaitingPod represents a pod waiting in the permit phase.
type WaitingPod struct {
	mu             sync.RWMutex
	pod            *v1.Pod
	pendingPlugins map[string]*time.Timer
	status         chan *framework.Status
}

func newWaitingPod(pod *v1.Pod, pluginsMaxWaitTime map[string]time.Duration) *WaitingPod {
	wp := &WaitingPod{
		pod: pod,
		// Allow() and Reject() calls are non-blocking. This property is guaranteed
		// by using non-blocking send to this channel. This channel has a buffer of size 1
		// to ensure that non-blocking send will not be ignored - possible situation when
		// receiving from this channel happens after non-blocking send.
		status:         make(chan *framework.Status, 1),
		pendingPlugins: make(map[string]*time.Timer, len(pluginsMaxWaitTime)),
	}

	wp.mu.Lock()
	defer wp.mu.Unlock()
	for k, v := range pluginsMaxWaitTime {
		plugin, waitTime := k, v
		wp.pendingPlugins[plugin] = time.AfterFunc(waitTime, func() {
			msg := fmt.Sprintf("rejected due to timeout after waiting %v at plugin %v", waitTime, plugin)
			wp.Reject(plugin, msg)
		})
	}

	return wp
}

func (w *WaitingPod) GetPod() *v1.Pod {
	return w.pod
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
	case w.status <- framework.NewStatus(framework.Unschedulable, msg).WithFailedPlugin(pluginName):
	default:
	}
}
