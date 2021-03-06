package v2

import (
	"time"

	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/info/v1"
)

const (
	TypeName   = "name"
	TypeDocker = "docker"
)

type CpuSpec struct {
	// Requested cpu shares. Default is 1024.
	Limit uint64 `json:"limit"`
	// Requested cpu hard limit. Default is unlimited (0).
	// Units: milli-cpus.
	MaxLimit uint64 `json:"max_limit"`
	// Cpu affinity mask.
	// TODO(rjnagal): Add a library to convert mask string to set of cpu bitmask.
	Mask string `json:"mask,omitempty"`
	// CPUQuota Default is disabled
	Quota uint64 `json:"quota,omitempty"`
	// Period is the CPU reference time in ns e.g the quota is compared against this.
	Period uint64 `json:"period,omitempty"`
}

type MemorySpec struct {
	// The amount of memory requested. Default is unlimited (-1).
	// Units: bytes.
	Limit uint64 `json:"limit,omitempty"`

	// The amount of guaranteed memory.  Default is 0.
	// Units: bytes.
	Reservation uint64 `json:"reservation,omitempty"`

	// The amount of swap space requested. Default is unlimited (-1).
	// Units: bytes.
	SwapLimit uint64 `json:"swap_limit,omitempty"`
}

// Instantaneous CPU stats
type CpuInstStats struct {
	Usage CpuInstUsage `json:"usage"`
}

// CPU usage time statistics.
type CpuInstUsage struct {
	// Total CPU usage.
	// Units: nanocores per second
	Total uint64 `json:"total"`

	// Per CPU/core usage of the container.
	// Unit: nanocores per second
	PerCpu []uint64 `json:"per_cpu_usage,omitempty"`

	// Time spent in user space.
	// Unit: nanocores per second
	User uint64 `json:"user"`

	// Time spent in kernel space.
	// Unit: nanocores per second
	System uint64 `json:"system"`
}

type RequestOptions struct {
	// Type of container identifier specified - "name", "dockerid", dockeralias"
	IdType string `json:"type"`
	// Number of stats to return
	Count int `json:"count"`
	// Whether to include stats for child subcontainers.
	Recursive bool `json:"recursive"`
	// Update stats if they are older than MaxAge
	// nil indicates no update, and 0 will always trigger an update.
	MaxAge *time.Duration `json:"max_age"`
}

// Filesystem usage statistics.
type FilesystemStats struct {
	// Total Number of bytes consumed by container.
	TotalUsageBytes *uint64 `json:"totalUsageBytes,omitempty"`
	// Number of bytes consumed by a container through its root filesystem.
	BaseUsageBytes *uint64 `json:"baseUsageBytes,omitempty"`
	// Number of inodes used within the container's root filesystem.
	// This only accounts for inodes that are shared across containers,
	// and does not include inodes used in mounted directories.
	InodeUsage *uint64 `json:"containter_inode_usage,omitempty"`
}

type FsInfo struct {
	// Time of generation of these stats.
	Timestamp time.Time `json:"timestamp"`

	// The block device name associated with the filesystem.
	Device string `json:"device"`

	// Path where the filesystem is mounted.
	Mountpoint string `json:"mountpoint"`

	// Filesystem usage in bytes.
	Capacity uint64 `json:"capacity"`

	// Bytes available for non-root use.
	Available uint64 `json:"available"`

	// Number of bytes used on this filesystem.
	Usage uint64 `json:"usage"`

	// Labels associated with this filesystem.
	Labels []string `json:"labels"`

	// Number of Inodes.
	Inodes *uint64 `json:"inodes,omitempty"`

	// Number of available Inodes (if known)
	InodesFree *uint64 `json:"inodes_free,omitempty"`
}

type ProcessInfo struct {
	User          string  `json:"user"`
	Pid           int     `json:"pid"`
	Ppid          int     `json:"parent_pid"`
	StartTime     string  `json:"start_time"`
	PercentCpu    float32 `json:"percent_cpu"`
	PercentMemory float32 `json:"percent_mem"`
	RSS           uint64  `json:"rss"`
	VirtualSize   uint64  `json:"virtual_size"`
	Status        string  `json:"status"`
	RunningTime   string  `json:"running_time"`
	CgroupPath    string  `json:"cgroup_path"`
	Cmd           string  `json:"cmd"`
	FdCount       int     `json:"fd_count"`
	Psr           int     `json:"psr"`
}

type ContainerInfo struct {
	// Describes the container.
	Spec ContainerSpec `json:"spec,omitempty"`

	// Historical statistics gathered from the container.
	Stats []*ContainerStats `json:"stats,omitempty"`
}

