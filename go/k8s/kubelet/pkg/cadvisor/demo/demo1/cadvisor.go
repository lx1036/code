package main

import (
	"k8s.io/klog/v2"
	"net/http"
	"time"

	"github.com/google/cadvisor/cache/memory"
	cadvisormetrics "github.com/google/cadvisor/container"
	"github.com/google/cadvisor/manager"
	"github.com/google/cadvisor/utils/sysfs"

	"k8s.io/utils/pointer"
)

const statsCacheDuration = 2 * time.Minute
const maxHousekeepingInterval = 15 * time.Second
const allowDynamicHousekeeping = true

func main() {
	sysFs := sysfs.NewRealSysFs()
	duration := maxHousekeepingInterval
	housekeepingConfig := manager.HouskeepingConfig{
		Interval:     &duration,
		AllowDynamic: pointer.BoolPtr(allowDynamicHousekeeping),
	}

	includedMetrics := cadvisormetrics.MetricSet{
		cadvisormetrics.CpuUsageMetrics:         struct{}{},
		cadvisormetrics.MemoryUsageMetrics:      struct{}{},
		cadvisormetrics.CpuLoadMetrics:          struct{}{},
		cadvisormetrics.DiskIOMetrics:           struct{}{},
		cadvisormetrics.NetworkUsageMetrics:     struct{}{},
		cadvisormetrics.AcceleratorUsageMetrics: struct{}{},
		cadvisormetrics.AppMetrics:              struct{}{},
		cadvisormetrics.ProcessMetrics:          struct{}{},
		cadvisormetrics.DiskUsageMetrics:        struct{}{},
	}

	cgroupRoots := []string{
		"/usr/local",
		//"/kubepods/burstable/pod14ba5d51-7c96-4007-be05-d89330322531",
	}

	m, err := manager.New(memory.New(statsCacheDuration, nil), sysFs, housekeepingConfig, includedMetrics, http.DefaultClient, cgroupRoots, "")
	if err != nil {
		klog.Fatal(err)
	}

	machineInfo, err := m.GetMachineInfo()
	if err != nil {
		klog.Fatal(err)
	}
	klog.Infof("machineInfo %v", machineInfo)

	//fsInfo, err := m.GetFsInfo("/var/lib/kubelet")
	fsInfo, err := m.GetFsInfo("/usr/local")
	if err != nil {
		klog.Fatal(err)
	}

	klog.Infof("fsInfo %v", fsInfo)
}
