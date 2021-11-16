package driver

import "github.com/containernetworking/plugins/pkg/ns"

// IPvlanDriver INFO: IPVlan Linux docs：https://www.kernel.org/doc/Documentation/networking/ipvlan.txt
type IPvlanDriver struct {
	name string
	ipv4 bool
	ipv6 bool
}

func (ipvlan *IPvlanDriver) Setup(cfg *SetupConfig, netNS ns.NetNS) error {
	panic("implement me")
}

func (ipvlan *IPvlanDriver) Teardown(cfg *TeardownCfg, netNS ns.NetNS) error {
	panic("implement me")
}

func (ipvlan *IPvlanDriver) Check(cfg *CheckConfig) error {
	panic("implement me")
}

func NewIPVlanDriver(ipv4, ipv6 bool) *IPvlanDriver {
	return &IPvlanDriver{
		name: "IPVLanL2",
		ipv4: ipv4,
		ipv6: ipv6,
	}
}
