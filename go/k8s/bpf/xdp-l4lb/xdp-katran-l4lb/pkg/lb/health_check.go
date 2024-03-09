package lb

import (
	"context"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-katran-l4lb/pkg/rpc"
)

func (lb *OpenLb) AddHealthcheckerDst(ctx context.Context, healthcheck *rpc.Healthcheck) (*rpc.Bool, error) {
	//TODO implement me
	panic("implement me")
}

func (lb *OpenLb) DelHealthcheckerDst(ctx context.Context, somark *rpc.Somark) (*rpc.Bool, error) {
	//TODO implement me
	panic("implement me")
}

func (lb *OpenLb) GetHealthcheckersDst(ctx context.Context, empty *rpc.Empty) (*rpc.HcMap, error) {
	//TODO implement me
	panic("implement me")
}
