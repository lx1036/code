package cpumanager

import (
	"fmt"
	"k8s.io/klog/v2"
	"sort"

	"k8s-lx1036/k8s/kubelet/pkg/cm/cpumanager/topology"
	"k8s-lx1036/k8s/kubelet/pkg/cm/cpuset"
)

// INFO: NUMA架构包含多个nodes，每个node包含多个cpu(core)/memory，每个cpu包含两个processor。`top` 看到的是 processor 数量。

/*

cat /proc/cpuinfo | grep "physical id" | sort | uniq # 查看物理cpu数量
physical id	: 0
physical id	: 1

cat /proc/cpuinfo | grep "cores" | sort | uniq # 查看每块cpu的核数
cpu cores	: 8

cat /proc/cpuinfo | grep "processor" | wc -l # 查看主机总逻辑线程数
32

是否开启超线程= 总逻辑线程数 / 物理cpu数量 * 每块cpu核数 == 2
这台机器 32/ (2 * 8) == 2，开启了超线程

NUMA(Non-uniform memory access): 主机板上如果插有多块 CPU 的话，那么就是 NUMA 架构。每块 CPU 独占一块面积，一般都有独立风扇。
查看本机NUMA结构，和上面数据是对应的，并且0-7和16-23 core在node0上，8-15和24-31在node1上。
所以kubelet topomanager在分配cpu时，尽可能让core在相同的node上，比如某个进程需要分配两个独占的core，那这两个core就应该在相同的node上
，当然前提是该node上有空余的两个core：

yum install -y numactl
numactl -H
node 0 cpus: 0 1 2 3 4 5 6 7 16 17 18 19 20 21 22 23
node 0 size: 63892 MB
node 0 free: 57744 MB
node 1 cpus: 8 9 10 11 12 13 14 15 24 25 26 27 28 29 30 31
node 1 size: 64508 MB
node 1 free: 58472 MB
node distances:
node   0   1
  0:  10  21
  1:  21  10

将一个pid写入 /sys/fs/cgroup/cpuset/test/cgroup.procs 和 /sys/fs/cgroup/cpuset/test/tasks 区别是：
操作系统以线程为调度单位，将一个一般的 pid 写入到 tasks 中，只有这个 pid 对应的线程，以及由它产生的其他进程、线程会属于这个控制组。
而把 pid 写入 cgroups.procs，操作系统则会把找到其所属进程的所有线程，把它们统统加入到当前控制组。


echo 0 > /sys/fs/cgroup/cpuset/test/cpuset.cpus
echo 0 > /sys/fs/cgroup/cpuset/test/cpuset.mems

echo $$ > /sys/fs/cgroup/cpuset/test/cgroup.procs # 写入当前进程编号
启动一个计算密集型的任务，申请用 4 个逻辑核: stress -c 4 &
只有 cpu0 利用率达到了 100%
top - 11:05:06 up 15 days, 20:41,  3 users,  load average: 2.51, 0.85, 0.35
Tasks: 580 total,   5 running, 360 sleeping,   0 stopped,   0 zombie
%Cpu0  :100.0 us,  0.0 sy,  0.0 ni,  0.0 id,  0.0 wa,  0.0 hi,  0.0 si,  0.0 st
%Cpu1  :  0.3 us,  0.0 sy,  0.0 ni, 99.7 id,  0.0 wa,  0.0 hi,  0.0 si,  0.0 st
%Cpu2  :  0.7 us,  0.3 sy,  0.0 ni, 99.0 id,  0.0 wa,  0.0 hi,  0.0 si,  0.0 st
%Cpu3  :  0.7 us,  0.0 sy,  0.0 ni, 99.3 id,  0.0 wa,  0.0 hi,  0.0 si,  0.0 st

*/

