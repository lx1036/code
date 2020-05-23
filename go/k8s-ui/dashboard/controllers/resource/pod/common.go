package pod

import (
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/common"
	corev1 "k8s.io/api/core/v1"
)

func getPodStatus(pods *corev1.PodList, events []corev1.Event) common.ResourceStatus {

	for _, pod := range pods.Items {
		warnings := event.GetPodsEventWarnings(events, []corev1.Pod{pod})

	}
}


