package main

import (
	"net/http"
	"time"

	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/cache/memory"
	cadvisormetrics "k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/container"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/manager"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/utils/sysfs"

	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"
)

// The amount of time for which to keep stats in memory.
const statsCacheDuration = 2 * time.Minute
const maxHousekeepingInterval = 15 * time.Second
const defaultHousekeepingInterval = 10 * time.Second
const allowDynamicHousekeeping = true

func main() {
	sysFs := sysfs.NewRealSysFs()

	includedMetrics := cadvisormetrics.MetricSet{
		cadvisormetrics.CpuUsageMetrics:         struct{}{},
		cadvisormetrics.MemoryUsageMetrics:      struct{}{},
		cadvisormetrics.CpuLoadMetrics:          struct{}{},
		cadvisormetrics.DiskIOMetrics:           struct{}{},
		cadvisormetrics.NetworkUsageMetrics:     struct{}{},
		cadvisormetrics.AcceleratorUsageMetrics: struct{}{},
		cadvisormetrics.AppMetrics:              struct{}{},
		cadvisormetrics.ProcessMetrics:          struct{}{},
	}
	duration := maxHousekeepingInterval
	housekeepingConfig := manager.HouskeepingConfig{
		Interval:     &duration,
		AllowDynamic: pointer.BoolPtr(allowDynamicHousekeeping),
	}
	// Create the cAdvisor container manager.
	cgroupRoots := []string{"/kubepods"}
	m, err := manager.New(memory.New(statsCacheDuration, nil),
		sysFs, housekeepingConfig, includedMetrics, http.DefaultClient, cgroupRoots, "")
	if err != nil {
		klog.Error(err)
		return
	}

	m.Start()

	machineInfo, err := m.GetMachineInfo()
	if err != nil {
		klog.Error(err)
		return
	}

	klog.Infof("machineInfo: %v", machineInfo)
}
