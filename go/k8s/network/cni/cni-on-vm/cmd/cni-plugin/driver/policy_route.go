package driver

import (
	"github.com/containernetworking/plugins/pkg/ns"
	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/types"
)

type PolicyRoute struct{}

func NewPolicyRoute() *PolicyRoute {
	return &PolicyRoute{}
}

func (d *PolicyRoute) Setup(cfg *types.SetupConfig, netNS ns.NetNS) error {

}
