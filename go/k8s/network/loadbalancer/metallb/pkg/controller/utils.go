package controller

import (
	"k8s-lx1036/k8s/network/loadbalancer/metallb/pkg/allocator"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// Ports turns a service definition into a set of allocator ports.
func Ports(svc *corev1.Service) []allocator.Port {
	var ret []allocator.Port
	for _, port := range svc.Spec.Ports {
		ret = append(ret, allocator.Port{
			Proto: string(port.Protocol),
			Port:  int(port.Port),
		})
	}
	return ret
}

// SharingKey extracts the sharing key for a service.
func SharingKey(svc *corev1.Service) string {
	return svc.Annotations["metallb.universe.tf/allow-shared-ip"]
}

// BackendKey extracts the backend key for a service.
func BackendKey(svc *corev1.Service) string {
	if svc.Spec.ExternalTrafficPolicy == corev1.ServiceExternalTrafficPolicyTypeLocal {
		return labels.Set(svc.Spec.Selector).String()
	}
	// Cluster traffic policy can share services regardless of backends.
	return ""
}
