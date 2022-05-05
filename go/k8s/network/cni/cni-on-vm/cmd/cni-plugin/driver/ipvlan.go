package driver

import (
	"fmt"
	"github.com/containernetworking/plugins/pkg/ns"
	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/utils/types"
	"k8s.io/klog/v2"
	"net"

	"github.com/vishvananda/netlink"
)

/*
```
sudo ip netns add net1
sudo ip link add ipv1 link eth0 type ipvlan mode l3
sudo ip link set ipv1 netns net1
sudo ip netns exec net1 ip link set ipv1 up
sudo ip netns exec net1 ip addr add 100.0.1.10/24 dev ipv1
sudo ip netns exec net1 ip route add default dev ipv1

sudo ip netns exec net1 ip route
default dev ipv1 scope link
100.0.1.0/24 dev ipv1 proto kernel scope link src 100.0.1.10

sudo ip netns exec net1 ip addr
1: lo: <LOOPBACK> mtu 65536 qdisc noop state DOWN group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
52154: ipv1@if2: <BROADCAST,MULTICAST,NOARP,UP,LOWER_UP> mtu 1500 qdisc noqueue state UNKNOWN group default qlen 1000
    link/ether 0c:c4:7a:56:f7:b2 brd ff:ff:ff:ff:ff:ff
    inet 100.0.1.10/24 scope global ipv1
       valid_lft forever preferred_lft forever
```
*/

var (
	_, defaultRoute, _ = net.ParseCIDR("0.0.0.0/0")
)

// IPvlanDriver INFO: IPVlan Linux docs：https://www.kernel.org/doc/Documentation/networking/ipvlan.txt
type IPvlanDriver struct {
	name string
	ipv4 bool
	ipv6 bool
}

func NewIPVlanDriver() *IPvlanDriver {
	return &IPvlanDriver{
		name: "IPVLan",
		ipv4: true,
	}
}

func (ipvlan *IPvlanDriver) Setup(cfg *types.SetupConfig, netNS ns.NetNS) error {
	parentLink, err := netlink.LinkByIndex(cfg.ENIIndex) // 弹性网卡 in host network namespace
	if err != nil {
		return fmt.Errorf("error get eni by index %d, %w", cfg.ENIIndex, err)
	}
	if cfg.MTU > 0 {
		if parentLink.Attrs().MTU != cfg.MTU {
			_ = netlink.LinkSetMTU(parentLink, cfg.MTU)
		}
	}

	// (1) 创建容器的 ipvlan eth0 interface
	link, err := netlink.LinkByName(cfg.HostVethIfName) // 为了兼容
	if err == nil {
		_ = netlink.LinkDel(link)
	}
	err = netlink.LinkAdd(&netlink.IPVlan{ // 在 container network namespace 内创建一个 ipvlan 网卡，且 name=HostVethIfName
		LinkAttrs: netlink.LinkAttrs{
			MTU:         cfg.MTU,
			Name:        cfg.HostVethIfName,
			ParentIndex: parentLink.Attrs().Index,
			Namespace:   netlink.NsFd(int(netNS.Fd())), // container network namespace
		},
		Mode: netlink.IPVLAN_MODE_L2,
	})
	if err != nil {
		klog.Errorf(fmt.Sprintf("add ipvlan interface name %s err:%v", cfg.HostVethIfName, err))
		return err
	}
	err = netNS.Do(func(netNS ns.NetNS) error { // rename to default "eth0"
		l, err := netlink.LinkByName(cfg.HostVethIfName)
		if err != nil {
			return err
		}
		err = netlink.LinkSetName(l, cfg.ContainerIfName)
		if err != nil {
			return err
		}
		return netlink.LinkSetUp(l) // `sudo ip netns exec net1 ip link set ipv1 up`
	})
	if err != nil {
		klog.Errorf(fmt.Sprintf("rename ipvlan interface name in container network namespace from %s to %s"))
		return err
	}

	// (2) add ip and add default route, arp neigh
	err = netNS.Do(func(netNS ns.NetNS) error {
		l, err := netlink.LinkByName(cfg.ContainerIfName)
		if err != nil {
			return err
		}
		err = netlink.AddrAdd(l, &netlink.Addr{ // `sudo ip netns exec net1 ip addr add 10.0.1.10/24 dev ipv1`
			IPNet: cfg.ContainerIPNet.IPv4,
		})
		if err != nil {
			return err
		}

		routes := []*netlink.Route{
			{
				LinkIndex: l.Attrs().Index,
				Scope:     netlink.SCOPE_UNIVERSE,
				Dst:       defaultRoute,
				Gw:        cfg.GatewayIP.IPv4,
				Flags:     int(netlink.FLAG_ONLINK),
			},
			{
				LinkIndex: l.Attrs().Index,
				Scope:     netlink.SCOPE_LINK,
				Dst:       cfg.HostIPSet.IPv4,
			},
		}
		for _, route := range routes {
			_ = netlink.RouteAdd(route)
		}

		return netlink.NeighAdd(&netlink.Neigh{ // arp host IP to container eth0 mac
			LinkIndex:    l.Attrs().Index,
			IP:           cfg.HostIPSet.IPv4.IP,
			HardwareAddr: l.Attrs().HardwareAddr,
			State:        netlink.NUD_PERMANENT,
		})
	})

}

func (ipvlan *IPvlanDriver) AddRoute() {

}

func (ipvlan *IPvlanDriver) Teardown(cfg *TeardownCfg, netNS ns.NetNS) error {
	panic("implement me")
}

func (ipvlan *IPvlanDriver) Check(cfg *CheckConfig) error {
	panic("implement me")
}
