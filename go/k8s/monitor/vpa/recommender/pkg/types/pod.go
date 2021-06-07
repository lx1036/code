package types

import (
	corev1 "k8s.io/api/core/v1"
	"time"
)

// PodID contains information needed to identify a Pod within a cluster.
type PodID struct {
	// Namespaces where the Pod is defined.
	Namespace string
	// PodName is the name of the pod unique within a namespace.
	PodName string
}

// ContainerID contains information needed to identify a Container within a cluster.
type ContainerID struct {
	PodID
	// ContainerName is the name of the container, unique within a pod.
	ContainerName string
}

// ResourceName represents the name of the resource monitored by recommender.
type ResourceName string

// ResourceAmount represents quantity of a certain resource within a container.
// Note this keeps CPU in millicores (which is not a standard unit in APIs)
// and memory in bytes.
// Allowed values are in the range from 0 to MaxResourceAmount.
type ResourceAmount int64

// Resources is a map from resource name to the corresponding ResourceAmount.
type Resources map[ResourceName]ResourceAmount

const (
	// ResourceCPU represents CPU in millicores (1core = 1000millicores).
	ResourceCPU ResourceName = "cpu"
	// ResourceMemory represents memory, in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024).
	ResourceMemory ResourceName = "memory"
	// MaxResourceAmount is the maximum allowed value of resource amount.
	MaxResourceAmount = ResourceAmount(1e14)
)

// ContainerUsageSample is a measure of resource usage of a container over some
// interval.
type ContainerUsageSample struct {
	// Start of the measurement interval.
	MeasureStart time.Time
	// Average CPU usage in cores or memory usage in bytes.
	Usage ResourceAmount
	// CPU or memory request at the time of measurment.
	Request ResourceAmount
	// Which resource is this sample for.
	Resource ResourceName
}

// ContainerUsageSampleWithKey holds a ContainerUsageSample together with the
// ID of the container it belongs to.
type ContainerUsageSampleWithKey struct {
	ContainerUsageSample
	Container ContainerID
}

// String representation of the labels.LabelSet. This is the value returned by
// labelSet.String(). As opposed to the LabelSet object, it can be used as a map key.
type labelSetKey string

// PodState holds runtime information about a single Pod.
type PodState struct {
	// Unique id of the Pod.
	ID PodID
	// Set of labels attached to the Pod.
	labelSetKey labelSetKey
	// Containers that belong to the Pod, keyed by the container name.
	Containers map[string]*ContainerState
	// PodPhase describing current life cycle phase of the Pod.
	Phase corev1.PodPhase
}
