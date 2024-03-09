package lb

import (
	"context"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-katran-l4lb/pkg/rpc"
)

func (lb *OpenLb) ModifyQuicRealsMapping(ctx context.Context, reals *rpc.ModifiedQuicReals) (*rpc.Bool, error) {
	//TODO implement me
	panic("implement me")
}

func (lb *OpenLb) GetQuicRealsMapping(ctx context.Context, empty *rpc.Empty) (*rpc.QuicReals, error) {
	//TODO implement me
	panic("implement me")
}
