package eviction

import (
	statsapi "k8s-lx1036/k8s/kubelet/pkg/apis/stats/v1alpha1"

	v1 "k8s.io/api/core/v1"
)

// fsStatsType defines the types of filesystem stats to collect.
type fsStatsType string

const (
	// fsStatsLocalVolumeSource identifies stats for pod local volume sources.
	fsStatsLocalVolumeSource fsStatsType = "localVolumeSource"
	// fsStatsLogs identifies stats for pod logs.
	fsStatsLogs fsStatsType = "logs"
	// fsStatsRoot identifies stats for pod container writable layers.
	fsStatsRoot fsStatsType = "root"
)

// KillPodFunc kills a pod.
// The pod status is updated, and then it is killed with the specified grace period.
// This function must block until either the pod is killed or an error is encountered.
// Arguments:
// pod - the pod to kill
// status - the desired status to associate with the pod (i.e. why its killed)
// gracePeriodOverride - the grace period override to use instead of what is on the pod spec
type KillPodFunc func(pod *v1.Pod, status v1.PodStatus, gracePeriodOverride *int64) error

// nodeReclaimFunc is a function that knows how to reclaim a resource from the node without impacting pods.
type nodeReclaimFunc func() error

// nodeReclaimFuncs is an ordered list of nodeReclaimFunc
type nodeReclaimFuncs []nodeReclaimFunc

// rankFunc sorts the pods in eviction order
type rankFunc func(pods []*v1.Pod, stats statsFunc)

// ImageGC is responsible for performing garbage collection of unused images.
type ImageGC interface {
	// DeleteUnusedImages deletes unused images.
	DeleteUnusedImages() error
}

// ContainerGC is responsible for performing garbage collection of unused containers.
type ContainerGC interface {
	// DeleteAllUnusedContainers deletes all unused containers, even those that belong to pods that are terminated, but not deleted.
	DeleteAllUnusedContainers() error
}

// statsFunc returns the usage stats if known for an input pod.
type statsFunc func(pod *v1.Pod) (statsapi.PodStats, bool)

// ActivePodsFunc returns pods bound to the kubelet that are active (i.e. non-terminal state)
type ActivePodsFunc func() []*v1.Pod
