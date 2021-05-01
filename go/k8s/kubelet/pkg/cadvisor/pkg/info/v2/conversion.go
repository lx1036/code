package v2

import (
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/info/v1"

	"k8s.io/klog/v2"
)

// Get V2 container spec from v1 container info.
func ContainerSpecFromV1(specV1 *v1.ContainerSpec, aliases []string, namespace string) ContainerSpec {
	specV2 := ContainerSpec{
		CreationTime:     specV1.CreationTime,
		HasCpu:           specV1.HasCpu,
		HasMemory:        specV1.HasMemory,
		HasHugetlb:       specV1.HasHugetlb,
		HasFilesystem:    specV1.HasFilesystem,
		HasNetwork:       specV1.HasNetwork,
		HasProcesses:     specV1.HasProcesses,
		HasDiskIo:        specV1.HasDiskIo,
		HasCustomMetrics: specV1.HasCustomMetrics,
		Image:            specV1.Image,
		Labels:           specV1.Labels,
		Envs:             specV1.Envs,
	}
	if specV1.HasCpu {
		specV2.Cpu.Limit = specV1.Cpu.Limit
		specV2.Cpu.MaxLimit = specV1.Cpu.MaxLimit
		specV2.Cpu.Mask = specV1.Cpu.Mask
	}
	if specV1.HasMemory {
		specV2.Memory.Limit = specV1.Memory.Limit
		specV2.Memory.Reservation = specV1.Memory.Reservation
		specV2.Memory.SwapLimit = specV1.Memory.SwapLimit
	}
	if specV1.HasCustomMetrics {
		specV2.CustomMetrics = specV1.CustomMetrics
	}
	specV2.Aliases = aliases
	specV2.Namespace = namespace
	return specV2
}

func ContainerStatsFromV1(containerName string, spec *v1.ContainerSpec, stats []*v1.ContainerStats) []*ContainerStats {
	newStats := make([]*ContainerStats, 0, len(stats))
	//var last *v1.ContainerStats
	for _, val := range stats {
		stat := &ContainerStats{
			Timestamp:        val.Timestamp,
			ReferencedMemory: val.ReferencedMemory,
		}
		if spec.HasCpu {
			stat.Cpu = &val.Cpu
			/*cpuInst, err := InstCpuStats(last, val)
			if err != nil {
				klog.Warningf("Could not get instant cpu stats: %v", err)
			} else {
				stat.CpuInst = cpuInst
			}*/
			//last = val
		}
		if spec.HasMemory {
			stat.Memory = &val.Memory
		}
		if spec.HasHugetlb {
			//stat.Hugetlb = &val.Hugetlb
		}
		if spec.HasNetwork {
			// TODO: Handle TcpStats
			/*stat.Network = &NetworkStats{
				Tcp:        TcpStat(val.Network.Tcp),
				Tcp6:       TcpStat(val.Network.Tcp6),
				Interfaces: val.Network.Interfaces,
			}*/
		}
		if spec.HasProcesses {
			//stat.Processes = &val.Processes
		}
		if spec.HasFilesystem {
			if len(val.Filesystem) == 1 {
				stat.Filesystem = &FilesystemStats{
					TotalUsageBytes: &val.Filesystem[0].Usage,
					BaseUsageBytes:  &val.Filesystem[0].BaseUsage,
					InodeUsage:      &val.Filesystem[0].Inodes,
				}
			} else if len(val.Filesystem) > 1 && containerName != "/" {
				// Cannot handle multiple devices per container.
				klog.V(4).Infof("failed to handle multiple devices for container %s. Skipping Filesystem stats", containerName)
			}
		}
		if spec.HasDiskIo {
			stat.DiskIo = &val.DiskIo
		}
		if spec.HasCustomMetrics {
			//stat.CustomMetrics = val.CustomMetrics
		}
		/*if len(val.Accelerators) > 0 {
			stat.Accelerators = val.Accelerators
		}
		if len(val.PerfStats) > 0 {
			stat.PerfStats = val.PerfStats
		}
		if len(val.PerfUncoreStats) > 0 {
			stat.PerfUncoreStats = val.PerfUncoreStats
		}
		if len(val.Resctrl.MemoryBandwidth) > 0 || len(val.Resctrl.Cache) > 0 {
			stat.Resctrl = val.Resctrl
		}*/
		// TODO(rjnagal): Handle load stats.
		newStats = append(newStats, stat)
	}
	return newStats
}
