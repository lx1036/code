package cgroups

import (
	"k8s.io/klog/v2"
	"testing"
)

func TestGetCgroupMounts(t *testing.T) {
	mounts, err := getCgroupMountsV1(true)
	if err != nil {
		panic(err)
	}

	for _, mount := range mounts {
		klog.Infof("Mountpoint %s, Root %s, Subsystems %v", mount.Mountpoint, mount.Root, mount.Subsystems)
	}
	// output:
	// Mountpoint /sys/fs/cgroup/systemd, Root /, Subsystems [systemd]
	// Mountpoint /sys/fs/cgroup/cpuset, Root /, Subsystems [cpuset]
	// Mountpoint /sys/fs/cgroup/cpu,cpuacct, Root /, Subsystems [cpuacct cpu]
	// Mountpoint /sys/fs/cgroup/memory, Root /, Subsystems [memory]
	// Mountpoint /sys/fs/cgroup/devices, Root /, Subsystems [devices]
	// Mountpoint /sys/fs/cgroup/freezer, Root /, Subsystems [freezer]
	// Mountpoint /sys/fs/cgroup/net_cls, Root /, Subsystems [net_cls]
	// Mountpoint /sys/fs/cgroup/blkio, Root /, Subsystems [blkio]
	// Mountpoint /sys/fs/cgroup/perf_event, Root /, Subsystems [perf_event]
	// Mountpoint /sys/fs/cgroup/hugetlb, Root /, Subsystems [hugetlb]

}
