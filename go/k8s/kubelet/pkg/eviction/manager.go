package eviction

import (
	"k8s.io/apimachinery/pkg/util/wait"
	"time"

	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/kubelet/util/format"
)

type Manager interface {
}

type managerImpl struct {
}

// Start starts the control loop to observe and response to low compute resources.
func (m *managerImpl) Start(diskInfoProvider DiskInfoProvider, podFunc ActivePodsFunc, podCleanedUpFunc PodCleanedUpFunc, monitoringInterval time.Duration) {
	// start the eviction manager monitoring
	go wait.Until(func() {
		if evictedPods := m.synchronize(diskInfoProvider, podFunc); evictedPods != nil {
			klog.Infof("eviction manager: pods %s evicted, waiting for pod to be cleaned up", format.Pods(evictedPods))
			m.waitForPodsCleanup(podCleanedUpFunc, evictedPods)
		}
	}, monitoringInterval, wait.NeverStop)
}

func NewManager() (Manager, Manager) {

}
