package queue

import (
	"k8s-lx1036/k8s/scheduler/pkg/framework"
	"k8s-lx1036/k8s/scheduler/pkg/metrics"

	corev1 "k8s.io/api/core/v1"
)

// UnschedulablePodsMap 存储不可被调度的 pod
type UnschedulablePods struct {
	// podInfoMap is a map key by a pod's full-name and the value is a pointer to the QueuedPodInfo.
	podInfoMap map[string]*framework.QueuedPodInfo

	keyFunc func(*corev1.Pod) string

	metricRecorder metrics.MetricRecorder
}

func newUnschedulablePodsMap(metricRecorder metrics.MetricRecorder) *UnschedulablePods {
	return &UnschedulablePods{
		podInfoMap:     make(map[string]*framework.QueuedPodInfo),
		keyFunc:        util.GetPodFullName,
		metricRecorder: metricRecorder,
	}
}

func (u *UnschedulablePods) addOrUpdate(pInfo *framework.QueuedPodInfo) {
	u.podInfoMap[u.keyFunc(pInfo.Pod)] = pInfo
}

func (u *UnschedulablePods) get(pod *corev1.Pod) *framework.QueuedPodInfo {
	podKey := u.keyFunc(pod)
	if pInfo, exists := u.podInfoMap[podKey]; exists {
		return pInfo
	}
	return nil
}

func (u *UnschedulablePods) delete(pod *corev1.Pod) {
	podID := u.keyFunc(pod)
	delete(u.podInfoMap, podID)
}
