package lb

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-katran-l4lb/pkg/rpc"
)

const (
	LruCntrOffset = 0
)

type OpenLbStats struct {
	// number of failed syscalls
	addrValidationFailed uint64

	// times provided ipaddress was invalid
	bpfFailedCalls uint64
}

type lbStats struct {
	v1 uint64
	v2 uint64
}

func (lb *OpenLb) GetLruStats(ctx context.Context, empty *rpc.Empty) (*rpc.Stats, error) {

	lb.getLbStats(lb.config.maxReals + LruCntrOffset)

}

func (lb *OpenLb) GetLruMissStats(ctx context.Context, empty *rpc.Empty) (*rpc.Stats, error) {
	//TODO implement me
	panic("implement me")
}

func (lb *OpenLb) GetLruFallbackStats(ctx context.Context, empty *rpc.Empty) (*rpc.Stats, error) {
	//TODO implement me
	panic("implement me")
}

func (lb *OpenLb) GetIcmpTooBigStats(ctx context.Context, empty *rpc.Empty) (*rpc.Stats, error) {
	//TODO implement me
	panic("implement me")
}

func (lb *OpenLb) getLbStats(position uint32) {
	if lb.config.disableForwarding {
		msg := "Ignoring addVip call on non-forwarding instance"
		logrus.Warn(msg)
		return &rpc.Bool{Success: false}, fmt.Errorf(msg)
	}

	var sumStats lbStats

	if !lb.config.testing {
		lb.statsMap.Lookup(&position)
	}

	return sumStats
}
