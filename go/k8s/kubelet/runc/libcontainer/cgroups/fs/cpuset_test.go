package fs

import (
	"testing"

	"k8s-lx1036/k8s/kubelet/runc/libcontainer/cgroups/fscommon"
)

// 测试 cpuset.Set() 函数
func TestCpusetSetCpus(t *testing.T) {
	helper := NewCgroupTestUtil("cpuset", t)
	defer helper.cleanup()

	const (
		cpusBefore = "0"
		cpusAfter  = "1-3"
	)

	helper.writeFileContents(map[string]string{
		"cpuset.cpus": cpusBefore,
	})

	helper.CgroupData.config.Resources.CpusetCpus = cpusAfter
	cpuset := &CpusetGroup{}
	if err := cpuset.Set(helper.CgroupPath, helper.CgroupData.config); err != nil {
		t.Fatal(err)
	}

	value, err := fscommon.GetCgroupParamString(helper.CgroupPath, "cpuset.cpus")
	if err != nil {
		t.Fatalf("Failed to parse cpuset.cpus - %s", err)
	}

	if value != cpusAfter {
		t.Fatal("Got the wrong value, set cpuset.cpus failed.")
	}
}
