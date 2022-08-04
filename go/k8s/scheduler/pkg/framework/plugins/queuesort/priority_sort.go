package queuesort

import (
	"k8s-lx1036/k8s/scheduler/pkg/framework"
	frameworkruntime "k8s-lx1036/k8s/scheduler/pkg/framework/runtime"
	"k8s.io/apimachinery/pkg/runtime"

	corev1helpers "k8s.io/component-helpers/scheduling/corev1"
)

// INFO: pod priority 优先级高在前

const Name = "PrioritySort"

// PrioritySort is a plugin that implements Priority based sorting.
type PrioritySort struct{}

// Name returns name of the plugin.
func (pl *PrioritySort) Name() string {
	return Name
}

// New initializes a new plugin and returns it.
func New(_ runtime.Object, fh *frameworkruntime.Framework) (framework.Plugin, error) {
	return &PrioritySort{}, nil
}

func (pl *PrioritySort) Less(pInfo1, pInfo2 *framework.QueuedPodInfo) bool {
	p1 := corev1helpers.PodPriority(pInfo1.Pod)
	p2 := corev1helpers.PodPriority(pInfo2.Pod)

	return (p1 > p2) || (p1 == p2 && pInfo1.Timestamp.Before(pInfo2.Timestamp))
}
