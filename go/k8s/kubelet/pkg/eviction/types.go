package eviction

import (
	statsapi "k8s-lx1036/k8s/kubelet/pkg/apis/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
)

// Signal defines a signal that can trigger eviction of pods on a node.
type Signal string

const (
	// SignalMemoryAvailable is memory available (i.e. capacity - workingSet), in bytes.
	SignalMemoryAvailable Signal = "memory.available"
	// SignalNodeFsAvailable is amount of storage available on filesystem that kubelet uses for volumes, daemon logs, etc.
	SignalNodeFsAvailable Signal = "nodefs.available"
	// SignalNodeFsInodesFree is amount of inodes available on filesystem that kubelet uses for volumes, daemon logs, etc.
	SignalNodeFsInodesFree Signal = "nodefs.inodesFree"
	// SignalImageFsAvailable is amount of storage available on filesystem that container runtime uses for storing images and container writable layers.
	SignalImageFsAvailable Signal = "imagefs.available"
	// SignalImageFsInodesFree is amount of inodes available on filesystem that container runtime uses for storing images and container writable layers.
	SignalImageFsInodesFree Signal = "imagefs.inodesFree"
	// SignalAllocatableMemoryAvailable is amount of memory available for pod allocation (i.e. allocatable - workingSet (of pods), in bytes.
	SignalAllocatableMemoryAvailable Signal = "allocatableMemory.available"
	// SignalPIDAvailable is amount of PID available for pod allocation
	SignalPIDAvailable Signal = "pid.available"
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

// OpForSignal maps Signals to ThresholdOperators.
// Today, the only supported operator is "LessThan". This may change in the future,
// for example if "consumed" (as opposed to "available") type signals are added.
// In both cases the directionality of the threshold is implicit to the signal type
// (for a given signal, the decision to evict will be made when crossing the threshold
// from either above or below, never both). There is thus no reason to expose the
// operator in the Kubelet's public API. Instead, we internally map signal types to operators.
var OpForSignal = map[Signal]ThresholdOperator{
	SignalMemoryAvailable:   OpLessThan,
	SignalNodeFsAvailable:   OpLessThan,
	SignalNodeFsInodesFree:  OpLessThan,
	SignalImageFsAvailable:  OpLessThan,
	SignalImageFsInodesFree: OpLessThan,
	SignalPIDAvailable:      OpLessThan,
}

// Config holds information about how eviction is configured.
type Config struct {
	// PressureTransitionPeriod is duration the kubelet has to wait before transitioning out of a pressure condition.
	PressureTransitionPeriod time.Duration
	// Maximum allowed grace period (in seconds) to use when terminating pods in response to a soft eviction threshold being met.
	MaxPodGracePeriodSeconds int64
	// Thresholds define the set of conditions monitored to trigger eviction.
	Thresholds []Threshold
	// KernelMemcgNotification if true will integrate with the kernel memcg notification to determine if memory thresholds are crossed.
	KernelMemcgNotification bool
	// PodCgroupRoot is the cgroup which contains all pods.
	PodCgroupRoot string
}

// Threshold defines a metric for when eviction should occur.
type Threshold struct {
	// Signal defines the entity that was measured.
	Signal Signal
	// Operator represents a relationship of a signal to a value.
	Operator ThresholdOperator
	// Value is the threshold the resource is evaluated against.
	Value ThresholdValue
	// GracePeriod represents the amount of time that a threshold must be met before eviction is triggered.
	GracePeriod time.Duration
	// MinReclaim represents the minimum amount of resource to reclaim if the threshold is met.
	MinReclaim *ThresholdValue
}

// ThresholdOperator is the operator used to express a Threshold.
type ThresholdOperator string

const (
	// OpLessThan is the operator that expresses a less than operator.
	OpLessThan ThresholdOperator = "LessThan"
)

// ThresholdValue is a value holder that abstracts literal versus percentage based quantity
type ThresholdValue struct {
	// The following fields are exclusive. Only the topmost non-zero field is used.

	// Quantity is a quantity associated with the signal that is evaluated against the specified operator.
	Quantity *resource.Quantity
	// Percentage represents the usage percentage over the total resource that is evaluated against the specified operator.
	Percentage float32
}

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
