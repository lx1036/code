package machine

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"

	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/info/v1"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/utils/sysfs"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/utils/sysinfo"

	"golang.org/x/sys/unix"
	"k8s.io/klog/v2"
)

var (
	coreRegExp = regexp.MustCompile(`(?m)^core id\s*:\s*([0-9]+)$`)
	nodeRegExp = regexp.MustCompile(`(?m)^physical id\s*:\s*([0-9]+)$`)
	// Power systems have a different format so cater for both
	cpuClockSpeedMHz     = regexp.MustCompile(`(?:cpu MHz|CPU MHz|clock)\s*:\s*([0-9]+\.[0-9]+)(?:MHz)?`)
	memoryCapacityRegexp = regexp.MustCompile(`MemTotal:\s*([0-9]+) kB`)
	swapCapacityRegexp   = regexp.MustCompile(`SwapTotal:\s*([0-9]+) kB`)

	cpuBusPath         = "/sys/bus/cpu/devices/"
	isMemoryController = regexp.MustCompile("mc[0-9]+")
	isDimm             = regexp.MustCompile("dimm[0-9]+")
	machineArch        = getMachineArch()
	maxFreqFile        = "/sys/devices/system/cpu/cpu0/cpufreq/cpuinfo_max_freq"
)

// GetPhysicalCores returns number of CPU cores reading /proc/cpuinfo file or if needed information from sysfs cpu path
func GetPhysicalCores(procInfo []byte) int {
	numCores := getUniqueMatchesCount(string(procInfo), coreRegExp)
	if numCores == 0 {
		// read number of cores from /sys/bus/cpu/devices/cpu*/topology/core_id to deal with processors
		// for which 'core id' is not available in /proc/cpuinfo
		//numCores = getUniqueCPUPropertyCount(cpuBusPath, sysFsCPUCoreID)
	}
	if numCores == 0 {
		klog.Errorf("Cannot read number of physical cores correctly, number of cores set to %d", numCores)
	}
	return numCores
}

// GetSockets returns number of CPU sockets reading /proc/cpuinfo file or if needed information from sysfs cpu path
func GetSockets(procInfo []byte) int {
	numSocket := getUniqueMatchesCount(string(procInfo), nodeRegExp)
	if numSocket == 0 {
		// read number of sockets from /sys/bus/cpu/devices/cpu*/topology/physical_package_id to deal with processors
		// for which 'physical id' is not available in /proc/cpuinfo
		//numSocket = getUniqueCPUPropertyCount(cpuBusPath, sysFsCPUPhysicalPackageID)
	}
	if numSocket == 0 {
		klog.Errorf("Cannot read number of sockets correctly, number of sockets set to %d", numSocket)
	}
	return numSocket
}

// GetClockSpeed returns the CPU clock speed, given a []byte formatted as the /proc/cpuinfo file.
func GetClockSpeed(procInfo []byte) (uint64, error) {
	// First look through sys to find a max supported cpu frequency.
	/*if utils.FileExists(maxFreqFile) {
		val, err := ioutil.ReadFile(maxFreqFile)
		if err != nil {
			return 0, err
		}
		var maxFreq uint64
		n, err := fmt.Sscanf(string(val), "%d", &maxFreq)
		if err != nil || n != 1 {
			return 0, fmt.Errorf("could not parse frequency %q", val)
		}
		return maxFreq, nil
	}*/

	// Fall back to /proc/cpuinfo
	matches := cpuClockSpeedMHz.FindSubmatch(procInfo)
	if len(matches) != 2 {
		return 0, fmt.Errorf("could not detect clock speed from output: %q", string(procInfo))
	}

	speed, err := strconv.ParseFloat(string(matches[1]), 64)
	if err != nil {
		return 0, err
	}
	// Convert to kHz
	return uint64(speed * 1000), nil
}

var (
	fixturesMemInfoPath = "../../../../fixtures/proc/meminfo"
)

func SetFixturesMemInfoPath(path string) {
	fixturesMemInfoPath = path
}
func GetFixturesMemInfoPath() string {
	return fixturesMemInfoPath
}

// GetMachineMemoryCapacity returns the machine's total memory from /proc/meminfo.
// Returns the total memory capacity as an uint64 (number of bytes).
func GetMachineMemoryCapacity() (uint64, error) {
	out, err := ioutil.ReadFile(GetFixturesMemInfoPath())
	if err != nil {
		return 0, err
	}

	memoryCapacity, err := parseCapacity(out, memoryCapacityRegexp)
	if err != nil {
		return 0, err
	}
	return memoryCapacity, err
}

// GetTopology returns CPU topology reading information from sysfs
func GetTopology(sysFs sysfs.SysFs) ([]v1.Node, int, error) {
	return sysinfo.GetNodesInfo(sysFs)
}

// getUniqueMatchesCount returns number of unique matches in given argument using provided regular expression
func getUniqueMatchesCount(s string, r *regexp.Regexp) int {
	matches := r.FindAllString(s, -1)
	uniques := make(map[string]bool)
	for _, match := range matches {
		uniques[match] = true
	}
	return len(uniques)
}

func getMachineArch() string {
	uname := unix.Utsname{}
	err := unix.Uname(&uname)
	if err != nil {
		klog.Errorf("Cannot get machine architecture, err: %v", err)
		return ""
	}
	return string(uname.Machine[:])
}

// parseCapacity matches a Regexp in a []byte, returning the resulting value in bytes.
// Assumes that the value matched by the Regexp is in KB.
func parseCapacity(b []byte, r *regexp.Regexp) (uint64, error) {
	matches := r.FindSubmatch(b)
	if len(matches) != 2 {
		return 0, fmt.Errorf("failed to match regexp in output: %q", string(b))
	}
	m, err := strconv.ParseUint(string(matches[1]), 10, 64)
	if err != nil {
		return 0, err
	}

	// Convert to bytes.
	return m * 1024, err
}
