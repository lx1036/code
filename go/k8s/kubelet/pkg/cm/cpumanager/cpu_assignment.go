package cpumanager

import (
	"k8s-lx1036/k8s/kubelet/pkg/cm/cpumanager/topology"
	"k8s-lx1036/k8s/kubelet/pkg/cm/cpuset"
)

// 根据topo来分配cpu，尽量cpus在一个core里
func takeByTopology(topo *topology.CPUTopology, availableCPUs cpuset.CPUSet, numCPUs int) (cpuset.CPUSet, error) {

	return cpuset.NewCPUSet(), nil
}
