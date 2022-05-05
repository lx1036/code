package driver

import (
	"encoding/binary"
	"fmt"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/vishvananda/netlink/nl"
	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/utils/types"
	"k8s.io/klog/v2"
	"net"

	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
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

const (
	QdiscClsact = "clsact"
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

	// (1) 创建容器侧的 ipvlan interface
	link, err := netlink.LinkByName(cfg.HostVethIfName) // 为了兼容
	if err == nil {
		_ = netlink.LinkDel(link)
	}
	err = netlink.LinkAdd(&netlink.IPVlan{ // 在 container network namespace 内创建一个 ipvlan 网卡，且 name=HostVethIfName
		LinkAttrs: netlink.LinkAttrs{
			Name:        cfg.HostVethIfName,
			ParentIndex: parentLink.Attrs().Index,
			MTU:         cfg.MTU,
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

	// (3) 创建宿主机侧的 ipvlan interface
	// 为了让Pod可以访问Service时可以经过宿主机的network namespace中的iptables/ipvs规则，
	// 所以另外增加了一个veth网卡打通Pod和宿主机的网络，并将集群的Service网段指向到这个veth网卡上。
	// (3.1) create and config slave ipvlan interface for service/hostStack cidr
	slaveLinkName := fmt.Sprintf("ipvlan_slave_%d", parentLink.Attrs().Index)
	slaveLink, err := netlink.LinkByName(slaveLinkName) // 为了兼容
	if err == nil {
		_ = netlink.LinkDel(slaveLink)
	}
	err = netlink.LinkAdd(&netlink.IPVlan{ // 在 host network namespace 这一侧
		LinkAttrs: netlink.LinkAttrs{
			Name:        slaveLinkName,
			ParentIndex: parentLink.Attrs().Index,
			MTU:         cfg.MTU,
		},
		Mode: netlink.IPVLAN_MODE_L2,
	})
	if err != nil {
		klog.Errorf(fmt.Sprintf("add slave ipvlan interface name %s err:%v", slaveLinkName, err))
		return err
	}
	slaveLink, _ = netlink.LinkByName(slaveLinkName)
	if err = netlink.LinkSetARPOff(slaveLink); err != nil {
		return err
	}
	err = netlink.AddrAdd(slaveLink, &netlink.Addr{
		IPNet: cfg.HostIPSet.IPv4,
		Scope: int(netlink.SCOPE_HOST),
	})
	if err != nil {
		return err
	}
	err = netlink.RouteAdd(&netlink.Route{ // add route to container
		LinkIndex: slaveLink.Attrs().Index,
		Scope:     netlink.SCOPE_LINK,
		Dst: &net.IPNet{ // podIP/32
			IP:   cfg.ContainerIPNet.IPv4.IP,
			Mask: net.CIDRMask(32, 32),
		},
	})
	if err != nil {
		return err
	}
	// (3.2) 确保有一个 clsact qdisc 排队规则，然后在该规则里创建 filter
	// tc clsact qdisc 为 tc 提供了一个加载 ebpf 的入口。确保 parentLink 必须有 clsact qdisc
	// 参考 ifb ingress/egress 限流例子：go/k8s/network/cilium/bpf/tc-bpf/tc/bandwidth/ifb_creator.go
	// add filter on host device to mirror traffic to ifb device
	// `tc qdisc add dev eth0 handle ffff: clsact`
	// `tc filter add dev eth0 parent ffff: protocol ip u32 match u32 0 0 action mirred egress redirect dev ifb0`
	qdiscs, _ := netlink.QdiscList(parentLink) // `tc qdisc show dev eth0`
	found := false
	for _, qdisc := range qdiscs {
		if qdisc.Type() == QdiscClsact {
			found = true
			break
		}
	}
	clsact := &netlink.GenericQdisc{ // `tc qdisc add dev eth0 handle ffff0000: clsact`
		QdiscAttrs: netlink.QdiscAttrs{
			LinkIndex: parentLink.Attrs().Index,
			Parent:    netlink.HANDLE_CLSACT,
			Handle:    netlink.HANDLE_CLSACT & 0xffff0000,
		},
		QdiscType: QdiscClsact,
	}
	if !found {
		_ = netlink.QdiscAdd(clsact)
	}
	//parent := uint32(netlink.HANDLE_CLSACT&0xffff0000 | netlink.HANDLE_MIN_EGRESS&0x0000ffff)
	cidrs := append(cfg.HostStackCIDRs, cfg.ServiceCIDR.IPv4)
	ptype := uint16(unix.PACKET_HOST)
	for _, cidr := range cidrs { // `tc filter add dev eth0 parent`
		err = netlink.FilterAdd(&netlink.U32{
			FilterAttrs: netlink.FilterAttrs{
				LinkIndex: parentLink.Attrs().Index,
				Parent:    clsact.QdiscAttrs.Handle,
				Priority:  40000,
				Protocol:  unix.ETH_P_IP,
			},
			Sel: &netlink.TcU32Sel{
				Nkeys: 1,
				Flags: nl.TC_U32_TERMINAL,
				Keys: []netlink.TcU32Key{
					{
						Mask: binary.BigEndian.Uint32(net.IP(cidr.Mask).To4()),
						Val:  binary.BigEndian.Uint32(cidr.IP.Mask(cidr.Mask).To4()),
						Off:  16,
					},
				},
			},
			Actions: []netlink.Action{
				&netlink.MirredAction{
					ActionAttrs: netlink.ActionAttrs{
						Action: netlink.TC_ACT_STOLEN,
					},
					MirredAction: netlink.TCA_INGRESS_REDIR, // ingress
					Ifindex:      slaveLink.Attrs().Index,   // mirred redirect to slaveLink
				},
				&netlink.TunnelKeyAction{
					ActionAttrs: netlink.ActionAttrs{
						Action: netlink.TC_ACT_PIPE,
					},
					Action: netlink.TCA_TUNNEL_KEY_UNSET,
				},
				&netlink.SkbEditAction{
					ActionAttrs: netlink.ActionAttrs{
						Action: netlink.TC_ACT_PIPE,
					},
					PType: &ptype,
				},
			},
		})
		if err != nil {
			klog.Errorf(fmt.Sprintf("tc filter add err:%v", err))
			continue
		}
	}

	return nil
}

func (ipvlan *IPvlanDriver) AddRoute() {

}

func (ipvlan *IPvlanDriver) Teardown(cfg *TeardownCfg, netNS ns.NetNS) error {
	panic("implement me")
}

func (ipvlan *IPvlanDriver) Check(cfg *CheckConfig) error {
	panic("implement me")
}
