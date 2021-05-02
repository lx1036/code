package main

import (
	"net/http"
	"testing"
	"time"

	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/cache/memory"
	cadvisormetrics "k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/container"
	_ "k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/container/docker/install"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/events"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/fs"
	v1 "k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/info/v1"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/machine"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/manager"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/utils/sysfs/fakesysfs"
	"k8s-lx1036/k8s/kubelet/runc/libcontainer/cgroups"

	"k8s.io/utils/pointer"
)

// The amount of time for which to keep stats in memory.
const statsCacheDuration = 2 * time.Minute
const maxHousekeepingInterval = 15 * time.Second
const defaultHousekeepingInterval = 10 * time.Second
const allowDynamicHousekeeping = true

func init() {
	fs.SetFixturesMountInfoPath("../pkg/fixtures/proc/self/mountinfo")
	fs.SetFixturesDiskstatsPath("../pkg/fixtures/proc/diskstats")
	machine.SetFixturesCPUInfoPath("../pkg/fixtures/proc/cpuinfo")
	machine.SetFixturesMemInfoPath("../pkg/fixtures/proc/meminfo")
	cgroups.SetFixturesMountInfoPath("../pkg/fixtures/proc/self/mountinfo")
	cgroups.SetFixturesCgroupPath("../pkg/fixtures/proc/self/cgroup")
}

func TestCadvisorEvents(test *testing.T) {
	stopCh := make(chan struct{})

	sysfs := &fakesysfs.FakeSysFs{}
	duration := maxHousekeepingInterval
	housekeepingConfig := manager.HouskeepingConfig{
		Interval:     &duration,
		AllowDynamic: pointer.BoolPtr(allowDynamicHousekeeping),
	}
	includedMetrics := cadvisormetrics.MetricSet{
		cadvisormetrics.CpuUsageMetrics:    struct{}{},
		cadvisormetrics.MemoryUsageMetrics: struct{}{},
		cadvisormetrics.CpuLoadMetrics:     struct{}{},
	}
	containerManager, err := manager.New(memory.New(statsCacheDuration, nil), sysfs, housekeepingConfig,
		includedMetrics, http.DefaultClient, nil, "")
	if err != nil {
		panic(err)
	}

	err = containerManager.Start()
	if err != nil {
		panic(err)
	}

	req := events.NewRequest()
	req.IncludeSubcontainers = true
	req.MaxEventsReturned = -1
	req.EventType = map[v1.EventType]bool{
		v1.EventOom:     true,
		v1.EventOomKill: true,
	}
	req.EventType[v1.EventContainerCreation] = true
	req.EventType[v1.EventContainerDeletion] = true
	evts, err := containerManager.GetPastEvents(req)
	if err != nil {
		panic(err)
	}
	result := make([]v1.Event, len(evts))
	for i, evt := range evts {
		result[i] = *evt
	}

	<-stopCh
}
