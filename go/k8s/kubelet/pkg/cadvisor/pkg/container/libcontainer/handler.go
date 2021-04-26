package libcontainer

import (
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/container"

	info "github.com/google/cadvisor/info/v1"
	"github.com/opencontainers/runc/libcontainer/cgroups"
)

type Handler struct {
	cgroupManager   cgroups.Manager
	rootFs          string
	pid             int
	includedMetrics container.MetricSet
	pidMetricsCache map[int]*info.CpuSchedstat
	cycles          uint64
}

func NewHandler(cgroupManager cgroups.Manager, rootFs string, pid int, includedMetrics container.MetricSet) *Handler {
	return &Handler{
		cgroupManager:   cgroupManager,
		rootFs:          rootFs,
		pid:             pid,
		includedMetrics: includedMetrics,
		pidMetricsCache: make(map[int]*info.CpuSchedstat),
	}
}
