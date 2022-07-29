package queue

import (
	"k8s-lx1036/k8s/scheduler/pkg/framework"
)

var (
	// UnschedulableTimeout is the event when a pod stays in unschedulable for longer than timeout.
	UnschedulableTimeout = framework.ClusterEvent{Resource: framework.WildCard, ActionType: framework.All, Label: "UnschedulableTimeout"}
	// AssignedPodAdd is the event when a pod is added that causes pods with matching affinity terms to be more schedulable.
	AssignedPodAdd = framework.ClusterEvent{Resource: framework.Pod, ActionType: framework.Add, Label: "AssignedPodAdd"}
)
