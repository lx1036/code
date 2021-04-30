package cgroups

import (
	"bufio"
	"os"
	"path/filepath"

	"k8s-lx1036/k8s/kubelet/runc/libcontainer/configs"
)

type CpuGroup struct {
}

func (cpuGroup *CpuGroup) Name() string {
	return Cpu
}

// GetStats INFO: 主要就是读取 cpu.stat 文件获取 cpu throttling 相关数据
func (cpuGroup *CpuGroup) GetStats(path string, stats *Stats) error {
	path, err := filepath.Abs("fixtures/cpu")
	if err != nil {
		return err
	}
	f, err := os.Open(filepath.Join(path, "cpu.stat"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()

	/*
		nr_periods 107235
		nr_throttled 0
		throttled_time 0
	*/

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		t, v, err := GetCgroupParamKeyValue(sc.Text())
		if err != nil {
			return err
		}
		switch t {
		case "nr_periods":
			stats.CpuStats.ThrottlingData.Periods = v

		case "nr_throttled":
			stats.CpuStats.ThrottlingData.ThrottledPeriods = v

		case "throttled_time":
			stats.CpuStats.ThrottlingData.ThrottledTime = v
		}
	}

	return nil
}

func (cpuGroup *CpuGroup) Apply(c *cgroupData) error {
	panic("implement me")
}

func (cpuGroup *CpuGroup) Set(path string, cgroup *configs.Cgroup) error {
	panic("implement me")
}
