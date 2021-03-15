package stats

import (
	kubecontainer "k8s.io/kubernetes/pkg/kubelet/container"
	kubepod "k8s.io/kubernetes/pkg/kubelet/pod"
)

// Provider provides the stats of the node and the pod-managed containers.
type Provider struct {
	cadvisor     cadvisor.Interface
	podManager   kubepod.Manager
	runtimeCache kubecontainer.RuntimeCache
	containerStatsProvider
	rlimitStatsProvider
}
