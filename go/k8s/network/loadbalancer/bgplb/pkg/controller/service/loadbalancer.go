package service

import (
	"errors"
	"fmt"
	"net"

	v1 "k8s-lx1036/k8s/network/loadbalancer/bgplb/pkg/apis/bgplb.k9s.io/v1"
	"k8s-lx1036/k8s/network/loadbalancer/bgplb/pkg/ipam"

	"github.com/cilium/ipam/service/ipallocator"
	corev1 "k8s.io/api/core/v1"
)

var (
	NoIPPoolErr = fmt.Errorf("has no ippool in cluster")
)

type LoadBalancer struct {
	allocators map[string]ipam.Allocator

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
	ippoolName := svc.Annotations["loadbalancer/ippool-name"]
	if len(ippoolName) == 0 {
		ippoolName = "default"
	}
	alloc, ok := l.allocators[ippoolName]
	if !ok {
		return nil, NoIPPoolErr
	}
	cidr := alloc.GetCidr()

	if lbIP != nil { // allocated loadbalancer ip
		allocResult, err := alloc.Allocate(lbIP, key)
		if err != nil { // obsolete loadbalancer ip and reallocate
			if errors.Is(err, &ipallocator.ErrNotInRange{}) {
				lbIP = nil
			} else {
				return nil, err
			}
		} else {
			lbIP = allocResult.IP
		}
	}

	if lbIP == nil {
		ip, err := alloc.AllocateNext(key)
		if err != nil {
			if errors.Is(err, ipallocator.ErrFull) {

			} else {
				return nil, err
			}
		} else {
			lbIP = ip.IP
		}
	}

	l.owner[key] = lbIP.String()
	svc.Status.LoadBalancer.Ingress = []corev1.LoadBalancerIngress{{IP: lbIP.String()}}
	return svc, nil
}

func (l *LoadBalancer) AddAllocator(name string, allocator ipam.Allocator) {
	l.allocators[name] = allocator
}

func (l *LoadBalancer) DeleteAllocator(name string) {
	delete(l.allocators, name)
}
