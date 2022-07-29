package queuesort

import (
	framework "k8s-lx1036/k8s/scheduler/pkg/framework"
	"k8s.io/apimachinery/pkg/runtime"

	"k8s.io/kubernetes/pkg/api/v1/pod"
)

// INFO: pod priority 优先级高在前

// PrioritySort is a plugin that implements Priority based sorting.
type PrioritySort struct{}

const Name = "PrioritySort"

// Less is the function used by the activeQ heap algorithm to sort pods.
// It sorts pods based on their priority. When priorities are equal, it uses
// PodQueueInfo.timestamp.
func (pl *PrioritySort) Less(pInfo1, pInfo2 *framework.QueuedPodInfo) bool {
	p1 := pod.GetPodPriority(pInfo1.Pod)
	p2 := pod.GetPodPriority(pInfo2.Pod)

	return (p1 > p2) || (p1 == p2 && pInfo1.Timestamp.Before(pInfo2.Timestamp))
}

// Name returns name of the plugin.
func (pl *PrioritySort) Name() string {
	return Name
}

// New initializes a new plugin and returns it.
func New(_ runtime.Object, handle framework.FrameworkHandle) (framework.Plugin, error) {
	return &PrioritySort{}, nil
}
