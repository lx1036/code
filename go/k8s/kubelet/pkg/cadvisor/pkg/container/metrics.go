package container

// MetricKind represents the kind of metrics that cAdvisor exposes.
type MetricKind string

const (
	CpuUsageMetrics                MetricKind = "cpu"
	ProcessSchedulerMetrics        MetricKind = "sched"
	PerCpuUsageMetrics             MetricKind = "percpu"
	MemoryUsageMetrics             MetricKind = "memory"
	CpuLoadMetrics                 MetricKind = "cpuLoad"
	DiskIOMetrics                  MetricKind = "diskIO"
	DiskUsageMetrics               MetricKind = "disk"
	NetworkUsageMetrics            MetricKind = "network"
	NetworkTcpUsageMetrics         MetricKind = "tcp"
	NetworkAdvancedTcpUsageMetrics MetricKind = "advtcp"
	NetworkUdpUsageMetrics         MetricKind = "udp"
	AcceleratorUsageMetrics        MetricKind = "accelerator"
	AppMetrics                     MetricKind = "app"
	ProcessMetrics                 MetricKind = "process"
	HugetlbUsageMetrics            MetricKind = "hugetlb"
	PerfMetrics                    MetricKind = "perf_event"
	ReferencedMemoryMetrics        MetricKind = "referenced_memory"
	CPUTopologyMetrics             MetricKind = "cpu_topology"
	ResctrlMetrics                 MetricKind = "resctrl"
)

// AllMetrics represents all kinds of metrics that cAdvisor supported.
var AllMetrics = MetricSet{
	CpuUsageMetrics:                struct{}{},
	ProcessSchedulerMetrics:        struct{}{},
	PerCpuUsageMetrics:             struct{}{},
	MemoryUsageMetrics:             struct{}{},
	CpuLoadMetrics:                 struct{}{},
	DiskIOMetrics:                  struct{}{},
	AcceleratorUsageMetrics:        struct{}{},
	DiskUsageMetrics:               struct{}{},
	NetworkUsageMetrics:            struct{}{},
	NetworkTcpUsageMetrics:         struct{}{},
	NetworkAdvancedTcpUsageMetrics: struct{}{},
	NetworkUdpUsageMetrics:         struct{}{},
	ProcessMetrics:                 struct{}{},
	AppMetrics:                     struct{}{},
	HugetlbUsageMetrics:            struct{}{},
	PerfMetrics:                    struct{}{},
	ReferencedMemoryMetrics:        struct{}{},
	CPUTopologyMetrics:             struct{}{},
	ResctrlMetrics:                 struct{}{},
}

func (mk MetricKind) String() string {
	return string(mk)
}

type MetricSet map[MetricKind]struct{}

func (ms MetricSet) Has(mk MetricKind) bool {
	_, exists := ms[mk]
	return exists
}

func (ms MetricSet) Add(mk MetricKind) {
	ms[mk] = struct{}{}
}

func (ms MetricSet) Difference(ms1 MetricSet) MetricSet {
	result := MetricSet{}
	for kind := range ms {
		if !ms1.Has(kind) {
			result.Add(kind)
		}
	}
	return result
}
