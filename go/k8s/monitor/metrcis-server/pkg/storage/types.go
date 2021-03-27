package storage

import (
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
)

type MetricsBatch struct {
	Nodes []NodeMetricsPoint
	Pods  []PodMetricsPoint
}

type NodeMetricsPoint struct {
	Name string
	MetricsPoint
}

type PodMetricsPoint struct {
	Name      string
	Namespace string

	Containers []ContainerMetricsPoint
}

type ContainerMetricsPoint struct {
	Name string
	MetricsPoint
}

type MetricsPoint struct {
	Timestamp time.Time
	// CpuUsage is the CPU usage rate, in cores
	CpuUsage resource.Quantity
	// MemoryUsage is the working set size, in bytes.
	MemoryUsage resource.Quantity
}
