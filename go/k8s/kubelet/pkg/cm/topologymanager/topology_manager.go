package topologymanager

import "k8s-lx1036/k8s/kubelet/pkg/cm/topologymanager/bitmask"

type TopologyHint struct {
	NUMANodeAffinity bitmask.BitMask

	// Preferred is set to true when the NUMANodeAffinity encodes a preferred
	// allocation for the Container. It is set to false otherwise.
	Preferred bool
}

type Store interface {
	GetAffinity(podUID string, containerName string) TopologyHint
}
