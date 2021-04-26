package sysfs

import "os"

type CacheInfo struct {
	// size in bytes
	Size uint64
	// cache type - instruction, data, unified
	Type string
	// distance from cpus in a multi-level hierarchy
	Level int
	// number of cpus that can access this cache.
	Cpus int
}

// Abstracts the lowest level calls to sysfs.
type SysFs interface {
	// Get NUMA nodes paths
	GetNodesPaths() ([]string, error)
	// Get paths to CPUs in provided directory e.g. /sys/devices/system/node/node0 or /sys/devices/system/cpu
	GetCPUsPaths(cpusPath string) ([]string, error)
	// Get physical core id for specified CPU
	GetCoreID(coreIDFilePath string) (string, error)
	// Get physical package id for specified CPU
	GetCPUPhysicalPackageID(cpuPath string) (string, error)
	// Get total memory for specified NUMA node
	GetMemInfo(nodeDir string) (string, error)
	// Get hugepages from specified directory
	GetHugePagesInfo(hugePagesDirectory string) ([]os.FileInfo, error)
	// Get hugepage_nr from specified directory
	GetHugePagesNr(hugePagesDirectory string, hugePageName string) (string, error)
	// Get directory information for available block devices.
	GetBlockDevices() ([]os.FileInfo, error)
	// Get Size of a given block device.
	GetBlockDeviceSize(string) (string, error)
	// Get scheduler type for the block device.
	GetBlockDeviceScheduler(string) (string, error)
	// Get device major:minor number string.
	GetBlockDeviceNumbers(string) (string, error)

	GetNetworkDevices() ([]os.FileInfo, error)
	GetNetworkAddress(string) (string, error)
	GetNetworkMtu(string) (string, error)
	GetNetworkSpeed(string) (string, error)
	GetNetworkStatValue(dev string, stat string) (uint64, error)

	// Get directory information for available caches accessible to given cpu.
	GetCaches(id int) ([]os.FileInfo, error)
	// Get information for a cache accessible from the given cpu.
	GetCacheInfo(cpu int, cache string) (CacheInfo, error)

	GetSystemUUID() (string, error)
	// IsCPUOnline determines if CPU status from kernel hotplug machanism standpoint.
	// See: https://www.kernel.org/doc/html/latest/core-api/cpu_hotplug.html
	IsCPUOnline(dir string) bool
}
