package clusterpool

import (
	"net"

	"github.com/cilium/ipam/cidrset"
)

// @see https://github.com/cilium/cilium/blob/v1.11.1/pkg/ipam/allocator/clusterpool/clusterpool.go

type CIDRAllocator struct {
	cidrSet *cidrset.CidrSet
}

func NewCIDRAllocator(cidr *net.IPNet, maskSize int) (*CIDRAllocator, error) {
	cidrSet, err := cidrset.NewCIDRSet(cidr, maskSize)
	if err != nil {
		return nil, err
	}

	return &CIDRAllocator{
		cidrSet: cidrSet,
	}, nil
}

func (cidr *CIDRAllocator) Allocate() (*net.IPNet, error) {
	return cidr.cidrSet.AllocateNext()
}

func (cidr *CIDRAllocator) Release(ipnet *net.IPNet) error {
	return cidr.cidrSet.Release(ipnet)
}

func (cidr *CIDRAllocator) InRange(ipnet *net.IPNet) bool {
	return cidr.cidrSet.InRange(ipnet)
}

func ForEachIP(ipnet net.IPNet, iterator func(ip string) error) error {
	next := make(net.IP, len(ipnet.IP))
	copy(next, ipnet.IP)
	for ipnet.Contains(next) {
		if err := iterator(next.String()); err != nil {
			return err
		}
		IncrIP(next)
	}
	return nil
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
