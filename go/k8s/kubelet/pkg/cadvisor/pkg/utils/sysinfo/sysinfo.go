package sysinfo

import (
	"fmt"
	"os"
	"regexp"
	"strconv"

	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/info/v1"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/utils/sysfs"

	"k8s.io/klog/v2"
)

var (
	schedulerRegExp      = regexp.MustCompile(`.*\[(.*)\].*`)
	nodeDirRegExp        = regexp.MustCompile(`node/node(\d*)`)
	cpuDirRegExp         = regexp.MustCompile(`/cpu(\d+)`)
	memoryCapacityRegexp = regexp.MustCompile(`MemTotal:\s*([0-9]+) kB`)

	cpusPath = "/sys/devices/system/cpu"
)

// GetNodesInfo returns information about NUMA nodes and their topology
// INFO: 这个函数很重要，获取宿主机 NUMA node topo, kubelet->cadvisor->GetNodesInfo() 从而获取机器的 cpu topo 信息
// 1. 读取 /sys/devices/system/node/ 目录获取 numa node
func GetNodesInfo(sysFs sysfs.SysFs) ([]v1.Node, int, error) {
	var nodes []v1.Node
	allLogicalCoresCount := 0

	nodesDirs, err := sysFs.GetNodesPaths()
	if err != nil {
		return nil, 0, err
	}

	if len(nodesDirs) == 0 {
		klog.Warningf("Nodes topology is not available, providing CPU topology")
		//return getCPUTopology(sysFs)
		return nil, 0, err
	}

	for _, nodeDir := range nodesDirs {
		id, err := getMatchedInt(nodeDirRegExp, nodeDir)
		if err != nil {
			return nil, 0, err
		}
		node := v1.Node{Id: id}

		cpuDirs, err := sysFs.GetCPUsPaths(nodeDir)
		if len(cpuDirs) == 0 {
			klog.Warningf("Found node without any CPU, nodeDir: %s, number of cpuDirs %d, err: %v", nodeDir, len(cpuDirs), err)
		} else {
			cores, err := getCoresInfo(sysFs, cpuDirs)
			if err != nil {
				return nil, 0, err
			}
			node.Cores = cores
			for _, core := range cores {
				allLogicalCoresCount += len(core.Threads)
			}
		}

		node.Memory, err = getNodeMemInfo(sysFs, nodeDir)
		if err != nil {
			return nil, 0, err
		}

		nodes = append(nodes, node)
	}

	return nodes, allLogicalCoresCount, err
}

// getCoresInfo retruns infromation about physical cores
func getCoresInfo(sysFs sysfs.SysFs, cpuDirs []string) ([]v1.Core, error) {
	cores := make([]v1.Core, 0, len(cpuDirs))
	for _, cpuDir := range cpuDirs {
		cpuID, err := getMatchedInt(cpuDirRegExp, cpuDir)
		if err != nil {
			return nil, fmt.Errorf("unexpected format of CPU directory, cpuDirRegExp %s, cpuDir: %s", cpuDirRegExp, cpuDir)
		}
		if !sysFs.IsCPUOnline(cpuDir) {
			continue
		}

		rawPhysicalID, err := sysFs.GetCoreID(cpuDir)
		if os.IsNotExist(err) {
			klog.Warningf("Cannot read core id for %s, core_id file does not exist, err: %s", cpuDir, err)
			continue
		} else if err != nil {
			return nil, err
		}
		physicalID, err := strconv.Atoi(rawPhysicalID)
		if err != nil {
			return nil, err
		}

		rawPhysicalPackageID, err := sysFs.GetCPUPhysicalPackageID(cpuDir)
		if os.IsNotExist(err) {
			klog.Warningf("Cannot read physical package id for %s, physical_package_id file does not exist, err: %s", cpuDir, err)
			continue
		} else if err != nil {
			return nil, err
		}

		physicalPackageID, err := strconv.Atoi(rawPhysicalPackageID)
		if err != nil {
			return nil, err
		}

		coreIDx := -1
		for id, core := range cores {
			if core.Id == physicalID && core.SocketID == physicalPackageID {
				coreIDx = id
			}
		}
		if coreIDx == -1 {
			cores = append(cores, v1.Core{})
			coreIDx = len(cores) - 1
		}
		desiredCore := &cores[coreIDx]

		desiredCore.Id = physicalID
		desiredCore.SocketID = physicalPackageID

		if len(desiredCore.Threads) == 0 {
			desiredCore.Threads = []int{cpuID}
		} else {
			desiredCore.Threads = append(desiredCore.Threads, cpuID)
		}
	}

	return cores, nil
}

// getNodeMemInfo returns information about total memory for NUMA node
func getNodeMemInfo(sysFs sysfs.SysFs, nodeDir string) (uint64, error) {
	rawMem, err := sysFs.GetMemInfo(nodeDir)
	if err != nil {
		//Ignore if per-node info is not available.
		klog.Warningf("Found node without memory information, nodeDir: %s", nodeDir)
		return 0, nil
	}
	matches := memoryCapacityRegexp.FindStringSubmatch(rawMem)
	if len(matches) != 2 {
		return 0, fmt.Errorf("failed to match regexp in output: %q", string(rawMem))
	}
	memory, err := strconv.ParseUint(matches[1], 10, 64)
	if err != nil {
		return 0, err
	}
	memory = memory * 1024 // Convert to bytes
	return uint64(memory), nil
}

func getMatchedInt(rgx *regexp.Regexp, str string) (int, error) {
	matches := rgx.FindStringSubmatch(str)
	if len(matches) != 2 {
		return 0, fmt.Errorf("failed to match regexp, str: %s", str)
	}
	valInt, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, err
	}
	return valInt, nil
}
