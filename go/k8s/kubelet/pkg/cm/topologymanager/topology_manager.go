package topologymanager

import (
	"k8s-lx1036/k8s/kubelet/pkg/cm/topologymanager/bitmask"
	"k8s-lx1036/k8s/kubelet/pkg/lifecycle"

	v1 "k8s.io/api/core/v1"
)

type TopologyHint struct {
	NUMANodeAffinity bitmask.BitMask

	// Preferred is set to true when the NUMANodeAffinity encodes a preferred
	// allocation for the Container. It is set to false otherwise.
	Preferred bool
}

type Store interface {
	GetAffinity(podUID string, containerName string) TopologyHint
}

type Manager interface {
	//Manager implements pod admit handler interface
	lifecycle.PodAdmitHandler
	//Adds a hint provider to manager to indicate the hint provider
	//wants to be consoluted when making topology hints
	AddHintProvider(HintProvider)
	//Adds pod to Manager for tracking
	AddContainer(pod *v1.Pod, containerID string) error
	//Removes pod from Manager tracking
	RemoveContainer(containerID string) error
	//Interface for storing pod topology hints
	Store
}

type HintProvider interface {
	GetTopologyHints(pod *v1.Pod, container *v1.Container) map[string][]TopologyHint

	Allocate(pod *v1.Pod, container *v1.Container) error
}
