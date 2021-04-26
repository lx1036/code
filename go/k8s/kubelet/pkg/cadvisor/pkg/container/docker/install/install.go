package install

import (
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/container"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/container/docker"

	"k8s.io/klog/v2"
)

func init() {
	err := container.RegisterPlugin("docker", docker.NewPlugin())
	if err != nil {
		klog.Fatalf("Failed to register docker plugin: %v", err)
	}
}
