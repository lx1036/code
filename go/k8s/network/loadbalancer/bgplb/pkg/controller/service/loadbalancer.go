package service

import (
	"errors"
	"fmt"
	"k8s.io/client-go/tools/cache"
	"net"

	v1 "k8s-lx1036/k8s/network/loadbalancer/bgplb/pkg/apis/bgplb.k9s.io/v1"
	"k8s-lx1036/k8s/network/loadbalancer/bgplb/pkg/ipam"

	"github.com/cilium/ipam/service/ipallocator"
	corev1 "k8s.io/api/core/v1"
)

var (
	NoIPPoolErr = fmt.Errorf("has no ippool in cluster for current service")
)

const (
	svcIPPoolAnnotation = "loadbalancer/ippool-name"
	defaultIPPoolName   = "default"
)

type LoadBalancer struct {
	// map ippool name to allocator
	allocators map[string]ipam.Allocator

	// owner maps an IP to the owner
	owner map[string]string
}

func NewLoadBalancer(ippools []v1.IPPool) (*LoadBalancer, error) {
	balancer := &LoadBalancer{
		allocators: make(map[string]ipam.Allocator),
		owner:      make(map[string]string),
	}

	for _, ippool := range ippools {
		key, _ := cache.MetaNamespaceKeyFunc(ippool)
		if err := balancer.AddAllocator(key, ippool); err != nil {
			return nil, err
		}
	}

	return balancer, nil
}

func (l *LoadBalancer) Allocate(service *corev1.Service, key string) (*corev1.Service, error) {
	var lbIP net.IP

	svc := service.DeepCopy()
	if len(svc.Status.LoadBalancer.Ingress) == 1 {
		lbIP = net.ParseIP(svc.Status.LoadBalancer.Ingress[0].IP)
	}

	// choose ippool allocator
	alloc, err := l.getAllocatorByService(svc)
	if err != nil {
		return nil, err
	}

	if lbIP != nil { // allocated loadbalancer ip
		allocResult, err := alloc.Allocate(lbIP, key)
		if err != nil { // obsolete loadbalancer ip and reallocate
			if errors.Is(err, &ipallocator.ErrNotInRange{}) {
				lbIP = nil
			} else if errors.Is(err, ipallocator.ErrAllocated) {
				return service, nil
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

func (l *LoadBalancer) Release(service *corev1.Service) error {
	var lbIP net.IP

	svc := service.DeepCopy()
	alloc, err := l.getAllocatorByService(svc)
	if err != nil {
		return err
	}

	if len(svc.Status.LoadBalancer.Ingress) == 1 {
		lbIP = net.ParseIP(svc.Status.LoadBalancer.Ingress[0].IP)
	}

	if lbIP != nil {
		return alloc.Release(lbIP)
	}

	return nil
}

func (l *LoadBalancer) getIPPoolNameByService(service *corev1.Service) string {
	ippoolName := service.Annotations[svcIPPoolAnnotation]
	if len(ippoolName) == 0 {
		ippoolName = defaultIPPoolName
	}

	return ippoolName
}

func (l *LoadBalancer) getAllocatorByService(service *corev1.Service) (ipam.Allocator, error) {
	ippoolName := l.getIPPoolNameByService(service)
	alloc, ok := l.allocators[ippoolName]
	if !ok {
		return nil, NoIPPoolErr
	}

	return alloc, nil
}

func (l *LoadBalancer) GetAllocator(name string) ipam.Allocator {
	return l.allocators[name]
}

func (l *LoadBalancer) AddAllocator(name string, ippool v1.IPPool) error {
	_, cidr, err := net.ParseCIDR(ippool.Spec.Cidr)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("parse ippool:%s cidr:%s err:%v", name, ippool.Spec.Cidr, err))
	}
	allocator := ipam.NewHostScopeAllocator(cidr)

	l.allocators[name] = allocator
	return nil
}

func (l *LoadBalancer) DeleteAllocator(name string) {
	delete(l.allocators, name)
}

func (l *LoadBalancer) ListAllocators() map[string]ipam.Allocator {
	return l.allocators
}
