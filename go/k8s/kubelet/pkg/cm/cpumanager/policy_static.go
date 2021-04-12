package cpumanager

import (
	"fmt"

	"k8s-lx1036/k8s/kubelet/pkg/cm/cpumanager/state"
	"k8s-lx1036/k8s/kubelet/pkg/cm/cpumanager/topology"
	"k8s-lx1036/k8s/kubelet/pkg/cm/cpuset"
	"k8s-lx1036/k8s/kubelet/pkg/cm/topologymanager"

	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

const PolicyStatic policyName = "static"

// staticPolicy 只对 Guaranteed 且 request 是整数的 Pod 才有效
// @see https://kubernetes.io/docs/tasks/administer-cluster/cpu-management-policies/#static-policy
// - The pod QoS class is Guaranteed.
// - The CPU request is a positive integer.
type staticPolicy struct {
	// cpu socket topology
	topology *topology.CPUTopology
	// set of CPUs that is not available for exclusive assignment
	reserved cpuset.CPUSet
	// topology manager reference to get container Topology affinity
	affinity topologymanager.Store
	// set of CPUs to reuse across allocations in a pod
	cpusToReuse map[string]cpuset.CPUSet
}

func (policy *staticPolicy) Name() string {
	panic("implement me")
}

func (policy *staticPolicy) Start(s state.State) error {
	panic("implement me")
}

func (policy *staticPolicy) Allocate(s state.State, pod *v1.Pod, container *v1.Container) error {
	panic("implement me")
}

func (policy *staticPolicy) RemoveContainer(s state.State, podUID string, containerName string) error {
	panic("implement me")
}

func (policy *staticPolicy) GetTopologyHints(s state.State, pod *v1.Pod, container *v1.Container) map[string][]topologymanager.TopologyHint {
	panic("implement me")
}

func NewStaticPolicy(topology *topology.CPUTopology, numReservedCPUs int,
	reservedCPUs cpuset.CPUSet, affinity topologymanager.Store) (Policy, error) {
	//allCPUs := topology.CPUDetails.CPUs()
	var reserved cpuset.CPUSet
	if reservedCPUs.Size() > 0 {
		reserved = reservedCPUs
	} else {
		//reserved, _ = takeByTopology(topology, allCPUs, numReservedCPUs)
	}

	if reserved.Size() != numReservedCPUs {
		err := fmt.Errorf("[cpumanager] unable to reserve the required amount of CPUs "+
			"(size of %s did not equal %d)", reserved, numReservedCPUs)
		return nil, err
	}

	klog.Infof("[cpumanager] reserved %d CPUs (\"%s\") not available for exclusive assignment", reserved.Size(), reserved)

	return &staticPolicy{
		topology:    topology,
		reserved:    reserved,
		affinity:    affinity,
		cpusToReuse: make(map[string]cpuset.CPUSet),
	}, nil
}
