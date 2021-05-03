package stats

import (
	"testing"
	"time"

	statsapi "k8s-lx1036/k8s/kubelet/pkg/apis/stats/v1alpha1"
	cadvisorapiv1 "k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/info/v1"
	cadvisorapiv2 "k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/info/v2"
	serverstats "k8s-lx1036/k8s/kubelet/pkg/server/stats"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

func checkCPUStats(t *testing.T, label string, seed int, stats *statsapi.CPUStats) {
	require.NotNil(t, stats.Time, label+".CPU.Time")
	require.NotNil(t, stats.UsageNanoCores, label+".CPU.UsageNanoCores")
	require.NotNil(t, stats.UsageNanoCores, label+".CPU.UsageCoreSeconds")
	assert.EqualValues(t, testTime(timestamp, seed).Unix(), stats.Time.Time.Unix(), label+".CPU.Time")
	assert.EqualValues(t, seed+offsetCPUUsageCores, *stats.UsageNanoCores, label+".CPU.UsageCores")
	assert.EqualValues(t, seed+offsetCPUUsageCoreSeconds, *stats.UsageCoreNanoSeconds, label+".CPU.UsageCoreSeconds")
}

func checkMemoryStats(t *testing.T, label string, seed int, info cadvisorapiv2.ContainerInfo, stats *statsapi.MemoryStats) {
	assert.EqualValues(t, testTime(timestamp, seed).Unix(), stats.Time.Time.Unix(), label+".Mem.Time")
	assert.EqualValues(t, seed+offsetMemUsageBytes, *stats.UsageBytes, label+".Mem.UsageBytes")
	assert.EqualValues(t, seed+offsetMemWorkingSetBytes, *stats.WorkingSetBytes, label+".Mem.WorkingSetBytes")
	assert.EqualValues(t, seed+offsetMemRSSBytes, *stats.RSSBytes, label+".Mem.RSSBytes")
	assert.EqualValues(t, seed+offsetMemPageFaults, *stats.PageFaults, label+".Mem.PageFaults")
	assert.EqualValues(t, seed+offsetMemMajorPageFaults, *stats.MajorPageFaults, label+".Mem.MajorPageFaults")
	if !info.Spec.HasMemory || isMemoryUnlimited(info.Spec.Memory.Limit) {
		assert.Nil(t, stats.AvailableBytes, label+".Mem.AvailableBytes")
	} else {
		expected := info.Spec.Memory.Limit - *stats.WorkingSetBytes
		assert.EqualValues(t, expected, *stats.AvailableBytes, label+".Mem.AvailableBytes")
	}
}

func checkNetworkStats(t *testing.T, label string, seed int, stats *statsapi.NetworkStats) {
	assert.NotNil(t, stats)
	assert.EqualValues(t, testTime(timestamp, seed).Unix(), stats.Time.Time.Unix(), label+".Net.Time")
	assert.EqualValues(t, "eth0", stats.Name, "default interface name is not eth0")
	assert.EqualValues(t, seed+offsetNetRxBytes, *stats.RxBytes, label+".Net.RxBytes")
	assert.EqualValues(t, seed+offsetNetRxErrors, *stats.RxErrors, label+".Net.RxErrors")
	assert.EqualValues(t, seed+offsetNetTxBytes, *stats.TxBytes, label+".Net.TxBytes")
	assert.EqualValues(t, seed+offsetNetTxErrors, *stats.TxErrors, label+".Net.TxErrors")

	assert.EqualValues(t, 2, len(stats.Interfaces), "network interfaces should contain 2 elements")

	assert.EqualValues(t, "eth0", stats.Interfaces[0].Name, "default interface name is not eth0")
	assert.EqualValues(t, seed+offsetNetRxBytes, *stats.Interfaces[0].RxBytes, label+".Net.TxErrors")
	assert.EqualValues(t, seed+offsetNetRxErrors, *stats.Interfaces[0].RxErrors, label+".Net.TxErrors")
	assert.EqualValues(t, seed+offsetNetTxBytes, *stats.Interfaces[0].TxBytes, label+".Net.TxErrors")
	assert.EqualValues(t, seed+offsetNetTxErrors, *stats.Interfaces[0].TxErrors, label+".Net.TxErrors")

	assert.EqualValues(t, "cbr0", stats.Interfaces[1].Name, "cbr0 interface name is not cbr0")
	assert.EqualValues(t, 100, *stats.Interfaces[1].RxBytes, label+".Net.TxErrors")
	assert.EqualValues(t, 100, *stats.Interfaces[1].RxErrors, label+".Net.TxErrors")
	assert.EqualValues(t, 100, *stats.Interfaces[1].TxBytes, label+".Net.TxErrors")
	assert.EqualValues(t, 100, *stats.Interfaces[1].TxErrors, label+".Net.TxErrors")
}

// container which had no stats should have zero-valued CPU usage
func checkEmptyCPUStats(t *testing.T, label string, seed int, stats *statsapi.CPUStats) {
	require.NotNil(t, stats.Time, label+".CPU.Time")
	require.NotNil(t, stats.UsageNanoCores, label+".CPU.UsageNanoCores")
	require.NotNil(t, stats.UsageNanoCores, label+".CPU.UsageCoreSeconds")
	assert.EqualValues(t, testTime(timestamp, seed).Unix(), stats.Time.Time.Unix(), label+".CPU.Time")
	assert.EqualValues(t, 0, *stats.UsageNanoCores, label+".CPU.UsageCores")
	assert.EqualValues(t, 0, *stats.UsageCoreNanoSeconds, label+".CPU.UsageCoreSeconds")
}

// container which had no stats should have zero-valued Memory usage
func checkEmptyMemoryStats(t *testing.T, label string, seed int, info cadvisorapiv2.ContainerInfo, stats *statsapi.MemoryStats) {
	assert.EqualValues(t, testTime(timestamp, seed).Unix(), stats.Time.Time.Unix(), label+".Mem.Time")
	require.NotNil(t, stats.WorkingSetBytes, label+".Mem.WorkingSetBytes")
	assert.EqualValues(t, 0, *stats.WorkingSetBytes, label+".Mem.WorkingSetBytes")
	assert.Nil(t, stats.UsageBytes, label+".Mem.UsageBytes")
	assert.Nil(t, stats.RSSBytes, label+".Mem.RSSBytes")
	assert.Nil(t, stats.PageFaults, label+".Mem.PageFaults")
	assert.Nil(t, stats.MajorPageFaults, label+".Mem.MajorPageFaults")
	assert.Nil(t, stats.AvailableBytes, label+".Mem.AvailableBytes")
}

func checkEphemeralStats(t *testing.T, label string, containerSeeds []int, volumeSeeds []int, stats *statsapi.FsStats) {
	var usedBytes, inodeUsage int
	for _, cseed := range containerSeeds {
		usedBytes = usedBytes + cseed + offsetFsTotalUsageBytes
		inodeUsage += cseed + offsetFsInodeUsage
	}
	for _, vseed := range volumeSeeds {
		usedBytes = usedBytes + vseed + offsetFsUsage
		inodeUsage += vseed + offsetFsInodeUsage
	}
	assert.EqualValues(t, usedBytes, int(*stats.UsedBytes), label+".UsedBytes")
	assert.EqualValues(t, inodeUsage, int(*stats.InodesUsed), label+".InodesUsed")
}
