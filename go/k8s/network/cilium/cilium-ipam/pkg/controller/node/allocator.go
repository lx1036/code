package node

import (
	"fmt"
	"k8s-lx1036/k8s/network/cilium/cilium-ipam/pkg/ipam/allocator/clusterpool"
	"net"

	"github.com/projectcalico/calico/libcalico-go/lib/selector"
	apiv1 "k8s-lx1036/k8s/network/cilium/cilium-ipam/pkg/apis/ipam.k9s.io/v1"
	"k8s.io/client-go/tools/cache"

	corev1 "k8s.io/api/core/v1"
)

type Pool struct {
	allocator *clusterpool.CIDRAllocator
	ippool    apiv1.IPPool
}

type LoadBalancer struct {
	// map ippool name to allocator
	allocators map[string]*Pool

	// owner maps an node to the IPNet
	owner map[string]*net.IPNet
}

func NewLoadBalancer(ippools []apiv1.IPPool) (*LoadBalancer, error) {
	balancer := &LoadBalancer{
		allocators: make(map[string]*Pool),
		owner:      make(map[string]*net.IPNet),
	}

	for _, ippool := range ippools {
		key, _ := cache.MetaNamespaceKeyFunc(ippool)
		if err := balancer.AddAllocator(key, ippool); err != nil {
			return nil, err
		}
	}

	return balancer, nil
}

func (l *LoadBalancer) Allocate(node *corev1.Node, key string) (*net.IPNet, error) {
	pool := l.GetAllocatorByNode(node)
	ipnet, err := pool.allocator.Allocate()

	l.owner[key] = ipnet

	return ipnet, err
}

func (l *LoadBalancer) Release(node *corev1.Node) error {
	pool := l.GetAllocatorByNode(node)
	key, _ := cache.MetaNamespaceKeyFunc(node)

	ipnet, ok := l.owner[key]
	if !ok {
		return fmt.Errorf("no allocated cidr for node:%s", key)
	}

	return pool.allocator.Release(ipnet)
}

func (l *LoadBalancer) GetAllocatorByNode(node *corev1.Node) *Pool {
	var p *Pool
	for _, pool := range l.allocators {
		ok, err := isIPPoolByNode(node, pool.ippool)
		if err != nil {
			continue
		}

		if ok {
			p = pool
			break
		}
	}

	return p
}

func (l *LoadBalancer) AddAllocator(name string, ippool apiv1.IPPool) error {
	_, cidr, err := net.ParseCIDR(ippool.Spec.Cidr)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("parse ippool:%s cidr:%s err:%v", name, ippool.Spec.Cidr, err))
	}

	allocator, err := clusterpool.NewCIDRAllocator(cidr, ippool.Spec.BlockSize)
	if err != nil {
		return err
	}

	l.allocators[name] = &Pool{
		ippool:    ippool,
		allocator: allocator,
	}

	return nil
}

func isIPPoolByNode(node *corev1.Node, ippool apiv1.IPPool) (bool, error) {
	sel, err := selector.Parse(ippool.Spec.NodeSelector)
	if err != nil {
		return false, err
	}

	return sel.Evaluate(node.Labels), nil
}
