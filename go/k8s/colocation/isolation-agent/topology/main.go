package main

import (
	"fmt"

	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/kubelet/cadvisor"
	"k8s.io/kubernetes/pkg/kubelet/cm/cpumanager/topology"
)

func main() {
	containerRuntime := "docker"
	rootDirectory := "/var/lib/kubelet"
	remoteRuntimeEndpoint := "unix:///var/run/dockershim.sock"
	cgroupRoots := []string{"/kubepods"}

	imageFsInfoProvider := cadvisor.NewImageFsInfoProvider(containerRuntime, remoteRuntimeEndpoint)
	cadvisorClient, err := cadvisor.New(imageFsInfoProvider, rootDirectory, cgroupRoots,
		cadvisor.UsingLegacyCadvisorStats(containerRuntime, remoteRuntimeEndpoint))
	if err != nil {
		panic(err)
	}

	machineInfo, err := cadvisorClient.MachineInfo()
	if err != nil {
		panic(err)
	}

	capacity := cadvisor.CapacityFromMachineInfo(machineInfo)
	klog.Info(fmt.Sprintf("cpu: %s, memory: %s", capacity.Cpu().String(), capacity.Memory().String()))

	numaNodeInfo, err := topology.GetNUMANodeInfo()
	if err != nil {
		panic(err)
	}

	cpuTopo, err := topology.Discover(machineInfo, numaNodeInfo)
	if err != nil {
		panic(err)
	}
	allCPUs := cpuTopo.CPUDetails.CPUs()
	klog.Info(fmt.Sprintf("NumCPUs[processor,逻辑核]: %d, NumCores[core,物理核]: %d, NumSockets[NUMA node]: %d, allCPUs: %s",
		cpuTopo.NumCPUs, cpuTopo.NumCores, cpuTopo.NumSockets, allCPUs.String()))

	// takeByTopology allocates CPUs associated with low-numbered cores from allCPUs.
	// For example: Given a system with 8 CPUs available and HT enabled,
	// if numReservedCPUs=2, then reserved={0,4}
	//[root@stark12 topology]# numactl -H
	//available: 2 nodes (0-1)
	//node 0 cpus: 0 1 2 3 4 5 12 13 14 15 16 17
	//node 0 size: 32002 MB
	//node 0 free: 19475 MB
	//node 1 cpus: 6 7 8 9 10 11 18 19 20 21 22 23
	//node 1 size: 32253 MB
	//node 1 free: 22313 MB
	//node distances:
	//node   0   1
	//0:  10  21
	//1:  21  10
	numReservedCPUs := 6 // 我想要2个 processor，根据 cpu topo 尽可能让这2个 processor 在一个 core 上
	reserved, err := takeByTopology(cpuTopo, allCPUs, numReservedCPUs)
	if err != nil {
		panic(err)
	}
	prodCPUSet := allCPUs.Difference(reserved)
	klog.Info(fmt.Sprintf("take cpuset %s for nonProdPod, cpuset %s for ProdPod", reserved.String(), prodCPUSet.String()))
}