// INFO: 是否开启超线程= 物理cpu数量 * 每块cpu核数 / 总逻辑线程数 == 2
/*

Socket: 是一个物理上的概念，指的是主板上的cpu插槽

Node:由于SMP体系中各个CPU访问内存只能通过单一的通道，导致内存访问成为瓶颈，cpu再多也无用。后来引入了NUMA，通过划分node，
每个node有本地RAM，这样node内访问RAM速度会非常快。但跨Node的RAM访问代价会相对高一点,
我们用Node之间的距离（Distance，抽象的概念）来定义各个Node之间互访资源的开销。

Core: 就是一个物理cpu,一个独立的硬件执行单元，比如寄存器，计算单元等

Thread: 就是超线程（HyperThreading）的概念，是一个逻辑cpu，共享core上的执行单元

HT: Hyperthreading  使操作系统认为处理器的核心数是实际核心数的2倍，超线程(hyper-threading)本质上就是CPU支持的同时多线程(simultaneous multi-threading)技术，
简单理解就是对CPU的虚拟化，一颗物理CPU可以被操作系统当做多颗CPU来使用。Hyper-threading只是一种“欺骗”手段

cpuset: cpuset作为cgroups的子系统，主要用于numa架构，用于设置cpu的亲和性，为 cgroup 中的 task 分配独立的 CPU和内存等。
cpuset使用sysfs提供用户态接口，可以通过普通文件读写，
工作流程为：cpuset调用sched_setaffinity来设置进程的cpu、内存的亲和性，调用mbind和set_mempolicy来设置内存的亲和性。

*/

type cpuAccumulator struct {
	topo          *topology.CPUTopology
	details       topology.CPUDetails
	numCPUsNeeded int
	result        cpuset.CPUSet
}

func newCPUAccumulator(topo *topology.CPUTopology, availableCPUs cpuset.CPUSet, numCPUs int) *cpuAccumulator {
	return &cpuAccumulator{
		topo:          topo,
		details:       topo.CPUDetails.KeepOnly(availableCPUs), // details=topo - availableCPUs, KeepOnly 求交集, availableCPUs 不能超过实际 cpu topo 逻辑核
		numCPUsNeeded: numCPUs,
		result:        cpuset.NewCPUSet(),
	}
}

func (a *cpuAccumulator) isSatisfied() bool {
	return a.numCPUsNeeded < 1
}

func (a *cpuAccumulator) isFailed() bool {
	return a.numCPUsNeeded > a.details.CPUs().Size()
}

func (a *cpuAccumulator) needs(n int) bool {
	return a.numCPUsNeeded >= n
}

// Returns true if the supplied socket is fully available in `topoDetails`.
func (a *cpuAccumulator) isSocketFree(socketID int) bool {
	cpusInSocket := a.details.CPUsInSockets(socketID).Size()
	cpusPerSocket := a.topo.CPUsPerSocket()
	return cpusInSocket == cpusPerSocket
}

// Returns true if the supplied core is fully available in `topoDetails`.
// free core 表示该 core(物理核)上所有逻辑核都没有被分配出去
func (a *cpuAccumulator) isCoreFree(coreID int) bool {
	return a.details.CPUsInCores(coreID).Size() == a.topo.CPUsPerCore()
}

// Returns free socket IDs as a slice sorted by:
// - socket ID, ascending.
// free socket 表示该 socket(numa node)上所有逻辑核(processor)都没有被分配出去
func (a *cpuAccumulator) freeSockets() []int {
	return a.details.Sockets().Filter(a.isSocketFree).ToSlice()
}

// Returns core IDs as a slice sorted by:
// - the number of whole available cores on the socket, ascending
// - socket ID, ascending
// - core ID, ascending
// free core 表示该 core(物理核)上所有逻辑核都没有被分配出去
func (a *cpuAccumulator) freeCores() []int {
	socketIDs := a.details.Sockets().ToSliceNoSort()
	sort.Slice(socketIDs,
		func(i, j int) bool {
			iCores := a.details.CoresInSockets(socketIDs[i]).Filter(a.isCoreFree)
			jCores := a.details.CoresInSockets(socketIDs[j]).Filter(a.isCoreFree)
			return iCores.Size() < jCores.Size() || socketIDs[i] < socketIDs[j]
		})

	coreIDs := []int{}
	for _, s := range socketIDs {
		coreIDs = append(coreIDs, a.details.CoresInSockets(s).Filter(a.isCoreFree).ToSlice()...)
	}

	return coreIDs
}

