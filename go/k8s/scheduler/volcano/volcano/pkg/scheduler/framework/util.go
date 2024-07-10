package framework

import (
	"k8s-lx1036/k8s/scheduler/volcano/volcano/pkg/scheduler/api"

	v1 "k8s.io/api/core/v1"
)

// PodLister is used in predicate and nodeorder plugin
type PodLister struct {
	Session *Session

	CachedPods       map[api.TaskID]*v1.Pod
	Tasks            map[api.TaskID]*api.TaskInfo
	TaskWithAffinity map[api.TaskID]*api.TaskInfo
}
