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

// static policy 启动是为了验证cpu assignment相关数据
func (policy *staticPolicy) Start(s state.State) error {
	if err := policy.validateState(s); err != nil {
		klog.Errorf("[cpumanager] static policy invalid state: %v, please drain node and remove policy state file", err)
		return err
	}
	return nil
}

func (policy *staticPolicy) validateState(s state.State) error {
	tmpAssignments := s.GetCPUAssignments()
	tmpDefaultCPUset := s.GetDefaultCPUSet()

	// Default cpuset cannot be empty when assignments exist
	if tmpDefaultCPUset.IsEmpty() {
		if len(tmpAssignments) != 0 {
			return fmt.Errorf("default cpuset cannot be empty")
		}
		// state is empty initialize
		allCPUs := policy.topology.CPUDetails.CPUs()
		s.SetDefaultCPUSet(allCPUs)
		return nil
	}

	// State has already been initialized from file (is not empty)
	// 1. Check if the reserved cpuset is not part of default cpuset because:
	// - kube/system reserved have changed (increased) - may lead to some containers not being able to start
	// - user tampered with file
	// policy.reserved 必须是 tmpDefaultCPUset 的子集
	if !policy.reserved.Intersection(tmpDefaultCPUset).Equals(policy.reserved) {
		return fmt.Errorf("not all reserved cpus: %s are present in defaultCpuSet: %s",
			policy.reserved.String(), tmpDefaultCPUset.String())
	}

	// 2. Check if state for static policy is consistent
	for pod := range tmpAssignments {
		for container, cset := range tmpAssignments[pod] {
			// None of the cpu in DEFAULT cset should be in s.assignments
			if !tmpDefaultCPUset.Intersection(cset).IsEmpty() {
				return fmt.Errorf("pod: %s, container: %s cpuset: %s overlaps with default cpuset %s",
					pod, container, cset.String(), tmpDefaultCPUset.String())
			}
		}
	}

	// 3. It's possible that the set of available CPUs has changed since
	// the state was written. This can be due to for example
	// offlining a CPU when kubelet is not running. If this happens,
	// CPU manager will run into trouble when later it tries to
	// assign non-existent CPUs to containers. Validate that the
	// topology that was received during CPU manager startup matches with
	// the set of CPUs stored in the state.
	totalKnownCPUs := tmpDefaultCPUset.Clone()
	tmpCPUSets := []cpuset.CPUSet{}
	for pod := range tmpAssignments {
		for _, cset := range tmpAssignments[pod] {
			tmpCPUSets = append(tmpCPUSets, cset)
		}
	}
	// union(default + assignment) == total_cpu
	totalKnownCPUs = totalKnownCPUs.UnionAll(tmpCPUSets)
	if !totalKnownCPUs.Equals(policy.topology.CPUDetails.CPUs()) {
		return fmt.Errorf("current set of available CPUs %s doesn't match with CPUs in state %s",
			policy.topology.CPUDetails.CPUs().String(), totalKnownCPUs.String())
	}

	return nil
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
	allCPUs := topology.CPUDetails.CPUs()
	var reserved cpuset.CPUSet
	if reservedCPUs.Size() > 0 {
		reserved = reservedCPUs
	} else {
		// takeByTopology allocates CPUs associated with low-numbered cores from
		// allCPUs.
		//
		// For example: Given a system with 8 CPUs available and HT enabled,
		// if numReservedCPUs=2, then reserved={0,4}
		reserved, _ = takeByTopology(topology, allCPUs, numReservedCPUs)
	}

	// reserved cpu 数量得一致
	if reserved.Size() != numReservedCPUs {
		err := fmt.Errorf("[cpumanager] unable to reserve the required amount of CPUs "+
			"(size of %s did not equal %d)", reserved, numReservedCPUs)
		return nil, err
	}

	klog.Infof("[cpumanager] reserved %d CPUs (%s) not available for exclusive assignment", reserved.Size(), reserved)

	return &staticPolicy{
		topology:    topology,
		reserved:    reserved,
		affinity:    affinity,
		cpusToReuse: make(map[string]cpuset.CPUSet),
	}, nil
}
