package main

import (
	"net/http"
	"testing"
	"time"

	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/cache/memory"
	cadvisormetrics "k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/container"
	_ "k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/container/docker/install"
	v1 "k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/info/v1"
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

func TestManagerStart(test *testing.T) {
	sysFs := sysfs.NewRealSysFs()

	includedMetrics := cadvisormetrics.MetricSet{
		cadvisormetrics.CpuUsageMetrics: struct{}{},
		//cadvisormetrics.MemoryUsageMetrics:      struct{}{},
		//cadvisormetrics.CpuLoadMetrics:          struct{}{},
		//cadvisormetrics.DiskIOMetrics:           struct{}{},
		//cadvisormetrics.NetworkUsageMetrics:     struct{}{},
		//cadvisormetrics.AcceleratorUsageMetrics: struct{}{},
		//cadvisormetrics.AppMetrics:              struct{}{},
		//cadvisormetrics.ProcessMetrics:          struct{}{},
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

	err = m.Start()
	if err != nil {
		klog.Error(err)
		return
	}

	machineInfo, err := m.GetMachineInfo()
	if err != nil {
		klog.Error(err)
		return
	}

	klog.Infof("machineInfo: %v", machineInfo)

	// INFO: 获取该容器的 stats, 其实就是调用 manager.GetContainerInfo(containerName string, query *v1.ContainerInfoRequest)
	containerInfo, err := m.GetContainerInfo("0e8b25a584ce27c6c88a59d9411cafc6ac82bd90ee67ccaead109ffbccd46cf4", &v1.ContainerInfoRequest{})
	if err != nil {
		panic(err)
	}
	for _, stat := range containerInfo.Stats {
		klog.Infof("container id %s, cpu usage: %v", containerInfo.Id, stat.Cpu)
	}
}
