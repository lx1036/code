package main

import (
	"flag"
	"fmt"
	"testing"

	clientv1 "github.com/google/cadvisor/client"
	clientv2 "github.com/google/cadvisor/client/v2"
	info "github.com/google/cadvisor/info/v1"

	"k8s.io/klog/v2"
)

var (
	podIP = flag.String("podip", "", "cadvisor pod ip")
)

// INFO: 直接读取 cadvisor pod api 获取 machineInfo 等数据
// go run . --podip=
func TestAdvisorClient(test *testing.T) {
	flag.Parse()

	cadvisorPodIP := *podIP
	if len(cadvisorPodIP) == 0 {
		klog.Errorf("cadvisor pod ip is required")
		return
	}

	client, err := clientv2.NewClient(fmt.Sprintf("http://%s:8080/", cadvisorPodIP))
	if err != nil {
		panic(err)
	}

	machineInfo, err := client.MachineInfo()
	if err != nil {
		panic(err)
	}

	klog.Infof(`machineInfo: NumCores: %d, NumPhysicalCores: %d, NumSockets: %d, MemoryCapacity: %d,
		MachineID: %s, Topology Core: %v,`, machineInfo.NumCores, machineInfo.NumPhysicalCores,
		machineInfo.NumSockets, machineInfo.MemoryCapacity, machineInfo.MachineID, machineInfo.Topology[0].Cores[0])

	version, err := client.VersionInfo()
	if err != nil {
		panic(err)
	}
	klog.Infof("version: %s", version)

	stats, err := client.MachineStats()
	if err != nil {
		panic(err)
	}
	klog.Infof("cpu stats: %v", stats[0].Cpu.Usage)

	attributes, err := client.Attributes()
	if err != nil {
		panic(err)
	}
	klog.Infof("attributes KernelVersion: %v", attributes.KernelVersion)

	clientV1, err := clientv1.NewClient(fmt.Sprintf("http://%s:8080/", cadvisorPodIP))
	if err != nil {
		panic(err)
	}

	eventsInfo, err := clientV1.EventStaticInfo("?oom_events=true")
	if err != nil {
		panic(err)
	}
	for idx, event := range eventsInfo {
		klog.Infof("static einfo %v: %v", idx, event)
	}

	// INFO: 获取该容器的 stats, 其实就是调用 manager.GetContainerInfo(containerName string, query *v1.ContainerInfoRequest)
	containerInfo, err := clientV1.DockerContainer("0e8b25a584ce27c6c88a59d9411cafc6ac82bd90ee67ccaead109ffbccd46cf4", &info.ContainerInfoRequest{})
	if err != nil {
		panic(err)
	}
	for _, stat := range containerInfo.Stats {
		klog.Infof("container id %s, cpu usage: %v", containerInfo.Id, stat.Cpu)
	}

	// watch oom events
	eventInfo := make(chan *info.Event)
	go func() {
		err = clientV1.EventStreamingInfo("?creation_events=true&stream=true&oom_events=true&deletion_events=true", eventInfo)
		if err != nil {
			klog.Errorf("got error retrieving event info: %v", err)
			return
		}
	}()
	for ev := range eventInfo {
		klog.Infof("streaming einfo: %v\n", ev)
	}
}
