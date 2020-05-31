package pod

import "k8s-lx1036/k8s-ui/dashboard/controllers/resource/common"

type PodMetrics struct {
	CpuUsage           *uint64              `json:"cpuUsage"`
	MemoryUsage        *uint64              `json:"memoryUsage"`
	CpuUsageHistory    []common.MetricPoint `json:"cpuUsageHistory"`
	MemoryUsageHistory []common.MetricPoint `json:"memoryUsageHistory"`
}
