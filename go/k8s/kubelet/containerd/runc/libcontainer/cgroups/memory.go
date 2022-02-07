package cgroups

import (
	"k8s-lx1036/k8s/kubelet/runc/libcontainer/configs"
)

type MemoryGroup struct {
}

func (memoryGroup *MemoryGroup) Name() string {
	return Memory
}

func (memoryGroup *MemoryGroup) GetStats(path string, stats *Stats) error {
	panic("implement me")
}

func (memoryGroup *MemoryGroup) Apply(c *cgroupData) error {
	panic("implement me")
}

func (memoryGroup *MemoryGroup) Set(path string, cgroup *configs.Cgroup) error {
	panic("implement me")
}
