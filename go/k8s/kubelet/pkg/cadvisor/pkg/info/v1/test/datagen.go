package test

import (
	"fmt"
	"math/rand"
	"time"

	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/info/v1"
)

// INFO: 随机生成 container stats
func GenerateRandomContainerSpec(numCores int) v1.ContainerSpec {
	ret := v1.ContainerSpec{
		CreationTime: time.Now(),
		HasCpu:       true,
		Cpu:          v1.CpuSpec{},
		HasMemory:    true,
		Memory:       v1.MemorySpec{},
	}
	ret.Cpu.Limit = uint64(1000 + rand.Int63n(2000))
	ret.Cpu.MaxLimit = uint64(1000 + rand.Int63n(2000))
	ret.Cpu.Mask = fmt.Sprintf("0-%d", numCores-1)
	ret.Memory.Limit = uint64(4096 + rand.Int63n(4096))
	return ret
}

func GenerateRandomStats(numStats, numCores int, duration time.Duration) []*v1.ContainerStats {
	ret := make([]*v1.ContainerStats, numStats)
	perCoreUsages := make([]uint64, numCores)
	currentTime := time.Now()
	for i := range perCoreUsages {
		perCoreUsages[i] = uint64(rand.Int63n(1000))
	}
	for i := 0; i < numStats; i++ {
		stats := new(v1.ContainerStats)
		stats.Timestamp = currentTime
		currentTime = currentTime.Add(duration)

		percore := make([]uint64, numCores)
		for i := range perCoreUsages {
			perCoreUsages[i] += uint64(rand.Int63n(1000))
			percore[i] = perCoreUsages[i]
			stats.Cpu.Usage.Total += percore[i]
		}
		stats.Cpu.Usage.PerCpu = percore
		stats.Cpu.Usage.User = stats.Cpu.Usage.Total
		stats.Cpu.Usage.System = 0
		stats.Memory.Usage = uint64(rand.Int63n(4096))
		stats.Memory.Cache = uint64(rand.Int63n(4096))
		stats.Memory.RSS = uint64(rand.Int63n(4096))
		stats.Memory.MappedFile = uint64(rand.Int63n(4096))
		stats.ReferencedMemory = uint64(rand.Int63n(1000))
		ret[i] = stats
	}

	return ret
}

func GenerateRandomContainerInfo(containerName string, numCores int, query *v1.ContainerInfoRequest, duration time.Duration) *v1.ContainerInfo {
	stats := GenerateRandomStats(query.NumStats, numCores, duration)
	spec := GenerateRandomContainerSpec(numCores)

	ret := &v1.ContainerInfo{
		ContainerReference: v1.ContainerReference{
			Name: containerName,
		},
		Spec:  spec,
		Stats: stats,
	}

	return ret
}
