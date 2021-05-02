package main

import (
	"fmt"
	"testing"

	"k8s-lx1036/k8s/kubelet/pkg/cadvisor"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/fs"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/machine"
	kubelettypes "k8s-lx1036/k8s/kubelet/pkg/types"

	"k8s.io/klog/v2"
)

func TestKubeletCadvisorClient(test *testing.T) {
	// RootFsInfo ImagesFsInfo ContainerInfoV2

	runtime := kubelettypes.DockerContainerRuntime
	runtimeEndpoint := "unix:///var/run/docker.sock"
	imageFsInfoProvider := cadvisor.NewImageFsInfoProvider(runtime, runtimeEndpoint)
	rootPath := "/tmp/cadvisor"
	cgroupRoots := []string{"/kubepods"}

	fs.SetFixturesMountInfoPath("../pkg/fixtures/proc/self/mountinfo")
	fs.SetFixturesDiskstatsPath("../pkg/fixtures/proc/diskstats")
	machine.SetFixturesCPUInfoPath("../pkg/fixtures/proc/cpuinfo")
	machine.SetFixturesMemInfoPath("../pkg/fixtures/proc/meminfo")
	cadvisorClient, err := cadvisor.New(imageFsInfoProvider, rootPath, cgroupRoots, true)
	if err != nil {
		panic(err)
	}

	rootFsInfo, err := cadvisorClient.RootFsInfo()
	if err != nil {
		panic(err)
	}

	klog.Info(fmt.Sprintf("rootPath %s fileysystem info: %v", rootPath, rootFsInfo))
}
