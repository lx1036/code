package node

import (
	"fmt"
	"net"

	apiv1 "k8s-lx1036/k8s/network/cilium/cilium-ipam/pkg/apis/ipam.k9s.io/v1"
	"k8s-lx1036/k8s/network/cilium/cilium-ipam/pkg/ipam/allocator/clusterpool"

	"github.com/projectcalico/calico/libcalico-go/lib/selector"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

const (
	ipv4PodCidr = "io.cilium.network.ipv4-pod-cidr"
)

var (
	NoIPPoolErr = fmt.Errorf("has no ippool in cluster for current node")
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

func (l *LoadBalancer) Allocate(node *corev1.Node, key string) (*corev1.Node, error) {
	var ipnet *net.IPNet
	var err error
	n := node.DeepCopy()
	if n.Annotations != nil && len(n.Annotations[ipv4PodCidr]) != 0 {
		_, ipnet, err = net.ParseCIDR(n.Annotations[ipv4PodCidr])
		if err != nil {
			ipnet = nil
		}
	}

	alloc, err := l.getAllocatorByNode(node)
	if err != nil {
		return nil, err
	}

	if ipnet != nil {
		ipnet, err = alloc.allocator.Allocate(ipnet)
		if err != nil {
			return nil, err
		}
	}

	if ipnet == nil {
		ipnet, err = alloc.allocator.AllocateNext()
		if err != nil {
			return nil, err
		}
	}

	l.owner[key] = ipnet

	if n.Annotations == nil {
		n.Annotations = make(map[string]string)
	}
	n.Annotations[ipv4PodCidr] = ipnet.String()

	return n, nil
}

func (l *LoadBalancer) Release(node *corev1.Node) error {
	pool, err := l.getAllocatorByNode(node)
	if err != nil {
		return err
	}
	key, _ := cache.MetaNamespaceKeyFunc(node)

	ipnet, ok := l.owner[key]
	if !ok {
		return fmt.Errorf("no allocated cidr for node:%s", key)
	}

	return pool.allocator.Release(ipnet)
}

func (l *LoadBalancer) getAllocatorByNode(node *corev1.Node) (*Pool, error) {
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

	if p == nil {
		return nil, NoIPPoolErr
	}

	return p, nil
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

func (l *LoadBalancer) DeleteAllocator(name string) {
	delete(l.allocators, name)
}

func isIPPoolByNode(node *corev1.Node, ippool apiv1.IPPool) (bool, error) {
	sel, err := selector.Parse(ippool.Spec.NodeSelector)
	if err != nil {
		klog.Errorf(fmt.Sprintf("parse ippool NodeSelector err:%v", err))
		return false, err
	}

	return sel.Evaluate(node.Labels), nil
}
