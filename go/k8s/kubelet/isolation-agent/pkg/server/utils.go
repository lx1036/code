package server

import (
	v1 "k8s.io/api/core/v1"
	v1qos "k8s.io/kubernetes/pkg/apis/core/v1/helper/qos"
)

// INFO: @see https://github.com/kubernetes/kubernetes/blob/release-1.19/pkg/kubelet/kubelet_pods.go#L96-L101
func GetActivePods(allPods []*v1.Pod) []*v1.Pod {
	activePods := filterOutTerminatedPods(allPods)
	return activePods
}

// filterOutTerminatedPods returns the given pods which the status manager
// does not consider failed or succeeded.
func filterOutTerminatedPods(pods []*v1.Pod) []*v1.Pod {
	var filteredPods []*v1.Pod
	for _, p := range pods {
		if podIsTerminated(p) {
			continue
		}
		filteredPods = append(filteredPods, p)
	}
	return filteredPods
}

// podIsTerminated returns true if the provided pod is in a terminal phase ("Failed", "Succeeded") or
// has been deleted and has no running containers. This corresponds to when a pod must accept changes to
// its pod spec (e.g. terminating containers allow grace period to be shortened).
func podIsTerminated(pod *v1.Pod) bool {
	_, podWorkerTerminal := podAndContainersAreTerminal(pod)
	return podWorkerTerminal
}

// podStatusIsTerminal reports when the specified pod has no running containers or is no longer accepting
// spec changes.
func podAndContainersAreTerminal(pod *v1.Pod) (containersTerminal, podWorkerTerminal bool) {
	status := pod.Status

	// A pod transitions into failed or succeeded from either container lifecycle (RestartNever container
	// fails) or due to external events like deletion or eviction. A terminal pod *should* have no running
	// containers, but to know that the pod has completed its lifecycle you must wait for containers to also
	// be terminal.
	containersTerminal = notRunning(status.ContainerStatuses)
	// The kubelet must accept config changes from the pod spec until it has reached a point where changes would
	// have no effect on any running container.
	podWorkerTerminal = status.Phase == v1.PodFailed || status.Phase == v1.PodSucceeded || (pod.DeletionTimestamp != nil && containersTerminal)
	return
}

// notRunning returns true if every status is terminated or waiting, or the status list
// is empty.
func notRunning(statuses []v1.ContainerStatus) bool {
	for _, status := range statuses {
		if status.State.Terminated == nil && status.State.Waiting == nil {
			return false
		}
	}
	return true
}

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
