package linux

import (
	"github.com/vishvananda/netlink"
	"net"

	"k8s-lx1036/k8s/network/cilium/cilium/pkg/config/defaults"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/datapath"
)

type addressFamilyIPv4 struct{}

func (a *addressFamilyIPv4) Router() net.IP {
	//TODO implement me
	panic("implement me")
}

func (a *addressFamilyIPv4) PrimaryExternal() net.IP {
	//TODO implement me
	panic("implement me")
}

func (a *addressFamilyIPv4) AllocationCIDR() *interface{} {
	//TODO implement me
	panic("implement me")
}

// INFO: netlink 获取本机器的所有真实网卡地址，如 eth0/eth1 等地址，和 cilium_host 虚拟网卡地址
func (a *addressFamilyIPv4) LocalAddresses() ([]net.IP, error) {
	return listLocalAddresses(netlink.FAMILY_V4)
}

func (a *addressFamilyIPv4) LoadBalancerNodeAddresses() []net.IP {
	//TODO implement me
	panic("implement me")
}

type linuxNodeAddressing struct {
	ipv4 addressFamilyIPv4
}

func NewNodeAddressing() datapath.NodeAddressing {
	return &linuxNodeAddressing{}
}

func (n *linuxNodeAddressing) IPv4() datapath.NodeAddressingFamily {
	return &n.ipv4
}

// INFO: netlink 获取本机器的所有真实网卡地址，如 eth0/eth1 等地址，和 cilium_host 虚拟网卡地址
func listLocalAddresses(family int) ([]net.IP, error) {
	var addresses []net.IP
	addrs, err := netlink.AddrList(nil, family)
	if err != nil {
		return nil, err
	}
	for _, addr := range addrs {
		if addr.Scope == int(netlink.SCOPE_LINK) {
			continue
		}
		if addr.IP.IsLoopback() {
			continue
		}

		addresses = append(addresses, addr.IP)
	}

	hostDevice, err := netlink.LinkByName(defaults.HostDevice) // cilium_host
	if err == nil && hostDevice != nil {
		addrs, err = netlink.AddrList(hostDevice, family)
		if err != nil {
			return nil, err
		}

		for _, addr := range addrs {
			if addr.Scope == int(netlink.SCOPE_LINK) {
				addresses = append(addresses, addr.IP)
			}
		}
	}

	return addresses, nil
}
