package cgroupfs

import (
	"k8s-lx1036/k8s/kubelet/runc/libcontainer/configs"
)

const (
	Devices   string = "devices"
	Hugetlb   string = "hugetlb"
	Freezer   string = "freezer"
	Pids      string = "pids"
	NetCLS    string = "net_cls"
	NetPrio   string = "net_prio"
	PerfEvent string = "perf_event"
	Cpuset    string = "cpuset"
	Cpu       string = "cpu"
	Cpuacct   string = "cpuacct"
	Memory    string = "memory"
	Blkio     string = "blkio"
	Rdma      string = "rdma"
)

var (
	subsystemsLegacy = []subsystem{
		&CpusetController{},
		//&MemoryGroup{},
		//&CpuGroup{},
		//&CpuacctGroup{},
		//&PidsGroup{},
	}
)

type subsystem interface {
	// Name returns the name of the subsystem.
	Name() string
	// Returns the stats, as 'stats', corresponding to the cgroup under 'path'.
	GetStats(path string, stats *Stats) error
	// Creates and joins the cgroup represented by 'cgroupData'.
	Apply(c *cgroupData) error
	// Set the cgroup represented by cgroup.
	Set(path string, cgroup *configs.Cgroup) error
}
