package v1

import "time"

// Container reference contains enough information to uniquely identify a container
type ContainerReference struct {
	// The container id
	Id string `json:"id,omitempty"`

	// The absolute name of the container. This is unique on the machine.
	Name string `json:"name"`

	// Other names by which the container is known within a certain namespace.
	// This is unique within that namespace.
	Aliases []string `json:"aliases,omitempty"`

	// Namespace under which the aliases of a container are unique.
	// An example of a namespace is "docker" for Docker containers.
	Namespace string `json:"namespace,omitempty"`
}

// Sorts by container name.
type ContainerReferenceSlice []ContainerReference

func (s ContainerReferenceSlice) Len() int           { return len(s) }
func (s ContainerReferenceSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ContainerReferenceSlice) Less(i, j int) bool { return s[i].Name < s[j].Name }

type CpuSpec struct {
	Limit    uint64 `json:"limit"`
	MaxLimit uint64 `json:"max_limit"`
	Mask     string `json:"mask,omitempty"`
	Quota    uint64 `json:"quota,omitempty"`
	Period   uint64 `json:"period,omitempty"`
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

type ProcessSpec struct {
	Limit uint64 `json:"limit,omitempty"`
}

type ContainerSpec struct {
	// Time at which the container was created.
	CreationTime time.Time `json:"creation_time,omitempty"`

	// Metadata labels associated with this container.
	Labels map[string]string `json:"labels,omitempty"`
	// Metadata envs associated with this container. Only whitelisted envs are added.
	Envs map[string]string `json:"envs,omitempty"`

	HasCpu bool    `json:"has_cpu"`
	Cpu    CpuSpec `json:"cpu,omitempty"`

	HasMemory bool       `json:"has_memory"`
	Memory    MemorySpec `json:"memory,omitempty"`

	HasHugetlb bool `json:"has_hugetlb"`

	HasNetwork bool `json:"has_network"`

	HasProcesses bool        `json:"has_processes"`
	Processes    ProcessSpec `json:"processes,omitempty"`

	HasFilesystem bool `json:"has_filesystem"`

	// HasDiskIo when true, indicates that DiskIo stats will be available.
	HasDiskIo bool `json:"has_diskio"`

	HasCustomMetrics bool `json:"has_custom_metrics"`
	//CustomMetrics    []MetricSpec `json:"custom_metrics,omitempty"`

	// Image name used for this container.
	Image string `json:"image,omitempty"`
}

// CPU usage time statistics.
type CpuUsage struct {
	// Total CPU usage.
	// Unit: nanoseconds.
	Total uint64 `json:"total"`

	// Per CPU/core usage of the container.
	// Unit: nanoseconds.
	PerCpu []uint64 `json:"per_cpu_usage,omitempty"`

	// Time spent in user space.
	// Unit: nanoseconds.
	User uint64 `json:"user"`

	// Time spent in kernel space.
	// Unit: nanoseconds.
	System uint64 `json:"system"`
}

// Cpu Completely Fair Scheduler statistics.
type CpuCFS struct {
	// Total number of elapsed enforcement intervals.
	Periods uint64 `json:"periods"`

	// Total number of times tasks in the cgroup have been throttled.
	ThrottledPeriods uint64 `json:"throttled_periods"`

	// Total time duration for which tasks in the cgroup have been throttled.
	// Unit: nanoseconds.
	ThrottledTime uint64 `json:"throttled_time"`
}

// Cpu Aggregated scheduler statistics
type CpuSchedstat struct {
	// https://www.kernel.org/doc/Documentation/scheduler/sched-stats.txt

	// time spent on the cpu
	RunTime uint64 `json:"run_time"`
	// time spent waiting on a runqueue
	RunqueueTime uint64 `json:"runqueue_time"`
	// # of timeslices run on this cpu
	RunPeriods uint64 `json:"run_periods"`
}

// All CPU usage metrics are cumulative from the creation of the container
type CpuStats struct {
	Usage     CpuUsage     `json:"usage"`
	CFS       CpuCFS       `json:"cfs"`
	Schedstat CpuSchedstat `json:"schedstat"`
	// Smoothed average of number of runnable threads x 1000.
	// We multiply by thousand to avoid using floats, but preserving precision.
	// Load is smoothed over the last 10 seconds. Instantaneous value can be read
	// from LoadStats.NrRunning.
	LoadAverage int32 `json:"load_average"`
}
type PerDiskStats struct {
	Device string            `json:"device"`
	Major  uint64            `json:"major"`
	Minor  uint64            `json:"minor"`
	Stats  map[string]uint64 `json:"stats"`
}

type DiskIoStats struct {
	IoServiceBytes []PerDiskStats `json:"io_service_bytes,omitempty"`
	IoServiced     []PerDiskStats `json:"io_serviced,omitempty"`
	IoQueued       []PerDiskStats `json:"io_queued,omitempty"`
	Sectors        []PerDiskStats `json:"sectors,omitempty"`
	IoServiceTime  []PerDiskStats `json:"io_service_time,omitempty"`
	IoWaitTime     []PerDiskStats `json:"io_wait_time,omitempty"`
	IoMerged       []PerDiskStats `json:"io_merged,omitempty"`
	IoTime         []PerDiskStats `json:"io_time,omitempty"`
}

type MemoryStats struct {
	// Current memory usage, this includes all memory regardless of when it was
	// accessed.
	// Units: Bytes.
	Usage uint64 `json:"usage"`

	// Maximum memory usage recorded.
	// Units: Bytes.
	MaxUsage uint64 `json:"max_usage"`

	// Number of bytes of page cache memory.
	// Units: Bytes.
	Cache uint64 `json:"cache"`

	// The amount of anonymous and swap cache memory (includes transparent
	// hugepages).
	// Units: Bytes.
	RSS uint64 `json:"rss"`

	// The amount of swap currently used by the processes in this cgroup
	// Units: Bytes.
	Swap uint64 `json:"swap"`

	// The amount of memory used for mapped files (includes tmpfs/shmem)
	MappedFile uint64 `json:"mapped_file"`

	// The amount of working set memory, this includes recently accessed memory,
	// dirty memory, and kernel memory. Working set is <= "usage".
	// Units: Bytes.
	WorkingSet uint64 `json:"working_set"`

	Failcnt uint64 `json:"failcnt"`

	ContainerData    MemoryStatsMemoryData `json:"container_data,omitempty"`
	HierarchicalData MemoryStatsMemoryData `json:"hierarchical_data,omitempty"`
}

type MemoryStatsMemoryData struct {
	Pgfault    uint64 `json:"pgfault"`
	Pgmajfault uint64 `json:"pgmajfault"`
}

type FsStats struct {
	// The block device name associated with the filesystem.
	Device string `json:"device,omitempty"`

	// Type of the filesytem.
	Type string `json:"type"`

	// Number of bytes that can be consumed by the container on this filesystem.
	Limit uint64 `json:"capacity"`

	// Number of bytes that is consumed by the container on this filesystem.
	Usage uint64 `json:"usage"`

	// Base Usage that is consumed by the container's writable layer.
	// This field is only applicable for docker container's as of now.
	BaseUsage uint64 `json:"base_usage"`

	// Number of bytes available for non-root user.
	Available uint64 `json:"available"`

	// HasInodes when true, indicates that Inodes info will be available.
	HasInodes bool `json:"has_inodes"`

	// Number of Inodes
	Inodes uint64 `json:"inodes"`

	// Number of available Inodes
	InodesFree uint64 `json:"inodes_free"`

	// Number of reads completed
	// This is the total number of reads completed successfully.
	ReadsCompleted uint64 `json:"reads_completed"`

	// Number of reads merged
	// Reads and writes which are adjacent to each other may be merged for
	// efficiency.  Thus two 4K reads may become one 8K read before it is
	// ultimately handed to the disk, and so it will be counted (and queued)
	// as only one I/O.  This field lets you know how often this was done.
	ReadsMerged uint64 `json:"reads_merged"`

	// Number of sectors read
	// This is the total number of sectors read successfully.
	SectorsRead uint64 `json:"sectors_read"`

	// Number of milliseconds spent reading
	// This is the total number of milliseconds spent by all reads (as
	// measured from __make_request() to end_that_request_last()).
	ReadTime uint64 `json:"read_time"`

	// Number of writes completed
	// This is the total number of writes completed successfully.
	WritesCompleted uint64 `json:"writes_completed"`

	// Number of writes merged
	// See the description of reads merged.
	WritesMerged uint64 `json:"writes_merged"`

	// Number of sectors written
	// This is the total number of sectors written successfully.
	SectorsWritten uint64 `json:"sectors_written"`

	// Number of milliseconds spent writing
	// This is the total number of milliseconds spent by all writes (as
	// measured from __make_request() to end_that_request_last()).
	WriteTime uint64 `json:"write_time"`

	// Number of I/Os currently in progress
	// The only field that should go to zero. Incremented as requests are
	// given to appropriate struct request_queue and decremented as they finish.
	IoInProgress uint64 `json:"io_in_progress"`

	// Number of milliseconds spent doing I/Os
	// This field increases so long as field 9 is nonzero.
	IoTime uint64 `json:"io_time"`

	// weighted number of milliseconds spent doing I/Os
	// This field is incremented at each I/O start, I/O completion, I/O
	// merge, or read of these stats by the number of I/Os in progress
	// (field 9) times the number of milliseconds spent doing I/O since the
	// last update of this field.  This can provide an easy measure of both
	// I/O completion time and the backlog that may be accumulating.
	WeightedIoTime uint64 `json:"weighted_io_time"`
}

type ContainerStats struct {
	// The time of this stat point.
	Timestamp time.Time   `json:"timestamp"`
	Cpu       CpuStats    `json:"cpu,omitempty"`
	DiskIo    DiskIoStats `json:"diskio,omitempty"`
	Memory    MemoryStats `json:"memory,omitempty"`
	//Hugetlb   map[string]HugetlbStats `json:"hugetlb,omitempty"`
	//Network   NetworkStats            `json:"network,omitempty"`
	// Filesystem statistics
	Filesystem []FsStats `json:"filesystem,omitempty"`

	// Task load stats
	//TaskStats LoadStats `json:"task_stats,omitempty"`

	// Metrics for Accelerators. Each Accelerator corresponds to one element in the array.
	//Accelerators []AcceleratorStats `json:"accelerators,omitempty"`

	// ProcessStats for Containers
	//Processes ProcessStats `json:"processes,omitempty"`

	// Custom metrics from all collectors
	//CustomMetrics map[string][]MetricVal `json:"custom_metrics,omitempty"`

	// Statistics originating from perf events
	//PerfStats []PerfStat `json:"perf_stats,omitempty"`

	// Statistics originating from perf uncore events.
	// Applies only for root container.
	//PerfUncoreStats []PerfUncoreStat `json:"perf_uncore_stats,omitempty"`

	// Referenced memory
	ReferencedMemory uint64 `json:"referenced_memory,omitempty"`

	// Resource Control (resctrl) statistics
	//Resctrl ResctrlStats `json:"resctrl,omitempty"`
}
