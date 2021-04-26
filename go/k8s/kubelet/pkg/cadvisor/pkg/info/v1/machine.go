package v1

import "time"

type InstanceType string

const (
	UnknownInstance = "Unknown"
)

type InstanceID string

const (
	UnNamedInstance InstanceID = "None"
)

type FsInfo struct {
	// Block device associated with the filesystem.
	Device string `json:"device"`
	// DeviceMajor is the major identifier of the device, used for correlation with blkio stats
	DeviceMajor uint64 `json:"-"`
	// DeviceMinor is the minor identifier of the device, used for correlation with blkio stats
	DeviceMinor uint64 `json:"-"`

	// Total number of bytes available on the filesystem.
	Capacity uint64 `json:"capacity"`

	// Type of device.
	Type string `json:"type"`

	// Total number of inodes available on the filesystem.
	Inodes uint64 `json:"inodes"`

	// HasInodes when true, indicates that Inodes info will be available.
	HasInodes bool `json:"has_inodes"`
}

type Node struct {
	Id int `json:"node_id"`
	// Per-node memory
	Memory uint64 `json:"memory"`
	//HugePages []HugePagesInfo `json:"hugepages"`
	Cores  []Core  `json:"cores"`
	Caches []Cache `json:"caches"`
}

type Core struct {
	Id       int     `json:"core_id"`
	Threads  []int   `json:"thread_ids"`
	Caches   []Cache `json:"caches"`
	SocketID int     `json:"socket_id"`
}

type Cache struct {
	// Size of memory cache in bytes.
	Size uint64 `json:"size"`
	// Type of memory cache: data, instruction, or unified.
	Type string `json:"type"`
	// Level (distance from cpus) in a multi-level cache hierarchy.
	Level int `json:"level"`
}

type MemoryInfo struct {
	// The amount of memory (in bytes).
	Capacity uint64 `json:"capacity"`

	// Number of memory DIMMs.
	DimmCount uint `json:"dimm_count"`
}

type DiskInfo struct {
	// device name
	Name string `json:"name"`

	// Major number
	Major uint64 `json:"major"`

	// Minor number
	Minor uint64 `json:"minor"`

	// Size in bytes
	Size uint64 `json:"size"`

	// I/O Scheduler - one of "none", "noop", "cfq", "deadline"
	Scheduler string `json:"scheduler"`
}

type NetInfo struct {
	// Device name
	Name string `json:"name"`

	// Mac Address
	MacAddress string `json:"mac_address"`

	// Speed in MBits/s
	Speed int64 `json:"speed"`

	// Maximum Transmission Unit
	Mtu int64 `json:"mtu"`
}

type MachineInfo struct {
	// The time of this information point.
	Timestamp time.Time `json:"timestamp"`

	// The number of cores in this machine.
	NumCores int `json:"num_cores"`

	// The number of physical cores in this machine.
	NumPhysicalCores int `json:"num_physical_cores"`

	// The number of cpu sockets in this machine.
	NumSockets int `json:"num_sockets"`

	// Maximum clock speed for the cores, in KHz.
	CpuFrequency uint64 `json:"cpu_frequency_khz"`

	// The amount of memory (in bytes) in this machine
	MemoryCapacity uint64 `json:"memory_capacity"`

	// Memory capacity and number of DIMMs by memory type
	//MemoryByType map[string]*MemoryInfo `json:"memory_by_type"`

	//NVMInfo NVMInfo `json:"nvm"`

	// HugePages on this machine.
	//HugePages []HugePagesInfo `json:"hugepages"`

	// The machine id
	MachineID string `json:"machine_id"`

	// The system uuid
	SystemUUID string `json:"system_uuid"`

	// The boot id
	BootID string `json:"boot_id"`

	// Filesystems on this machine.
	Filesystems []FsInfo `json:"filesystems"`

	// Disk map
	DiskMap map[string]DiskInfo `json:"disk_map"`

	// Network devices
	NetworkDevices []NetInfo `json:"network_devices"`

	// Machine Topology
	// Describes cpu/memory layout and hierarchy.
	Topology []Node `json:"topology"`

	// Cloud provider the machine belongs to.
	//CloudProvider CloudProvider `json:"cloud_provider"`

	// Type of cloud instance (e.g. GCE standard) the machine is.
	InstanceType InstanceType `json:"instance_type"`

	// ID of cloud instance (e.g. instance-1) given to it by the cloud provider.
	InstanceID InstanceID `json:"instance_id"`
}

type VersionInfo struct {
	// Kernel version.
	KernelVersion string `json:"kernel_version"`

	// OS image being used for cadvisor container, or host image if running on host directly.
	ContainerOsVersion string `json:"container_os_version"`

	// Docker version.
	DockerVersion string `json:"docker_version"`

	// Docker API Version
	DockerAPIVersion string `json:"docker_api_version"`

	// cAdvisor version.
	CadvisorVersion string `json:"cadvisor_version"`
	// cAdvisor git revision.
	CadvisorRevision string `json:"cadvisor_revision"`
}

type MachineInfoFactory interface {
	GetMachineInfo() (*MachineInfo, error)
	GetVersionInfo() (*VersionInfo, error)
}
