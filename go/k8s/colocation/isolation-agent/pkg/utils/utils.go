package utils

import (
	v1 "k8s.io/api/core/v1"
	v1qos "k8s.io/kubernetes/pkg/apis/core/v1/helper/qos"
)

// IsProdPod 在线 pod
func IsProdPod(pod *v1.Pod) bool {
	qos := v1qos.GetPodQOS(pod)
	if qos == v1.PodQOSGuaranteed || qos == v1.PodQOSBurstable {
		return true
	}

	return false
}

// IsNonProdPod 离线 pod
func IsNonProdPod(pod *v1.Pod) bool {
	if v1qos.GetPodQOS(pod) == v1.PodQOSBestEffort {
		return true
	}

	return false
}
