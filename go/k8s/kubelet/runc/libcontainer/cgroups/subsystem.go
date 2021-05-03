package cgroups

import (
	"errors"

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
	subsystemsLegacy = subsystemSet{
		&CpusetGroup{},
		&CpuGroup{},
		&MemoryGroup{},
		&CpuacctGroup{},
		//&PidsGroup{},
	}
)

var errSubsystemDoesNotExist = errors.New("cgroup: subsystem does not exist")

type subsystemSet []subsystem

func (s subsystemSet) Get(name string) (subsystem, error) {
	for _, ss := range s {
		if ss.Name() == name {
			return ss, nil
		}
	}
	return nil, errSubsystemDoesNotExist
}

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
