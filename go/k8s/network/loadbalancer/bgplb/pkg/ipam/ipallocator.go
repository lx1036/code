package ipam

import (
	"net"

	"github.com/cilium/ipam/service/ipallocator"
)

// AllocationResult is the result of an allocation
type AllocationResult struct {
	// IP is the allocated IP
	IP net.IP

	// CIDRs is a list of all CIDRs to which the IP has direct access to.
	// This is primarily useful if the IP has been allocated out of a VPC
	// subnet range and the VPC provides routing to a set of CIDRs in which
	// the IP is routable.
	CIDRs []string

	// PrimaryMAC is the MAC address of the primary interface. This is useful
	// when the IP is a secondary address of an interface which is
	// represented on the node as a Linux device and all routing of the IP
	// must occur through that master interface.
	PrimaryMAC string

	// GatewayIP is the IP of the gateway which must be used for this IP.
	// If the allocated IP is derived from a VPC, then the gateway
	// represented the gateway of the VPC or VPC subnet.
	GatewayIP string

	// ExpirationUUID is the UUID of the expiration timer. This field is
	// only set if AllocateNextWithExpiration is used.
	ExpirationUUID string

	// InterfaceNumber is a field for generically identifying an interface.
	// This is only useful in ENI mode.
	InterfaceNumber string
}

// Allocator is the interface for an IP allocator implementation
type Allocator interface {
	// Allocate allocates a specific IP or fails
	Allocate(ip net.IP, owner string) (*AllocationResult, error)

	// Release releases a previously allocated IP or fails
	Release(ip net.IP) error

	// AllocateNext allocates the next available IP or fails if no more IPs
	// are available
	AllocateNext(owner string) (*AllocationResult, error)

	GetCidr() *net.IPNet

	Free() int

	IsFull() bool

	Used() int

	FirstIP() net.IP

	LastIP() net.IP
}

// @see https://github.com/cilium/cilium/blob/v1.12.0-rc0/pkg/ipam/hostscope.go

type hostScopeAllocator struct {
	allocCIDR *net.IPNet
	allocator *ipallocator.Range
}

// Allocate allocate specified ip
func (alloc *hostScopeAllocator) Allocate(ip net.IP, owner string) (*AllocationResult, error) {
	if err := alloc.allocator.Allocate(ip); err != nil {
		return nil, err
	}

	return &AllocationResult{IP: ip}, nil
}

func (alloc *hostScopeAllocator) Release(ip net.IP) error {
	return alloc.allocator.Release(ip)
}

func (alloc *hostScopeAllocator) AllocateNext(owner string) (*AllocationResult, error) {
	ip, err := alloc.allocator.AllocateNext()
	if err != nil {
		return nil, err
	}

	return &AllocationResult{IP: ip}, nil
}

func (alloc *hostScopeAllocator) GetCidr() *net.IPNet {
	return alloc.allocCIDR
}

func (alloc *hostScopeAllocator) Free() int {
	return alloc.allocator.Free()
}

func (alloc *hostScopeAllocator) IsFull() bool {
	return alloc.Free() == 0
}

func (alloc *hostScopeAllocator) Used() int {
	return alloc.allocator.Used()
}

// FirstIP 第一个可用 IP
func (alloc *hostScopeAllocator) FirstIP() net.IP {
	ip := alloc.allocCIDR.IP
	dst := make(net.IP, len(ip))
	copy(dst, ip)

	IncrIP(dst)

	return dst
}

// LastIP 最后一个可用 IP
func (alloc *hostScopeAllocator) LastIP() net.IP {
	bcst := alloc.Broadcast()
	DecrIP(bcst)
	return bcst
}

// Broadcast 最后一个 IP
func (alloc *hostScopeAllocator) Broadcast() net.IP {
	ip := alloc.allocCIDR.IP
	dst := make(net.IP, len(ip))
	copy(dst, ip)

	mask := alloc.allocCIDR.Mask
	for i := 0; i < len(mask); i++ {
		ipIdx := len(dst) - i - 1
		dst[ipIdx] = ip[ipIdx] | ^mask[len(mask)-i-1]
	}

	return dst
}

func NewHostScopeAllocator(n *net.IPNet) Allocator {
	cidrRange, err := ipallocator.NewCIDRRange(n)
	if err != nil {
		panic(err)
	}
	a := &hostScopeAllocator{
		allocCIDR: n,
		allocator: cidrRange,
	}

	return a
}

// IncrIP IP地址自增
func IncrIP(ip net.IP) {
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] > 0 {
			break
		}
	}
}

// DecrIP IP地址自减
func DecrIP(ip net.IP) {
	length := len(ip)
	for i := length - 1; i >= 0; i-- {
		ip[length-1]--
		if ip[length-1] < 0xFF {
			break
		}
		for j := 1; j < length; j++ {
			ip[length-j-1]--
			if ip[length-j-1] < 0xFF {
				return
			}
		}
	}
}
