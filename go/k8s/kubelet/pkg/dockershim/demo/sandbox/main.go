package main

import (
	"fmt"
	"time"

	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/kubelet/cri/remote"
)

func main() {
	remoteRuntimeEndpoint := "unix:///var/run/dockershim.sock"
	remoteRuntimeService, err := remote.NewRemoteRuntimeService(remoteRuntimeEndpoint, time.Minute*2)
	if err != nil {
		panic(err)
	}

	podsSandbox, err := remoteRuntimeService.ListPodSandbox(&runtimeapi.PodSandboxFilter{})
	if err != nil {
		panic(err)
	}
	for _, sandbox := range podsSandbox {
		// pod name default/cgroup1-75cb7bc8c5-vbzww sandbox id: af88d9ab332141c168be34dc49752787d0640f0877b181d94077e24fcf9de497
		klog.Info(fmt.Sprintf("pod name %s/%s sandbox id: %s", sandbox.Metadata.Namespace, sandbox.Metadata.Name, sandbox.Id))

		podSandboxStatus, err := remoteRuntimeService.PodSandboxStatus(sandbox.Id)
		if err != nil {
			klog.Error(err)
		}

		// podSandboxStatus Network: &PodSandboxNetworkStatus{Ip:10.129.114.11,AdditionalIps:[]*PodIP{},}
		klog.Info(fmt.Sprintf("podSandboxStatus Network: %v", podSandboxStatus.Network))
	}
}
