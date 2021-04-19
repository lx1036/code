package cpumanager

import (
	"k8s-lx1036/k8s/kubelet/pkg/cm/cpumanager/state"
	"k8s-lx1036/k8s/kubelet/pkg/cm/topologymanager"

	v1 "k8s.io/api/core/v1"
)

type policyName string

// Policy implements logic for pod container to CPU assignment.
type Policy interface {
	Name() string
	Start(s state.State) error
	// Allocate call is idempotent
	Allocate(s state.State, pod *v1.Pod, container *v1.Container) error
	// RemoveContainer call is idempotent
	RemoveContainer(s state.State, podUID string, containerName string) error
	// GetTopologyHints implements the topologymanager.HintProvider Interface
	// and is consulted to achieve NUMA aware resource alignment among this
	// and other resource controllers.
	GetTopologyHints(s state.State, pod *v1.Pod, container *v1.Container) map[string][]topologymanager.TopologyHint
}
