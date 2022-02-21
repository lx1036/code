package ipam

import (
	"net"

	"github.com/cilium/cilium/pkg/ip"

	"github.com/cilium/ipam/service/ipallocator"
)

type hostScopeAllocator struct {
	allocCIDR []*net.IPNet
	allocator *ipallocator.Range
}

func newHostScopeAllocator(cidrs []*net.IPNet) Allocator {
	cidrRange, err := ipallocator.NewCIDRRange(cidrs)
	if err != nil {
		panic(err)
	}
	a := &hostScopeAllocator{
		allocCIDR: cidrs,
		allocator: cidrRange,
	}

	return a
}
