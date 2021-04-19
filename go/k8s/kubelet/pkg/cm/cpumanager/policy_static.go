package cpumanager

import (
	"fmt"

	"k8s-lx1036/k8s/kubelet/pkg/cm/cpumanager/state"
	"k8s-lx1036/k8s/kubelet/pkg/cm/cpumanager/topology"
	"k8s-lx1036/k8s/kubelet/pkg/cm/cpuset"
	"k8s-lx1036/k8s/kubelet/pkg/cm/topologymanager"
	"k8s-lx1036/k8s/kubelet/pkg/cm/topologymanager/bitmask"

	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	v1qos "k8s.io/kubernetes/pkg/apis/core/v1/helper/qos"
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

func (policy *staticPolicy) guaranteedCPUs(pod *v1.Pod, container *v1.Container) int {
	if v1qos.GetPodQOS(pod) != v1.PodQOSGuaranteed { // 必须是 guaranteed pod
		return 0
	}

	// cpu request 必须是整数
	cpuQuantity := container.Resources.Requests[v1.ResourceCPU]
	if cpuQuantity.Value()*1000 != cpuQuantity.MilliValue() {
		return 0
	}

	return int(cpuQuantity.Value())
}

// 创建容器时，分配cpu
func (policy *staticPolicy) Allocate(s state.State, pod *v1.Pod, container *v1.Container) error {
	if numCPUs := policy.guaranteedCPUs(pod, container); numCPUs != 0 {
		// INFO: container belongs in an exclusively allocated pool, 容器要绑核独占
		klog.Infof("[cpumanager] static policy: Allocate %d cpus exclusively for (pod: %s, container: %s)",
			numCPUs, pod.Name, container.Name)

		// 如果state里已经有该container.Name的分配情况，重复使用
		if cset, ok := s.GetCPUSet(string(pod.UID), container.Name); ok {
			//policy.updateCPUsToReuse(pod, container, cpuset)
			klog.Infof("[cpumanager] static policy: container already present in state for cpu %s, skipping (pod: %s, container: %s)",
				cset.String(), pod.Name, container.Name)
			return nil
		}

		// Call Topology Manager to get the aligned socket affinity across all hint providers.
		hint := policy.affinity.GetAffinity(string(pod.UID), container.Name)
		klog.Infof("[cpumanager] Pod %v, Container %v Topology Affinity is: %v", pod.UID, container.Name, hint)

		// Allocate CPUs according to the NUMA affinity contained in the hint.
		// INFO: 要给container分配numCPUs cpus
		cset, err := policy.allocateCPUs(s, numCPUs, hint.NUMANodeAffinity, policy.cpusToReuse[string(pod.UID)])
		if err != nil {
			klog.Errorf("[cpumanager] unable to allocate %d CPUs (pod: %s, container: %s, error: %v)", numCPUs, pod.Name, container.Name, err)
			return err
		}

		// 设置哈希表，已经分配出去的 [podUID][container.Name][cpuset]
		s.SetCPUSet(string(pod.UID), container.Name, cset)
		policy.updateCPUsToReuse(pod, container, cset)
	}

	// container belongs in the shared pool (nothing to do; use default cpuset)
	return nil
}

// INFO:
func (policy *staticPolicy) updateCPUsToReuse(pod *v1.Pod, container *v1.Container, cset cpuset.CPUSet) {
	// If pod entries to m.cpusToReuse other than the current pod exist, delete them.
	for podUID := range policy.cpusToReuse {
		if podUID != string(pod.UID) {
			delete(policy.cpusToReuse, podUID)
		}
	}

	// If no cpuset exists for cpusToReuse by this pod yet, create one.
	if _, ok := policy.cpusToReuse[string(pod.UID)]; !ok {
		policy.cpusToReuse[string(pod.UID)] = cpuset.NewCPUSet()
	}

	// Check if the container is an init container.
	// If so, add its cpuset to the cpuset of reusable CPUs for any new allocations.
	for _, initContainer := range pod.Spec.InitContainers {
		if container.Name == initContainer.Name {
			policy.cpusToReuse[string(pod.UID)] = policy.cpusToReuse[string(pod.UID)].Union(cset)
			return
		}
	}

	// Otherwise it is an app container.
	// Remove its cpuset from the cpuset of reusable CPUs for any new allocations.
	policy.cpusToReuse[string(pod.UID)] = policy.cpusToReuse[string(pod.UID)].Difference(cset)
}

// 可分配cpu = unassigned - reserved
func (policy *staticPolicy) assignableCPUs(s state.State) cpuset.CPUSet {
	return s.GetDefaultCPUSet().Difference(policy.reserved)
}

func (policy *staticPolicy) allocateCPUs(s state.State, numCPUs int, numaAffinity bitmask.BitMask, reusableCPUs cpuset.CPUSet) (cpuset.CPUSet, error) {
	klog.Infof("[cpumanager] allocateCpus: (numCPUs: %d, socket: %v)", numCPUs, numaAffinity)

	// 可分配cpu = unassigned - reserved(1,2,3,4,5,6,7), 0 cpu作为reserved cpu
	assignableCPUs := policy.assignableCPUs(s).Union(reusableCPUs)

	// If there are aligned CPUs in numaAffinity, attempt to take those first.
	result := cpuset.NewCPUSet()
	if numaAffinity != nil {

	}

	// Get any remaining CPUs from what's leftover after attempting to grab aligned ones.
	remainingCPUs, err := takeByTopology(policy.topology, assignableCPUs.Difference(result), numCPUs-result.Size())
	if err != nil {
		return cpuset.NewCPUSet(), err
	}
	result = result.Union(remainingCPUs)

	// 从剩余cpu - result(分配出去的cpus)
	// Remove allocated CPUs from the shared CPUSet.
	s.SetDefaultCPUSet(s.GetDefaultCPUSet().Difference(result))

	klog.Infof("[cpumanager] allocateCPUs: returning %v", result)
	return result, nil
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
