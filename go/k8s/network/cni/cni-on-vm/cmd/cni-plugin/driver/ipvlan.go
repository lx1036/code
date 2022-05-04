package driver

import (
	"fmt"
	"github.com/containernetworking/plugins/pkg/ns"
	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/types"

	"github.com/vishvananda/netlink"
)

// IPvlanDriver INFO: IPVlan Linux docs：https://www.kernel.org/doc/Documentation/networking/ipvlan.txt
type IPvlanDriver struct {
	name string
	ipv4 bool
	ipv6 bool
}

func NewIPVlanDriver() *IPvlanDriver {
	return &IPvlanDriver{
		name: "IPVLanL2",
		ipv4: true,
	}
}

func (ipvlan *IPvlanDriver) Setup(cfg *types.SetupConfig, netNS ns.NetNS) error {

	parentLink, err := netlink.LinkByIndex(cfg.ENIIndex) // 弹性网卡 in host network namespace
	if err != nil {
		return fmt.Errorf("error get eni by index %d, %w", cfg.ENIIndex, err)
	}

}

func (ipvlan *IPvlanDriver) AddRoute() {

}

func (ipvlan *IPvlanDriver) Teardown(cfg *TeardownCfg, netNS ns.NetNS) error {
	panic("implement me")
}

func (ipvlan *IPvlanDriver) Check(cfg *CheckConfig) error {
	panic("implement me")
}
