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

func newContainerStats(libcontainerStats *libcontainer.Stats, includedMetrics container.MetricSet) *v1.ContainerStats {
	containerStats := &v1.ContainerStats{
		Timestamp: time.Now(),
	}

	if s := libcontainerStats.CgroupStats; s != nil {
		setCPUStats(s, containerStats, includedMetrics.Has(container.PerCpuUsageMetrics))
		if includedMetrics.Has(container.DiskIOMetrics) {
			setDiskIoStats(s, containerStats)
		}
		setMemoryStats(s, containerStats)
		if includedMetrics.Has(container.HugetlbUsageMetrics) {
			setHugepageStats(s, containerStats)
		}
	}

	if len(libcontainerStats.Interfaces) > 0 {
		setNetworkStats(libcontainerStats, containerStats)
	}

	return containerStats
}
