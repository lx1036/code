package lb

import (
	"context"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/pkg/rpc"
)

/**
 * position of elements inside control vector
 */
const (
	MacAddrPos  = 0
	Ipv4TunPos  = 1
	Ipv6TunPos  = 2
	MainIntfPos = 3
	HcIntfPos   = 4
)

// value for ctl mac, could contains e.g. mac address of default router
// or other flags
type CtlValue struct {
	value   uint64
	ifindex uint32
	mac     [6]uint8
}

func (lb *OpenLb) ChangeMac(ctx context.Context, mac *rpc.Mac) (*rpc.Bool, error) {
	//TODO implement me
	panic("implement me")
}

func (lb *OpenLb) GetMac(ctx context.Context, empty *rpc.Empty) (*rpc.Mac, error) {
	//TODO implement me
	panic("implement me")
}
