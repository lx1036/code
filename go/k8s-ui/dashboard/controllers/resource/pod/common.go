package pod

import (
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/common"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/event"
	corev1 "k8s.io/api/core/v1"
)

func getPodStatus(pods *corev1.PodList, events []corev1.Event) common.ResourceStatus {
	info := common.ResourceStatus{}
	for _, pod := range pods.Items {
		warnings := event.GetPodsEventWarnings(events, []corev1.Pod{pod})
		switch getPodStatusPhase(pod, warnings) {
		case corev1.PodFailed:
			info.Failed++
		case corev1.PodSucceeded:
			info.Succeeded++
		case corev1.PodRunning:
			info.Running++
		case corev1.PodPending:
			info.Pending++
		}
	}

	return info
}

func getPodStatusPhase(pod corev1.Pod, warnings []event.Event) corev1.PodPhase {
	// For terminated pods that failed
	if pod.Status.Phase == corev1.PodFailed {
		return corev1.PodFailed
	}

	// For successfully terminated pods
	if pod.Status.Phase == corev1.PodSucceeded {
		return corev1.PodSucceeded
	}

	ready := false
	initialized := false
	for _, c := range pod.Status.Conditions {
		if c.Type == corev1.PodReady {
			ready = c.Status == corev1.ConditionTrue
		}
		if c.Type == corev1.PodInitialized {
			initialized = c.Status == corev1.ConditionTrue
		}
	}

	if initialized && ready && pod.Status.Phase == corev1.PodRunning {
		return corev1.PodRunning
	}

	// If the pod would otherwise be pending but has warning then label it as
	// failed and show and error to the user.
	if len(warnings) > 0 {
		return corev1.PodFailed
	}

	// pending
	return corev1.PodPending
}
