package event

import (
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/common"
	corev1 "k8s.io/api/core/v1"
)

func GetPodsEventWarnings(events []corev1.Event, pods []corev1.Pod) []common.Event {

}

