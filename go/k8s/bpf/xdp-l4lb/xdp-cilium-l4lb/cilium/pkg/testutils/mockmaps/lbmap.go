package mockmaps

import (
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/loadbalancer"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/maps/lbmap"
)

type LBMockMap struct {
	BackendByID      map[uint16]*loadbalancer.Backend
	ServiceByID      map[uint16]*loadbalancer.SVC
	AffinityMatch    lbmap.BackendIDByServiceIDSet
	SourceRanges     lbmap.SourceRangeSetByServiceID
	DummyMaglevTable map[uint16]int // svcID => backends count
}
