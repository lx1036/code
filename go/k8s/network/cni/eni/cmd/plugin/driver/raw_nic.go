package driver

import "github.com/containernetworking/plugins/pkg/ns"

// RawNicDriver put nic in net ns
type RawNicDriver struct {
	name string
	ipv4 bool
	ipv6 bool
}

func NewRawNICDriver(ipv4, ipv6 bool) *RawNicDriver {
	return &RawNicDriver{
		name: "rawNIC",
		ipv4: ipv4,
		ipv6: ipv6,
	}
}

func (driver *RawNicDriver) Setup(cfg *SetupConfig, netNS ns.NetNS) error {
	panic("implement me")
}

func (driver *RawNicDriver) Teardown(cfg *TeardownCfg, netNS ns.NetNS) error {
	panic("implement me")
}

func (driver *RawNicDriver) Check(cfg *CheckConfig) error {
	panic("implement me")
}