// Returns CPU IDs as a slice sorted by:
// - socket affinity with result
// - number of CPUs available on the same socket
// - number of CPUs available on the same core
// - socket ID.
// - core ID.
func (a *cpuAccumulator) freeCPUs() []int {
	result := []int{}
	cores := a.details.Cores().ToSlice()

	sort.Slice(
		cores,
		func(i, j int) bool {
			iCore := cores[i]
			jCore := cores[j]

			iCPUs := a.topo.CPUDetails.CPUsInCores(iCore).ToSlice()
			jCPUs := a.topo.CPUDetails.CPUsInCores(jCore).ToSlice()

			iSocket := a.topo.CPUDetails[iCPUs[0]].SocketID
			jSocket := a.topo.CPUDetails[jCPUs[0]].SocketID

			// Compute the number of CPUs in the result reside on the same socket
			// as each core.
			iSocketColoScore := a.topo.CPUDetails.CPUsInSockets(iSocket).Intersection(a.result).Size()
			jSocketColoScore := a.topo.CPUDetails.CPUsInSockets(jSocket).Intersection(a.result).Size()

			// Compute the number of available CPUs available on the same socket
			// as each core.
			iSocketFreeScore := a.details.CPUsInSockets(iSocket).Size()
			jSocketFreeScore := a.details.CPUsInSockets(jSocket).Size()

			// Compute the number of available CPUs on each core.
			iCoreFreeScore := a.details.CPUsInCores(iCore).Size()
			jCoreFreeScore := a.details.CPUsInCores(jCore).Size()

			return iSocketColoScore > jSocketColoScore ||
				iSocketFreeScore < jSocketFreeScore ||
				iCoreFreeScore < jCoreFreeScore ||
				iSocket < jSocket ||
				iCore < jCore
		})

	// For each core, append sorted CPU IDs to result.
	for _, core := range cores {
		result = append(result, a.details.CPUsInCores(core).ToSlice()...)
	}

	return result
}

// INFO: 分配
func (a *cpuAccumulator) take(cpus cpuset.CPUSet) {
	a.result = a.result.Union(cpus)
	a.details = a.details.KeepOnly(a.details.CPUs().Difference(a.result))
	a.numCPUsNeeded -= cpus.Size()
}

// 根据topo来分配cpu，尽量cpus在一个core里
func takeByTopology(topo *topology.CPUTopology, availableCPUs cpuset.CPUSet, numCPUs int) (cpuset.CPUSet, error) {
	acc := newCPUAccumulator(topo, availableCPUs, numCPUs)
	if acc.isSatisfied() {
		return acc.result, nil
	}
	if acc.isFailed() {
		return cpuset.NewCPUSet(), fmt.Errorf("not enough cpus available to satisfy request")
	}

	// CPU  - logical CPU, cadvisor - thread
	// Core - physical CPU, cadvisor - Core
	// Socket - socket, cadvisor - Node
	// INFO: 先看逻辑核，即socket
	// Algorithm: topology-aware best-fit
	// 1. Acquire whole sockets, if available and the container requires at
	//    least a socket's-worth of CPUs.
	// 如果需要的cpus大于一个socket包含的cpus
	if acc.needs(acc.topo.CPUsPerSocket()) {
		for _, s := range acc.freeSockets() {
			klog.V(4).Infof("[cpumanager] takeByTopology: claiming socket [%d]", s)
			acc.take(acc.details.CPUsInSockets(s))
			if acc.isSatisfied() {
				return acc.result, nil
			}
			if !acc.needs(acc.topo.CPUsPerSocket()) {
				break
			}
		}
	}

	// INFO: 再看物理核，即 core
	// 2. Acquire whole cores, if available and the container requires at least
	//    a core's-worth of CPUs.
	if acc.needs(acc.topo.CPUsPerCore()) {
		for _, c := range acc.freeCores() {
			klog.V(4).Infof("[cpumanager] takeByTopology: claiming core [%d]", c)
			acc.take(acc.details.CPUsInCores(c))
			if acc.isSatisfied() {
				return acc.result, nil
			}
			if !acc.needs(acc.topo.CPUsPerCore()) {
				break
			}
		}
	}

	// INFO: 最后看线程，即 cpu
	// 3. Acquire single threads, preferring to fill partially-allocated cores
	//    on the same sockets as the whole cores we have already taken in this
	//    allocation.
	for _, c := range acc.freeCPUs() {
		klog.V(4).Infof("[cpumanager] takeByTopology: claiming CPU [%d]", c)
		if acc.needs(1) {
			acc.take(cpuset.NewCPUSet(c))
		}
		if acc.isSatisfied() {
			return acc.result, nil
		}
	}

	return cpuset.NewCPUSet(), fmt.Errorf("failed to allocate cpus")
}
