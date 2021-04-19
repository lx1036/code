package api

import (
	"k8s.io/apimachinery/pkg/api/resource"
	"time"
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

// ThresholdOperator is the operator used to express a Threshold.
type ThresholdOperator string

const (
	// OpLessThan is the operator that expresses a less than operator.
	OpLessThan ThresholdOperator = "LessThan"
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

// ThresholdValue is a value holder that abstracts literal versus percentage based quantity
type ThresholdValue struct {
	// The following fields are exclusive. Only the topmost non-zero field is used.

	// Quantity is a quantity associated with the signal that is evaluated against the specified operator.
	Quantity *resource.Quantity
	// Percentage represents the usage percentage over the total resource that is evaluated against the specified operator.
	Percentage float32
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

// GetThresholdQuantity returns the expected quantity value for a thresholdValue
func GetThresholdQuantity(value ThresholdValue, capacity *resource.Quantity) *resource.Quantity {
	if value.Quantity != nil {
		res := value.Quantity.DeepCopy()
		return &res
	}
	return resource.NewQuantity(int64(float64(capacity.Value())*float64(value.Percentage)), resource.BinarySI)
}
