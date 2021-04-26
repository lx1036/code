package machine

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/fs"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/info/v1"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/utils/sysfs"

	"golang.org/x/sys/unix"
	"k8s.io/klog/v2"
)

func KernelVersion() string {
	uname := &unix.Utsname{}

	if err := unix.Uname(uname); err != nil {
		return "Unknown"
	}

	return string(uname.Release[:bytes.IndexByte(uname.Release[:], 0)])
}

func getInfoFromFiles(filePaths string) string {
	if len(filePaths) == 0 {
		return ""
	}
	for _, file := range strings.Split(filePaths, ",") {
		id, err := ioutil.ReadFile(file)
		if err == nil {
			return strings.TrimSpace(string(id))
		}
	}
	klog.Warningf("Couldn't collect info from any of the files in %q", filePaths)
	return ""
}

// INFO: 这个函数很重要，直接获取机器的 machine info
func Info(sysFs sysfs.SysFs, fsInfo fs.FsInfo, inHostNamespace bool) (*v1.MachineInfo, error) {
	rootFs := "fixtures"
	if !inHostNamespace {
		rootFs = "/rootfs"
	}

	cpuinfo, err := ioutil.ReadFile(filepath.Join(rootFs, "proc/cpuinfo"))
	if err != nil {
		return nil, err
	}

	clockSpeed, err := GetClockSpeed(cpuinfo)
	if err != nil {
		return nil, err
	}

	filesystems, err := fsInfo.GetGlobalFsInfo()
	if err != nil {
		klog.Errorf("Failed to get global filesystem information: %v", err)
	}

	memoryCapacity, err := GetMachineMemoryCapacity()
	if err != nil {
		return nil, err
	}

	/*memoryByType, err := GetMachineMemoryByType(memoryControllerPath)
	if err != nil {
		return nil, err
	}*/

	topology, numCores, err := GetTopology(sysFs)
	if err != nil {
		klog.Errorf("Failed to get topology information: %v", err)
	}

	machineInfo := &v1.MachineInfo{
		Timestamp:        time.Now(),
		NumCores:         numCores,
		NumPhysicalCores: GetPhysicalCores(cpuinfo),
		NumSockets:       GetSockets(cpuinfo),
		CpuFrequency:     clockSpeed,
		MemoryCapacity:   memoryCapacity,
		//MemoryByType:     memoryByType,
		//NVMInfo:          nvmInfo,
		//HugePages:        hugePagesInfo,
		//DiskMap:          diskMap,
		//NetworkDevices:   netDevices,
		Topology: topology,
		//MachineID:        getInfoFromFiles(filepath.Join(rootFs, *machineIDFilePath)),
		//SystemUUID:       systemUUID,
		//BootID:           getInfoFromFiles(filepath.Join(rootFs, *bootIDFilePath)),
		//CloudProvider:    cloudProvider,
		//InstanceType:     instanceType,
		//InstanceID:       instanceID,
	}

	for i := range filesystems {
		filesystem := filesystems[i]
		inodes := uint64(0)
		if filesystem.Inodes != nil {
			inodes = *filesystem.Inodes
		}
		machineInfo.Filesystems = append(machineInfo.Filesystems,
			v1.FsInfo{
				Device:      filesystem.Device,
				DeviceMajor: uint64(filesystem.Major),
				DeviceMinor: uint64(filesystem.Minor),
				Type:        filesystem.Type.String(),
				Capacity:    filesystem.Capacity,
				Inodes:      inodes,
				HasInodes:   filesystem.Inodes != nil})
	}

	return machineInfo, nil
}
