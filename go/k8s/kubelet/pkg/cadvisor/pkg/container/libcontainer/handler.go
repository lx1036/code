package libcontainer

import (
	"time"

	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/container"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/info/v1"
	"k8s-lx1036/k8s/kubelet/runc/libcontainer"
	"k8s-lx1036/k8s/kubelet/runc/libcontainer/cgroups"
)

type Handler struct {
	cgroupManager   cgroups.Manager
	rootFs          string
	pid             int
	includedMetrics container.MetricSet
	pidMetricsCache map[int]*v1.CpuSchedstat
	cycles          uint64
}

func NewHandler(cgroupManager cgroups.Manager, rootFs string, pid int, includedMetrics container.MetricSet) *Handler {
	return &Handler{
		cgroupManager:   cgroupManager,
		rootFs:          rootFs,
		pid:             pid,
		includedMetrics: includedMetrics,
		pidMetricsCache: make(map[int]*v1.CpuSchedstat),
	}
}

// Get cgroup and networking stats of the specified container
func (h *Handler) GetStats() (*v1.ContainerStats, error) {
	// INFO: 使用 cgroupManager 读取 cgroups 文件数据
	cgroupStats, err := h.cgroupManager.GetStats()
	if err != nil {
		return nil, err
	}

	libcontainerStats := &libcontainer.Stats{
		CgroupStats: cgroupStats,
	}
	stats := newContainerStats(libcontainerStats, h.includedMetrics)

	return stats, nil
}

// Convert libcontainer stats to info.ContainerStats.
func setCPUStats(s *cgroups.Stats, ret *v1.ContainerStats, withPerCPU bool) {
	ret.Cpu.Usage.User = s.CpuStats.CpuUsage.UsageInUsermode
	ret.Cpu.Usage.System = s.CpuStats.CpuUsage.UsageInKernelmode
	ret.Cpu.Usage.Total = s.CpuStats.CpuUsage.TotalUsage
	ret.Cpu.CFS.Periods = s.CpuStats.ThrottlingData.Periods
	ret.Cpu.CFS.ThrottledPeriods = s.CpuStats.ThrottlingData.ThrottledPeriods
	ret.Cpu.CFS.ThrottledTime = s.CpuStats.ThrottlingData.ThrottledTime

	if !withPerCPU {
		return
	}
	if len(s.CpuStats.CpuUsage.PercpuUsage) == 0 {
		// libcontainer's 'GetStats' can leave 'PercpuUsage' nil if it skipped the
		// cpuacct subsystem.
		return
	}
	ret.Cpu.Usage.PerCpu = s.CpuStats.CpuUsage.PercpuUsage
}

func setMemoryStats(s *cgroups.Stats, ret *v1.ContainerStats) {
	ret.Memory.Usage = s.MemoryStats.Usage.Usage
	ret.Memory.MaxUsage = s.MemoryStats.Usage.MaxUsage
	ret.Memory.Failcnt = s.MemoryStats.Usage.Failcnt

	if s.MemoryStats.UseHierarchy {
		ret.Memory.Cache = s.MemoryStats.Stats["total_cache"]
		ret.Memory.RSS = s.MemoryStats.Stats["total_rss"]
		ret.Memory.Swap = s.MemoryStats.Stats["total_swap"]
		ret.Memory.MappedFile = s.MemoryStats.Stats["total_mapped_file"]
	} else {
		ret.Memory.Cache = s.MemoryStats.Stats["cache"]
		ret.Memory.RSS = s.MemoryStats.Stats["rss"]
		ret.Memory.Swap = s.MemoryStats.Stats["swap"]
		ret.Memory.MappedFile = s.MemoryStats.Stats["mapped_file"]
	}
	if v, ok := s.MemoryStats.Stats["pgfault"]; ok {
		ret.Memory.ContainerData.Pgfault = v
		ret.Memory.HierarchicalData.Pgfault = v
	}
	if v, ok := s.MemoryStats.Stats["pgmajfault"]; ok {
		ret.Memory.ContainerData.Pgmajfault = v
		ret.Memory.HierarchicalData.Pgmajfault = v
	}

	inactiveFileKeyName := "total_inactive_file"

	workingSet := ret.Memory.Usage
	if v, ok := s.MemoryStats.Stats[inactiveFileKeyName]; ok {
		if workingSet < v {
			workingSet = 0
		} else {
			workingSet -= v
		}
	}
	ret.Memory.WorkingSet = workingSet
}

func newContainerStats(libcontainerStats *libcontainer.Stats, includedMetrics container.MetricSet) *v1.ContainerStats {
	containerStats := &v1.ContainerStats{
		Timestamp: time.Now(),
	}

	if s := libcontainerStats.CgroupStats; s != nil {
		setCPUStats(s, containerStats, includedMetrics.Has(container.PerCpuUsageMetrics))
		if includedMetrics.Has(container.DiskIOMetrics) {
			//setDiskIoStats(s, containerStats)
		}
		setMemoryStats(s, containerStats)
		if includedMetrics.Has(container.HugetlbUsageMetrics) {
			//setHugepageStats(s, containerStats)
		}
	}

	if len(libcontainerStats.Interfaces) > 0 {
		//setNetworkStats(libcontainerStats, containerStats)
	}

	return containerStats
}
