package utils

import (
	corev1 "k8s.io/api/core/v1"
)

func IsHeadlessService(svc *corev1.Service) bool {
	return svc.Spec.Type == corev1.ServiceTypeClusterIP &&
		(svc.Spec.ClusterIP == corev1.ClusterIPNone || len(svc.Spec.ClusterIP) == 0)
}
