package cgroups

import (
	"k8s.io/klog"
	"testing"

	"github.com/opencontainers/runtime-spec/specs-go"
)

func TestCPUSet(test *testing.T) {
	// 1. Create a new cgroup
	// 会创建目录 fixtures/cpuset/test，这里用 ./fixtures 目录代替 /sys/fs/cgroup 目录
	// 可以参见 fixtures/proc/self/mountinfo 文件，等同于 linux /proc/self/mountinfo 文件
	// 只有 container123 cgroup 的 cpuset.cpus="0-9", cpuset.mems="0"
	shares := uint64(100)
	control, err := New(V1, StaticPath("/kubepods/besteffort/pod123/container123"), &specs.LinuxResources{
		CPU: &specs.LinuxCPU{
			Cpus:   "0-9",
			Mems:   "0",
			Shares: &shares,
		},
	})
	if err != nil {
		panic(err)
	}

	//defer control.Delete()

	control.State()

	// 2. Load an existing cgroup
	existingCgroup, err := Load(V1, StaticPath("/kubepods/besteffort/pod123/container123"))
	if err != nil {
		panic(err)
	}
	for _, subsystem := range existingCgroup.Subsystems() {
		klog.Infof("Subsystem: %v", subsystem.Name())
	}

	// 3. Add a process to the cgroup
	// pid "1234" 写入 cgroup.procs, Add() 会覆盖写，而不是 append 追加写
	err = existingCgroup.Add(Process{Pid: 1234})
	if err != nil {
		panic(err)
	}
	err = existingCgroup.Add(Process{Pid: 2345})
	if err != nil {
		panic(err)
	}

	// 4. Update the cgroup
	shares = uint64(200)
	err = existingCgroup.Update(&specs.LinuxResources{
		CPU: &specs.LinuxCPU{
			Cpus:   "10-23",
			Mems:   "1",
			Shares: &shares,
		},
	})
	if err != nil {
		panic(err)
	}

	// 5. List all processes in the cgroup or recursively
	processes, err := existingCgroup.Processes(Cpuset, true)
	if err != nil {
		panic(err)
	}
	for _, process := range processes {
		klog.Infof("process: %v", process)
	}

	//
	//// 6. Get Stats on the cgroup
	//stats, err := control.Stat(cgroups.IgnoreNotExist)
	//if err != nil {
	//	panic(err)
	//}
	//klog.Infof("cgroup stats: %v", *stats)
	//
	//// 7. Move process across cgroups
	//// This allows you to take processes from one cgroup and move them to another
	//shares2 := uint64(100)
	//destination, err := cgroups.New(cgroups.V1, cgroups.StaticPath("/test2"), &specs.LinuxResources{
	//	CPU: &specs.LinuxCPU{
	//		Shares: &shares2,
	//	},
	//})
	//if err != nil {
	//	panic(err)
	//}
	//defer destination.Delete()
	//err = control.MoveTo(destination)
	//if err != nil {
	//	panic(err)
	//}
	//
	//// 8. Create subcgroup
	//subCgroup, err := control.New("child", &specs.LinuxResources{
	//	CPU: &specs.LinuxCPU{
	//		Shares: &shares2,
	//	},
	//})
	//if err != nil {
	//	panic(err)
	//}
	//klog.Infof("cgroup state: %s", subCgroup.State())

}
