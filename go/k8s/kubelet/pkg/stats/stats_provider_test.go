package stats

import (
	"time"

	statsapi "k8s-lx1036/k8s/kubelet/pkg/apis/stats/v1alpha1"
	cadvisorapiv1 "k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/info/v1"
	cadvisorapiv2 "k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/info/v2"
	serverstats "k8s-lx1036/k8s/kubelet/pkg/server/stats"

	fuzz "github.com/google/gofuzz"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type fakeResourceAnalyzer struct {
	podVolumeStats serverstats.PodVolumeStats
}

func (o *fakeResourceAnalyzer) Start()                                           {}
func (o *fakeResourceAnalyzer) Get(bool) (*statsapi.Summary, error)              { return nil, nil }
func (o *fakeResourceAnalyzer) GetCPUAndMemoryStats() (*statsapi.Summary, error) { return nil, nil }
func (o *fakeResourceAnalyzer) GetPodVolumeStats(uid types.UID) (serverstats.PodVolumeStats, bool) {
	return o.podVolumeStats, true
}

const (
	// Offsets from seed value in generated container stats.
	offsetCPUUsageCores = iota
	offsetCPUUsageCoreSeconds
	offsetMemPageFaults
	offsetMemMajorPageFaults
	offsetMemUsageBytes
	offsetMemRSSBytes
	offsetMemWorkingSetBytes
	offsetNetRxBytes
	offsetNetRxErrors
	offsetNetTxBytes
	offsetNetTxErrors
	offsetFsCapacity
	offsetFsAvailable
	offsetFsUsage
	offsetFsInodes
	offsetFsInodesFree
	offsetFsTotalUsageBytes
	offsetFsBaseUsageBytes
	offsetFsInodeUsage
	offsetAcceleratorDutyCycle
)

var (
	timestamp    = time.Now()
	creationTime = timestamp.Add(-5 * time.Minute)
)

func testTime(base time.Time, seed int) time.Time {
	return base.Add(time.Duration(seed) * time.Second)
}

func generateCustomMetricSpec() []cadvisorapiv1.MetricSpec {
	f := fuzz.New().NilChance(0).Funcs(
		func(e *cadvisorapiv1.MetricSpec, c fuzz.Continue) {
			c.Fuzz(&e.Name)
			switch c.Intn(3) {
			case 0:
				e.Type = cadvisorapiv1.MetricGauge
			case 1:
				e.Type = cadvisorapiv1.MetricCumulative
			case 2:
				e.Type = cadvisorapiv1.MetricType("delta")
			}
			switch c.Intn(2) {
			case 0:
				e.Format = cadvisorapiv1.IntType
			case 1:
				e.Format = cadvisorapiv1.FloatType
			}
			c.Fuzz(&e.Units)
		})
	var ret []cadvisorapiv1.MetricSpec
	f.Fuzz(&ret)
	return ret
}

func generateCustomMetrics(spec []cadvisorapiv1.MetricSpec) map[string][]cadvisorapiv1.MetricVal {
	ret := map[string][]cadvisorapiv1.MetricVal{}
	for _, metricSpec := range spec {
		f := fuzz.New().NilChance(0).Funcs(
			func(e *cadvisorapiv1.MetricVal, c fuzz.Continue) {
				switch metricSpec.Format {
				case cadvisorapiv1.IntType:
					c.Fuzz(&e.IntValue)
				case cadvisorapiv1.FloatType:
					c.Fuzz(&e.FloatValue)
				}
			})

		var metrics []cadvisorapiv1.MetricVal
		f.Fuzz(&metrics)
		ret[metricSpec.Name] = metrics
	}
	return ret
}

func getTestContainerInfo(seed int, podName string, podNamespace string, containerName string) cadvisorapiv2.ContainerInfo {
	labels := map[string]string{}
	if podName != "" {
		labels = map[string]string{
			"io.kubernetes.pod.name":       podName,
			"io.kubernetes.pod.uid":        "UID" + podName,
			"io.kubernetes.pod.namespace":  podNamespace,
			"io.kubernetes.container.name": containerName,
		}
	}

	// by default, kernel will set memory.limit_in_bytes to 1 << 63 if not bounded
	unlimitedMemory := uint64(1 << 63)
	spec := cadvisorapiv2.ContainerSpec{
		CreationTime: testTime(creationTime, seed),
		HasCpu:       true,
		HasMemory:    true,
		HasNetwork:   true,
		Labels:       labels,
		Memory: cadvisorapiv2.MemorySpec{
			Limit: unlimitedMemory,
		},
		CustomMetrics: generateCustomMetricSpec(),
	}

	totalUsageBytes := uint64(seed + offsetFsTotalUsageBytes)
	baseUsageBytes := uint64(seed + offsetFsBaseUsageBytes)
	inodeUsage := uint64(seed + offsetFsInodeUsage)

	stats := cadvisorapiv2.ContainerStats{
		Timestamp: testTime(timestamp, seed),
		Cpu:       &cadvisorapiv1.CpuStats{},
		CpuInst:   &cadvisorapiv2.CpuInstStats{},
		Memory: &cadvisorapiv1.MemoryStats{
			Usage:      uint64(seed + offsetMemUsageBytes),
			WorkingSet: uint64(seed + offsetMemWorkingSetBytes),
			RSS:        uint64(seed + offsetMemRSSBytes),
			ContainerData: cadvisorapiv1.MemoryStatsMemoryData{
				Pgfault:    uint64(seed + offsetMemPageFaults),
				Pgmajfault: uint64(seed + offsetMemMajorPageFaults),
			},
		},
		Network: &cadvisorapiv2.NetworkStats{
			Interfaces: []cadvisorapiv1.InterfaceStats{{
				Name:     "eth0",
				RxBytes:  uint64(seed + offsetNetRxBytes),
				RxErrors: uint64(seed + offsetNetRxErrors),
				TxBytes:  uint64(seed + offsetNetTxBytes),
				TxErrors: uint64(seed + offsetNetTxErrors),
			}, {
				Name:     "cbr0",
				RxBytes:  100,
				RxErrors: 100,
				TxBytes:  100,
				TxErrors: 100,
			}},
		},
		CustomMetrics: generateCustomMetrics(spec.CustomMetrics),
		Filesystem: &cadvisorapiv2.FilesystemStats{
			TotalUsageBytes: &totalUsageBytes,
			BaseUsageBytes:  &baseUsageBytes,
			InodeUsage:      &inodeUsage,
		},
		Accelerators: []cadvisorapiv1.AcceleratorStats{
			{
				Make:        "nvidia",
				Model:       "Tesla K80",
				ID:          "foobar",
				MemoryTotal: uint64(seed + offsetMemUsageBytes),
				MemoryUsed:  uint64(seed + offsetMemUsageBytes),
				DutyCycle:   uint64(seed + offsetAcceleratorDutyCycle),
			},
		},
	}

	stats.Cpu.Usage.Total = uint64(seed + offsetCPUUsageCoreSeconds)
	stats.CpuInst.Usage.Total = uint64(seed + offsetCPUUsageCores)
	return cadvisorapiv2.ContainerInfo{
		Spec:  spec,
		Stats: []*cadvisorapiv2.ContainerStats{&stats},
	}
}

func getPodVolumeStats(seed int, volumeName string) statsapi.VolumeStats {
	availableBytes := uint64(seed + offsetFsAvailable)
	capacityBytes := uint64(seed + offsetFsCapacity)
	usedBytes := uint64(seed + offsetFsUsage)
	inodes := uint64(seed + offsetFsInodes)
	inodesFree := uint64(seed + offsetFsInodesFree)
	inodesUsed := uint64(seed + offsetFsInodeUsage)
	fsStats := statsapi.FsStats{
		Time:           metav1.NewTime(time.Now()),
		AvailableBytes: &availableBytes,
		CapacityBytes:  &capacityBytes,
		UsedBytes:      &usedBytes,
		Inodes:         &inodes,
		InodesFree:     &inodesFree,
		InodesUsed:     &inodesUsed,
	}

	return statsapi.VolumeStats{
		FsStats: fsStats,
		Name:    volumeName,
	}
}
