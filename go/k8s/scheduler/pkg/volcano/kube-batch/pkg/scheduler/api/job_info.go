package api

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// JobInfo will have all info of a Job
type JobInfo struct {
	UID JobID

	Name      string
	Namespace string

	Queue QueueID

	Priority int32

	NodeSelector map[string]string
	MinAvailable int32

	NodesFitDelta NodeResourceMap

	// All tasks of the Job.
	TaskStatusIndex map[TaskStatus]tasksMap
	Tasks           tasksMap

	Allocated    *Resource
	TotalRequest *Resource

	CreationTimestamp metav1.Time
	PodGroup          *PodGroup
}