type ContainerSpec struct {
	// Time at which the container was created.
	CreationTime time.Time `json:"creation_time,omitempty"`

	// Other names by which the container is known within a certain namespace.
	// This is unique within that namespace.
	Aliases []string `json:"aliases,omitempty"`

	// Namespace under which the aliases of a container are unique.
	// An example of a namespace is "docker" for Docker containers.
	Namespace string `json:"namespace,omitempty"`

	// Metadata labels associated with this container.
	Labels map[string]string `json:"labels,omitempty"`
	// Metadata envs associated with this container. Only whitelisted envs are added.
	Envs map[string]string `json:"envs,omitempty"`

	HasCpu bool    `json:"has_cpu"`
	Cpu    CpuSpec `json:"cpu,omitempty"`

	HasMemory bool       `json:"has_memory"`
	Memory    MemorySpec `json:"memory,omitempty"`

	HasHugetlb bool `json:"has_hugetlb"`

	HasCustomMetrics bool            `json:"has_custom_metrics"`
	CustomMetrics    []v1.MetricSpec `json:"custom_metrics,omitempty"`

	HasProcesses bool           `json:"has_processes"`
	Processes    v1.ProcessSpec `json:"processes,omitempty"`

	// Following resources have no associated spec, but are being isolated.
	HasNetwork    bool `json:"has_network"`
	HasFilesystem bool `json:"has_filesystem"`
	HasDiskIo     bool `json:"has_diskio"`

	// Image name used for this container.
	Image string `json:"image,omitempty"`
}

type TcpStat struct {
	// Count of TCP connections in state "Established"
	Established uint64
	// Count of TCP connections in state "Syn_Sent"
	SynSent uint64
	// Count of TCP connections in state "Syn_Recv"
	SynRecv uint64
	// Count of TCP connections in state "Fin_Wait1"
	FinWait1 uint64
	// Count of TCP connections in state "Fin_Wait2"
	FinWait2 uint64
	// Count of TCP connections in state "Time_Wait
	TimeWait uint64
	// Count of TCP connections in state "Close"
	Close uint64
	// Count of TCP connections in state "Close_Wait"
	CloseWait uint64
	// Count of TCP connections in state "Listen_Ack"
	LastAck uint64
	// Count of TCP connections in state "Listen"
	Listen uint64
	// Count of TCP connections in state "Closing"
	Closing uint64
}

type NetworkStats struct {
	// Network stats by interface.
	Interfaces []v1.InterfaceStats `json:"interfaces,omitempty"`
	// TCP connection stats (Established, Listen...)
	Tcp TcpStat `json:"tcp"`
	// TCP6 connection stats (Established, Listen...)
	Tcp6 TcpStat `json:"tcp6"`
	// UDP connection stats
	Udp v1.UdpStat `json:"udp"`
	// UDP6 connection stats
	Udp6 v1.UdpStat `json:"udp6"`
	// TCP advanced stats
	TcpAdvanced v1.TcpAdvancedStat `json:"tcp_advanced"`
}

type ContainerStats struct {
	// The time of this stat point.
	Timestamp time.Time `json:"timestamp"`
	// CPU statistics
	// In nanoseconds (aggregated)
	Cpu *v1.CpuStats `json:"cpu,omitempty"`
	// In nanocores per second (instantaneous)
	CpuInst *CpuInstStats `json:"cpu_inst,omitempty"`
	// Disk IO statistics
	DiskIo *v1.DiskIoStats `json:"diskio,omitempty"`
	// Memory statistics
	Memory *v1.MemoryStats `json:"memory,omitempty"`
	// Hugepage statistics
	//Hugetlb *map[string]v1.HugetlbStats `json:"hugetlb,omitempty"`
	// Network statistics
	Network *NetworkStats `json:"network,omitempty"`
	// Processes statistics
	Processes *v1.ProcessStats `json:"processes,omitempty"`
	// Filesystem statistics
	Filesystem *FilesystemStats `json:"filesystem,omitempty"`
	// Task load statistics
	//Load *v1.LoadStats `json:"load_stats,omitempty"`
	// Metrics for Accelerators. Each Accelerator corresponds to one element in the array.
	Accelerators []v1.AcceleratorStats `json:"accelerators,omitempty"`
	// Custom Metrics
	CustomMetrics map[string][]v1.MetricVal `json:"custom_metrics,omitempty"`
	// Perf events counters
	//PerfStats []v1.PerfStat `json:"perf_stats,omitempty"`
	// Statistics originating from perf uncore events.
	// Applies only for root container.
	//PerfUncoreStats []v1.PerfUncoreStat `json:"perf_uncore_stats,omitempty"`
	// Referenced memory
	ReferencedMemory uint64 `json:"referenced_memory,omitempty"`
	// Resource Control (resctrl) statistics
	//Resctrl v1.ResctrlStats `json:"resctrl,omitempty"`
}
