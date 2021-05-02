package stats

import (
	"time"

	statsapi "k8s-lx1036/k8s/kubelet/pkg/apis/stats/v1alpha1"
	cadvisorapiv1 "k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/info/v1"
	cadvisorapiv2 "k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/info/v2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

// defaultNetworkInterfaceName is used for collectng network stats.
// This logic relies on knowledge of the container runtime implementation and
// is not reliable.
const defaultNetworkInterfaceName = "eth0"

// latestContainerStats returns the latest container stats from cadvisor, or nil if none exist
func latestContainerStats(info *cadvisorapiv2.ContainerInfo) (*cadvisorapiv2.ContainerStats, bool) {
	stats := info.Stats
	if len(stats) < 1 {
		return nil, false
	}
	latest := stats[len(stats)-1]
	if latest == nil {
		return nil, false
	}
	return latest, true
}

// cadvisorInfoToNetworkStats returns the statsapi.NetworkStats converted from
// the container info from cadvisor.
func cadvisorInfoToNetworkStats(info *cadvisorapiv2.ContainerInfo) *statsapi.NetworkStats {
	if !info.Spec.HasNetwork {
		return nil
	}
	cstat, found := latestContainerStats(info)
	if !found {
		return nil
	}

	if cstat.Network == nil {
		return nil
	}

	iStats := statsapi.NetworkStats{
		Time: metav1.NewTime(cstat.Timestamp),
	}

	for i := range cstat.Network.Interfaces {
		inter := cstat.Network.Interfaces[i]
		iStat := statsapi.InterfaceStats{
			Name:     inter.Name,
			RxBytes:  &inter.RxBytes,
			RxErrors: &inter.RxErrors,
			TxBytes:  &inter.TxBytes,
			TxErrors: &inter.TxErrors,
		}

		if inter.Name == defaultNetworkInterfaceName {
			iStats.InterfaceStats = iStat
		}

		iStats.Interfaces = append(iStats.Interfaces, iStat)
	}

	return &iStats
}

func uint64Ptr(i uint64) *uint64 {
	return &i
}
func isMemoryUnlimited(v uint64) bool {
	// Size after which we consider memory to be "unlimited". This is not
	// MaxInt64 due to rounding by the kernel.
	// TODO: cadvisor should export this https://github.com/google/cadvisor/blob/master/metrics/prometheus.go#L596
	const maxMemorySize = uint64(1 << 62)

	return v > maxMemorySize
}

func cadvisorInfoToCPUandMemoryStats(info *cadvisorapiv2.ContainerInfo) (*statsapi.CPUStats, *statsapi.MemoryStats) {
	cstat, found := latestContainerStats(info)
	if !found {
		return nil, nil
	}
	var cpuStats *statsapi.CPUStats
	var memoryStats *statsapi.MemoryStats
	cpuStats = &statsapi.CPUStats{
		Time:                 metav1.NewTime(cstat.Timestamp),
		UsageNanoCores:       uint64Ptr(0),
		UsageCoreNanoSeconds: uint64Ptr(0),
	}
	if info.Spec.HasCpu {
		if cstat.CpuInst != nil {
			cpuStats.UsageNanoCores = &cstat.CpuInst.Usage.Total
		}
		if cstat.Cpu != nil {
			cpuStats.UsageCoreNanoSeconds = &cstat.Cpu.Usage.Total
		}
	}
	if info.Spec.HasMemory && cstat.Memory != nil {
		pageFaults := cstat.Memory.ContainerData.Pgfault
		majorPageFaults := cstat.Memory.ContainerData.Pgmajfault
		memoryStats = &statsapi.MemoryStats{
			Time:            metav1.NewTime(cstat.Timestamp),
			UsageBytes:      &cstat.Memory.Usage,
			WorkingSetBytes: &cstat.Memory.WorkingSet,
			RSSBytes:        &cstat.Memory.RSS,
			PageFaults:      &pageFaults,
			MajorPageFaults: &majorPageFaults,
		}
		// availableBytes = memory limit (if known) - workingset
		if !isMemoryUnlimited(info.Spec.Memory.Limit) {
			availableBytes := info.Spec.Memory.Limit - cstat.Memory.WorkingSet
			memoryStats.AvailableBytes = &availableBytes
		}
	} else {
		memoryStats = &statsapi.MemoryStats{
			Time:            metav1.NewTime(cstat.Timestamp),
			WorkingSetBytes: uint64Ptr(0),
		}
	}

	return cpuStats, memoryStats
}

// cadvisorInfoToContainerCPUAndMemoryStats returns the statsapi.ContainerStats converted
// from the container and filesystem info.
func cadvisorInfoToContainerCPUAndMemoryStats(name string, info *cadvisorapiv2.ContainerInfo) *statsapi.ContainerStats {
	result := &statsapi.ContainerStats{
		StartTime: metav1.NewTime(info.Spec.CreationTime),
		Name:      name,
	}

	cpu, memory := cadvisorInfoToCPUandMemoryStats(info)
	result.CPU = cpu
	result.Memory = memory

	return result
}

func buildLogsStats(cstat *cadvisorapiv2.ContainerStats, rootFs *cadvisorapiv2.FsInfo) *statsapi.FsStats {
	fsStats := &statsapi.FsStats{
		Time:           metav1.NewTime(cstat.Timestamp),
		AvailableBytes: &rootFs.Available,
		CapacityBytes:  &rootFs.Capacity,
		InodesFree:     rootFs.InodesFree,
		Inodes:         rootFs.Inodes,
	}

	if rootFs.Inodes != nil && rootFs.InodesFree != nil {
		logsInodesUsed := *rootFs.Inodes - *rootFs.InodesFree
		fsStats.InodesUsed = &logsInodesUsed
	}
	return fsStats
}

func buildRootfsStats(cstat *cadvisorapiv2.ContainerStats, imageFs *cadvisorapiv2.FsInfo) *statsapi.FsStats {
	return &statsapi.FsStats{
		Time:           metav1.NewTime(cstat.Timestamp),
		AvailableBytes: &imageFs.Available,
		CapacityBytes:  &imageFs.Capacity,
		InodesFree:     imageFs.InodesFree,
		Inodes:         imageFs.Inodes,
	}
}

// cadvisorInfoToUserDefinedMetrics returns the statsapi.UserDefinedMetric
// converted from the container info from cadvisor.
func cadvisorInfoToUserDefinedMetrics(info *cadvisorapiv2.ContainerInfo) []statsapi.UserDefinedMetric {
	type specVal struct {
		ref     statsapi.UserDefinedMetricDescriptor
		valType cadvisorapiv1.DataType
		time    time.Time
		value   float64
	}
	udmMap := map[string]*specVal{}
	for _, spec := range info.Spec.CustomMetrics {
		udmMap[spec.Name] = &specVal{
			ref: statsapi.UserDefinedMetricDescriptor{
				Name:  spec.Name,
				Type:  statsapi.UserDefinedMetricType(spec.Type),
				Units: spec.Units,
			},
			valType: spec.Format,
		}
	}
	for _, stat := range info.Stats {
		for name, values := range stat.CustomMetrics {
			specVal, ok := udmMap[name]
			if !ok {
				klog.Warningf("spec for custom metric %q is missing from cAdvisor output. Spec: %+v, Metrics: %+v", name, info.Spec, stat.CustomMetrics)
				continue
			}
			for _, value := range values {
				// Pick the most recent value
				if value.Timestamp.Before(specVal.time) {
					continue
				}
				specVal.time = value.Timestamp
				specVal.value = value.FloatValue
				if specVal.valType == cadvisorapiv1.IntType {
					specVal.value = float64(value.IntValue)
				}
			}
		}
	}
	var udm []statsapi.UserDefinedMetric
	for _, specVal := range udmMap {
		udm = append(udm, statsapi.UserDefinedMetric{
			UserDefinedMetricDescriptor: specVal.ref,
			Time:                        metav1.NewTime(specVal.time),
			Value:                       specVal.value,
		})
	}
	return udm
}

// cadvisorInfoToContainerStats returns the statsapi.ContainerStats converted
// from the container and filesystem info.
func cadvisorInfoToContainerStats(name string, info *cadvisorapiv2.ContainerInfo, rootFs, imageFs *cadvisorapiv2.FsInfo) *statsapi.ContainerStats {
	result := &statsapi.ContainerStats{
		StartTime: metav1.NewTime(info.Spec.CreationTime),
		Name:      name,
	}
	cstat, found := latestContainerStats(info)
	if !found {
		return result
	}

	cpu, memory := cadvisorInfoToCPUandMemoryStats(info)
	result.CPU = cpu
	result.Memory = memory

	if rootFs != nil {
		// The container logs live on the node rootfs device
		result.Logs = buildLogsStats(cstat, rootFs)
	}

	if imageFs != nil {
		// The container rootFs lives on the imageFs devices (which may not be the node root fs)
		result.Rootfs = buildRootfsStats(cstat, imageFs)
	}

	cfs := cstat.Filesystem
	if cfs != nil {
		if cfs.BaseUsageBytes != nil {
			if result.Rootfs != nil {
				rootfsUsage := *cfs.BaseUsageBytes
				result.Rootfs.UsedBytes = &rootfsUsage
			}
			if cfs.TotalUsageBytes != nil && result.Logs != nil {
				logsUsage := *cfs.TotalUsageBytes - *cfs.BaseUsageBytes
				result.Logs.UsedBytes = &logsUsage
			}
		}
		if cfs.InodeUsage != nil && result.Rootfs != nil {
			rootInodes := *cfs.InodeUsage
			result.Rootfs.InodesUsed = &rootInodes
		}
	}

	for _, acc := range cstat.Accelerators {
		result.Accelerators = append(result.Accelerators, statsapi.AcceleratorStats{
			Make:        acc.Make,
			Model:       acc.Model,
			ID:          acc.ID,
			MemoryTotal: acc.MemoryTotal,
			MemoryUsed:  acc.MemoryUsed,
			DutyCycle:   acc.DutyCycle,
		})
	}

	result.UserDefinedMetrics = cadvisorInfoToUserDefinedMetrics(info)

	return result
}

func cadvisorInfoToProcessStats(info *cadvisorapiv2.ContainerInfo) *statsapi.ProcessStats {
	cstat, found := latestContainerStats(info)
	if !found || cstat.Processes == nil {
		return nil
	}
	num := cstat.Processes.ProcessCount

	return &statsapi.ProcessStats{ProcessCount: uint64Ptr(num)}
}

func calcEphemeralStorage(containers []statsapi.ContainerStats, volumes []statsapi.VolumeStats, rootFsInfo *cadvisorapiv2.FsInfo,
	podLogStats *statsapi.FsStats, isCRIStatsProvider bool) *statsapi.FsStats {
	result := &statsapi.FsStats{
		Time:           metav1.NewTime(rootFsInfo.Timestamp),
		AvailableBytes: &rootFsInfo.Available,
		CapacityBytes:  &rootFsInfo.Capacity,
		InodesFree:     rootFsInfo.InodesFree,
		Inodes:         rootFsInfo.Inodes,
	}
	for _, container := range containers {
		addContainerUsage(result, &container, isCRIStatsProvider)
	}
	for _, volume := range volumes {
		result.UsedBytes = addUsage(result.UsedBytes, volume.FsStats.UsedBytes)
		result.InodesUsed = addUsage(result.InodesUsed, volume.InodesUsed)
		result.Time = maxUpdateTime(&result.Time, &volume.FsStats.Time)
	}
	if podLogStats != nil {
		result.UsedBytes = addUsage(result.UsedBytes, podLogStats.UsedBytes)
		result.InodesUsed = addUsage(result.InodesUsed, podLogStats.InodesUsed)
		result.Time = maxUpdateTime(&result.Time, &podLogStats.Time)
	}
	return result
}

func addContainerUsage(stat *statsapi.FsStats, container *statsapi.ContainerStats, isCRIStatsProvider bool) {
	if rootFs := container.Rootfs; rootFs != nil {
		stat.Time = maxUpdateTime(&stat.Time, &rootFs.Time)
		stat.InodesUsed = addUsage(stat.InodesUsed, rootFs.InodesUsed)
		stat.UsedBytes = addUsage(stat.UsedBytes, rootFs.UsedBytes)
		if logs := container.Logs; logs != nil {
			stat.UsedBytes = addUsage(stat.UsedBytes, logs.UsedBytes)
			// We have accurate container log inode usage for CRI stats provider.
			if isCRIStatsProvider {
				stat.InodesUsed = addUsage(stat.InodesUsed, logs.InodesUsed)
			}
			stat.Time = maxUpdateTime(&stat.Time, &logs.Time)
		}
	}
}

func addUsage(first, second *uint64) *uint64 {
	if first == nil {
		return second
	} else if second == nil {
		return first
	}
	total := *first + *second
	return &total
}

func maxUpdateTime(first, second *metav1.Time) metav1.Time {
	if first.Before(second) {
		return *second
	}
	return *first
}
