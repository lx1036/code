package topologymanager

import (
	"k8s-lx1036/k8s/kubelet/pkg/lifecycle"

	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

type fakeManager struct{}

func NewFakeManager() Manager {
	klog.Infof("[fake topologymanager] NewFakeManager")
	return &fakeManager{}
}

func (m *fakeManager) GetAffinity(podUID string, containerName string) TopologyHint {
	klog.Infof("[fake topologymanager] GetAffinity podUID: %v container name:  %v", podUID, containerName)
	return TopologyHint{}
}

func (m *fakeManager) AddHintProvider(h HintProvider) {
	klog.Infof("[fake topologymanager] AddHintProvider HintProvider:  %v", h)
}

func (m *fakeManager) AddContainer(pod *v1.Pod, containerID string) error {
	klog.Infof("[fake topologymanager] AddContainer  pod: %v container id:  %v", pod, containerID)
	return nil
}

func (m *fakeManager) RemoveContainer(containerID string) error {
	klog.Infof("[fake topologymanager] RemoveContainer container id:  %v", containerID)
	return nil
}

func (m *fakeManager) Admit(attrs *lifecycle.PodAdmitAttributes) lifecycle.PodAdmitResult {
	klog.Infof("[fake topologymanager] Topology Admit Handler")
	return lifecycle.PodAdmitResult{
		Admit: true,
	}
}
