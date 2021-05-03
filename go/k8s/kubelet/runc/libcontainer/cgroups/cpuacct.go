package cgroups

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"k8s-lx1036/k8s/kubelet/runc/libcontainer/configs"
)

const (
	cgroupCpuacctStat     = "cpuacct.stat"
	cgroupCpuacctUsageAll = "cpuacct.usage_all"

	userModeColumn              = 1
	kernelModeColumn            = 2
	cuacctUsageAllColumnsNumber = 3

	nanosecondsInSecond = 1000000000
	// The value comes from `C.sysconf(C._SC_CLK_TCK)`, and
	// on Linux it's a constant which is safe to be hard coded,
	// so we can avoid using cgo here. For details, see:
	// https://github.com/containerd/cgroups/pull/12
	clockTicks uint64 = 100
)

type CpuacctGroup struct {
}

func (cpuacctGroup *CpuacctGroup) Name() string {
	return Cpuacct
}

func (cpuacctGroup *CpuacctGroup) GetStats(path string, stats *Stats) error {
	userModeUsage, kernelModeUsage, err := getCpuUsageBreakdown(path)
	if err != nil {
		return err
	}

	// INFO: 读 cpuacct.usage 文件
	totalUsage, err := GetCgroupParamUint(path, "cpuacct.usage")
	if err != nil {
		return err
	}

	// INFO: 读 cpuacct.usage_percpu 文件: 0 12376452075 309341 0 0 0 0 0 0 0 0 0 0 12601097394 0 0 0 0 0 0 54457 0 0 0
	// 主要在 cpu1 和 cpu13 逻辑核上
	percpuUsage, err := getPercpuUsage(path)
	if err != nil {
		return err
	}

	percpuUsageInKernelmode, percpuUsageInUsermode, err := getPercpuUsageInModes(path)
	if err != nil {
		return err
	}

	stats.CpuStats.CpuUsage.TotalUsage = totalUsage
	stats.CpuStats.CpuUsage.PercpuUsage = percpuUsage
	stats.CpuStats.CpuUsage.PercpuUsageInKernelmode = percpuUsageInKernelmode
	stats.CpuStats.CpuUsage.PercpuUsageInUsermode = percpuUsageInUsermode
	stats.CpuStats.CpuUsage.UsageInUsermode = userModeUsage
	stats.CpuStats.CpuUsage.UsageInKernelmode = kernelModeUsage
	return nil
}

func getPercpuUsage(path string) ([]uint64, error) {
	percpuUsage := []uint64{}
	data, err := ioutil.ReadFile(filepath.Join(path, "cpuacct.usage_percpu"))
	if err != nil {
		return percpuUsage, err
	}
	for _, value := range strings.Fields(string(data)) {
		value, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return percpuUsage, fmt.Errorf("Unable to convert param value to uint64: %s", err)
		}
		percpuUsage = append(percpuUsage, value)
	}
	return percpuUsage, nil
}

var (
	fixturesCPUAcctPath = "../../../fixtures/cpuacct"
)

func SetFixturesCPUAcctPath(path string) {
	fixturesCPUAcctPath = path
}
func GetFixturesCPUAcctPath() string {
	return fixturesCPUAcctPath
}

// INFO: 读取 cpuacct.usage_all 文件
func getPercpuUsageInModes(path string) ([]uint64, []uint64, error) {
	usageKernelMode := []uint64{}
	usageUserMode := []uint64{}

	path, err := filepath.Abs(GetFixturesCPUAcctPath())
	if err != nil {
		return nil, nil, err
	}
	file, err := os.Open(filepath.Join(path, cgroupCpuacctUsageAll))
	if os.IsNotExist(err) {
		return usageKernelMode, usageUserMode, nil
	} else if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan() //skipping header line

	for scanner.Scan() {
		lineFields := strings.SplitN(scanner.Text(), " ", cuacctUsageAllColumnsNumber+1)
		if len(lineFields) != cuacctUsageAllColumnsNumber {
			continue
		}

		usageInKernelMode, err := strconv.ParseUint(lineFields[kernelModeColumn], 10, 64)
		if err != nil {
			return nil, nil, fmt.Errorf("Unable to convert CPU usage in kernel mode to uint64: %s", err)
		}
		usageKernelMode = append(usageKernelMode, usageInKernelMode)

		usageInUserMode, err := strconv.ParseUint(lineFields[userModeColumn], 10, 64)
		if err != nil {
			return nil, nil, fmt.Errorf("Unable to convert CPU usage in user mode to uint64: %s", err)
		}
		usageUserMode = append(usageUserMode, usageInUserMode)
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("Problem in reading %s line by line, %s", cgroupCpuacctUsageAll, err)
	}

	return usageKernelMode, usageUserMode, nil
}

// Returns user and kernel usage breakdown in nanoseconds.
// INFO: 读取 cpuacct.stat 文件
func getCpuUsageBreakdown(path string) (uint64, uint64, error) {
	userModeUsage := uint64(0)
	kernelModeUsage := uint64(0)
	const (
		userField   = "user"
		systemField = "system"
	)

	// Expected format:
	// user <usage in ticks>
	// system <usage in ticks>
	path, err := filepath.Abs(GetFixturesCPUAcctPath())
	if err != nil {
		return userModeUsage, kernelModeUsage, err
	}
	data, err := ioutil.ReadFile(filepath.Join(path, cgroupCpuacctStat))
	if err != nil {
		return userModeUsage, kernelModeUsage, err
	}
	fields := strings.Fields(string(data))
	if len(fields) < 4 {
		return userModeUsage, kernelModeUsage, fmt.Errorf("failure - %s is expected to have at least 4 fields", filepath.Join(path, cgroupCpuacctStat))
	}
	if fields[0] != userField {
		return userModeUsage, kernelModeUsage, fmt.Errorf("unexpected field %q in %q, expected %q", fields[0], cgroupCpuacctStat, userField)
	}
	if fields[2] != systemField {
		return userModeUsage, kernelModeUsage, fmt.Errorf("unexpected field %q in %q, expected %q", fields[2], cgroupCpuacctStat, systemField)
	}
	if userModeUsage, err = strconv.ParseUint(fields[1], 10, 64); err != nil {
		return 0, 0, err
	}
	if kernelModeUsage, err = strconv.ParseUint(fields[3], 10, 64); err != nil {
		return 0, 0, err
	}

	return (userModeUsage * nanosecondsInSecond) / clockTicks, (kernelModeUsage * nanosecondsInSecond) / clockTicks, nil
}

func (cpuacctGroup *CpuacctGroup) Apply(c *cgroupData) error {
	panic("implement me")
}

func (cpuacctGroup *CpuacctGroup) Set(path string, cgroup *configs.Cgroup) error {
	panic("implement me")
}
