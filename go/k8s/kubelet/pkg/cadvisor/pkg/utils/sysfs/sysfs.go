package sysfs

import (
	"fmt"
	"os"
	"path/filepath"
)

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

const (
	blockDir     = "/sys/block"
	cacheDir     = "/sys/devices/system/cpu/cpu"
	netDir       = "/sys/class/net"
	dmiDir       = "/sys/class/dmi"
	ppcDevTree   = "/proc/device-tree"
	s390xDevTree = "/etc" // s390/s390x changes

	coreIDFilePath    = "/topology/core_id"
	packageIDFilePath = "/topology/physical_package_id"
	meminfoFile       = "meminfo"

	cpuDirPattern  = "cpu*[0-9]"
	nodeDirPattern = "node*[0-9]"

	//HugePagesNrFile name of nr_hugepages file in sysfs
	HugePagesNrFile = "nr_hugepages"
)

var (
	nodeDir = "/sys/devices/system/node/"
)

type realSysFs struct{}

func (fs *realSysFs) GetNodesPaths() ([]string, error) {
	pathPattern := fmt.Sprintf("%s%s", nodeDir, nodeDirPattern)
	return filepath.Glob(pathPattern)
}

func (fs *realSysFs) GetCPUsPaths(cpusPath string) ([]string, error) {
	panic("implement me")
}

func (fs *realSysFs) GetCoreID(coreIDFilePath string) (string, error) {
	panic("implement me")
}

func (fs *realSysFs) GetCPUPhysicalPackageID(cpuPath string) (string, error) {
	panic("implement me")
}

func (fs *realSysFs) GetMemInfo(nodeDir string) (string, error) {
	panic("implement me")
}

func (fs *realSysFs) GetHugePagesInfo(hugePagesDirectory string) ([]os.FileInfo, error) {
	panic("implement me")
}

func (fs *realSysFs) GetHugePagesNr(hugePagesDirectory string, hugePageName string) (string, error) {
	panic("implement me")
}

func (fs *realSysFs) GetBlockDevices() ([]os.FileInfo, error) {
	panic("implement me")
}

func (fs *realSysFs) GetBlockDeviceSize(s string) (string, error) {
	panic("implement me")
}

func (fs *realSysFs) GetBlockDeviceScheduler(s string) (string, error) {
	panic("implement me")
}

func (fs *realSysFs) GetBlockDeviceNumbers(s string) (string, error) {
	panic("implement me")
}

func (fs *realSysFs) GetNetworkDevices() ([]os.FileInfo, error) {
	panic("implement me")
}

func (fs *realSysFs) GetNetworkAddress(s string) (string, error) {
	panic("implement me")
}

func (fs *realSysFs) GetNetworkMtu(s string) (string, error) {
	panic("implement me")
}

func (fs *realSysFs) GetNetworkSpeed(s string) (string, error) {
	panic("implement me")
}

func (fs *realSysFs) GetNetworkStatValue(dev string, stat string) (uint64, error) {
	panic("implement me")
}

func (fs *realSysFs) GetCaches(id int) ([]os.FileInfo, error) {
	panic("implement me")
}

func (fs *realSysFs) GetCacheInfo(cpu int, cache string) (CacheInfo, error) {
	panic("implement me")
}

func (fs *realSysFs) GetSystemUUID() (string, error) {
	panic("implement me")
}

func (fs *realSysFs) IsCPUOnline(dir string) bool {
	panic("implement me")
}

func NewRealSysFs() SysFs {
	return &realSysFs{}
}
