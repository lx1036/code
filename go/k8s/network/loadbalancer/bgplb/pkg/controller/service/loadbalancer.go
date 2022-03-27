package service

import (
	"fmt"
	v1 "k8s-lx1036/k8s/network/loadbalancer/bgplb/pkg/apis/bgplb.k9s.io/v1"
	"k8s-lx1036/k8s/network/loadbalancer/bgplb/pkg/ipam"
	"net"

	corev1 "k8s.io/api/core/v1"
)

var (
	NoIPPoolErr = fmt.Errorf("has no ippool in cluster")
)

type LoadBalancer struct {
	allocator map[string]ipam.Allocator

	// owner maps an IP to the owner
	owner map[string]string
}

func NewLoadBalancer(ippools []v1.IPPool) {

}

func (l *LoadBalancer) Allocate(service *corev1.Service, key string) (*corev1.Service, error) {
	var lbIP net.IP

	svc := service.DeepCopy()
	if len(svc.Status.LoadBalancer.Ingress) == 1 {
		lbIP = net.ParseIP(svc.Status.LoadBalancer.Ingress[0].IP)
	}

	// choose ippool
	var cidr *net.IPNet
	ippoolName := svc.Annotations["loadbalancer/ippool-name"]
	if len(ippoolName) == 0 {
		ippoolName = "default"
	}
	alloc, ok := l.allocator[ippoolName]
	if !ok {
		return nil, NoIPPoolErr
	}

	if lbIP != nil {
		if cidr.Contains(lbIP) {
			ip, err := alloc.Allocate(lbIP, key)
		} else {
			lbIP = nil
		}
	}

	if lbIP == nil {
		ip, err := alloc.AllocateNext(key)
		lbIP = ip.IP
	}

	l.owner[key] = lbIP.String()
	svc.Status.LoadBalancer.Ingress = []corev1.LoadBalancerIngress{{IP: lbIP.String()}}
	return svc, nil
}

func (l *LoadBalancer) AddAllocator(name string, allocator ipam.Allocator) {
	l.allocator[name] = allocator
}

func (l *LoadBalancer) DeleteAllocator(name string) {
	delete(l.allocator, name)
}
