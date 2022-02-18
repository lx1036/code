package clusterpool

import (
	"context"
	"fmt"
	"net"

	ipPkg "github.com/cilium/cilium/pkg/ip"
	"github.com/cilium/cilium/pkg/ipam"
	"github.com/cilium/cilium/pkg/ipam/allocator"
	"github.com/cilium/cilium/pkg/ipam/allocator/podcidr"
	"github.com/cilium/ipam/cidrset"
)

// @see https://github.com/cilium/cilium/blob/v1.11.1/pkg/ipam/allocator/clusterpool/clusterpool.go

type ErrCIDRColision struct {
	cidr      string
	allocator podcidr.CIDRAllocator
}

func (e ErrCIDRColision) Error() string {
	return fmt.Sprintf("requested CIDR %s colides with %s", e.cidr, e.allocator)
}

func (e *ErrCIDRColision) Is(target error) bool {
	t, ok := target.(*ErrCIDRColision)
	if !ok {
		return false
	}
	return t.cidr == e.cidr
}

// AllocatorOperator is an implementation of IPAM allocator interface for Cilium
// IPAM.
type AllocatorOperator struct {
	v4CIDRSet []podcidr.CIDRAllocator
}

// Init sets up Cilium allocator based on given options
func (a *AllocatorOperator) Init(ctx context.Context, clusterPoolIPv4CIDR []string, nodeCIDRMaskSizeIPv4 int) error {
	v4Allocators, err := newCIDRSets(false, clusterPoolIPv4CIDR, nodeCIDRMaskSizeIPv4)
	if err != nil {
		return fmt.Errorf("unable to initialize IPv4 allocator %w", err)
	}
	a.v4CIDRSet = v4Allocators

	return nil
}

func (a *AllocatorOperator) Start(ctx context.Context, updater ipam.CiliumNodeGetterUpdater) (allocator.NodeEventHandler, error) {
	nodeManager := podcidr.NewNodesPodCIDRManager(a.v4CIDRSet, a.v6CIDRSet, updater, iMetrics)

	return nodeManager, nil
}

func NewCIDRSets(isV6 bool, strCIDRs []string, maskSize int) ([]podcidr.CIDRAllocator, error) {
	cidrAllocators := make([]podcidr.CIDRAllocator, 0, len(strCIDRs))
	for _, strCIDR := range strCIDRs {
		addr, cidr, err := net.ParseCIDR(strCIDR)
		if err != nil {
			return nil, err
		}
		// Check if CIDRs collide with each other.
		for _, cidrAllocator := range cidrAllocators {
			if cidrAllocator.InRange(cidr) {
				return nil, &ErrCIDRColision{
					cidr:      strCIDR,
					allocator: cidrAllocator,
				}
			}
		}
		cidrSet, err := newCIDRSet(isV6, addr, cidr, maskSize)
		if err != nil {
			return nil, err
		}
		cidrAllocators = append(cidrAllocators, cidrSet)
	}
	return cidrAllocators, nil
}

func newCIDRSet(isV6 bool, addr net.IP, cidr *net.IPNet, maskSize int) (podcidr.CIDRAllocator, error) {
	switch {
	case isV6 && ipPkg.IsIPv4(addr):
		return nil, fmt.Errorf("CIDR is not v6 family: %s", cidr)
	case !isV6 && !ipPkg.IsIPv4(addr):
		return nil, fmt.Errorf("CIDR is not v4 family: %s", cidr)
	}

	return cidrset.NewCIDRSet(cidr, maskSize)
}
