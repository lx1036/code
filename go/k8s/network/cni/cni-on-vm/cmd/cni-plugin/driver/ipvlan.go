package driver

import (
	"encoding/binary"
	"fmt"
	"github.com/containernetworking/plugins/pkg/ip"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/vishvananda/netlink/nl"
	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/utils/types"
	"k8s.io/klog/v2"
	"net"

	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

// INFO: @see https://github.com/containernetworking/plugins/blob/main/plugins/main/ipvlan/ipvlan.go
//  https://www.cni.dev/plugins/current/main/ipvlan/

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
	ifname             = "net1"
	conf               = `{
	"cniVersion": "0.3.0",
	"name": "cni-plugin-sbr-test",
	"type": "sbr",
	"prevResult": {
		"cniVersion": "0.3.0",
		"interfaces": [
			{
				"name": "net1"
			}
		],
		"ips": [
			{
				"version": "4",
				"address": "192.168.1.209/24",
				"gateway": "192.168.1.1",
				"interface": 0
			}
		],
		"routes": []
	}
}`
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

	// (1) 创建容器侧的 ipvlan interface, configure ip and add default route, arp neigh
	// due to kernel bug we have to create with tmpname or it might
	// collide with the name on the host and error out
	// INFO: cfg.ContainerIfName 有可能是 eth0 等，会触发内核 bug
	tmpName, _ := ip.RandomVethName()
	err = netlink.LinkAdd(&netlink.IPVlan{
		LinkAttrs: netlink.LinkAttrs{
			Name:        tmpName,
			ParentIndex: parentLink.Attrs().Index, // parent link 是弹性网卡
			MTU:         cfg.MTU,
			Flags:       net.FlagUp,                    // `ip link set ipv1 up`
			Namespace:   netlink.NsFd(int(netNS.Fd())), // 在 container network namespace 内创建一个 ipvlan 网卡，且 name=HostVethIfName
		},
		Mode: netlink.IPVLAN_MODE_L2,
	})
	if err != nil {
		klog.Errorf(fmt.Sprintf("add ipvlan interface name %s err:%v", tmpName, err))
		return err
	}
	err = netNS.Do(func(netNS ns.NetNS) error {
		containerLink, err := netlink.LinkByName(tmpName)
		if err != nil {
			return err
		}
		err = netlink.LinkSetName(containerLink, cfg.ContainerIfName)
		if err != nil {
			return err
		}
		containerLink, err = netlink.LinkByName(cfg.ContainerIfName)
		if err != nil {
			return err
		}
		// ipvlan 网卡 和 parentLink 共享相同的 MAC 地址
		err = netlink.AddrAdd(containerLink, &netlink.Addr{ // `sudo ip netns exec net1 ip addr add 10.0.1.10/24 dev ipv1`
			IPNet: cfg.ContainerIPNet.IPv4, // 这里是 10.0.1.10/24，不是 10.0.1.10/32
		})
		if err != nil {
			return err
		}

		routes := []*netlink.Route{
			{
				Dst:       defaultRoute,
				Gw:        cfg.GatewayIP.IPv4,
				LinkIndex: containerLink.Attrs().Index,
				Scope:     netlink.SCOPE_UNIVERSE,
				Flags:     int(netlink.FLAG_ONLINK),
			},
			{
				Dst:       cfg.HostIPSet.IPv4,
				LinkIndex: containerLink.Attrs().Index,
				Scope:     netlink.SCOPE_LINK,
			},
		}
		for _, route := range routes {
			_ = netlink.RouteAdd(route)
		}

		return netlink.NeighAdd(&netlink.Neigh{ // arp host IP to container eth0 mac
			LinkIndex:    containerLink.Attrs().Index,
			IP:           cfg.HostIPSet.IPv4.IP,
			HardwareAddr: containerLink.Attrs().HardwareAddr,
			State:        netlink.NUD_PERMANENT,
		})
	})

	// (3) 配置 tc filter
	// 为了让Pod可以访问Service时可以经过宿主机的network namespace中的iptables/ipvs规则，
	// 所以另外增加了一个veth网卡打通Pod和宿主机的网络，并将集群的Service网段指向到这个veth网卡上。
	// (3.1) create and config slave ipvlan interface for service/hostStack cidr
	slaveLinkName := fmt.Sprintf("ipvlan_slave_%d", parentLink.Attrs().Index)
	slaveLink, err := netlink.LinkByName(slaveLinkName)
	if err != nil {
		if _, ok := err.(netlink.LinkNotFoundError); ok {
			err = netlink.LinkAdd(&netlink.IPVlan{ // 在 host network namespace 这一侧
				LinkAttrs: netlink.LinkAttrs{
					Name:        slaveLinkName,
					ParentIndex: parentLink.Attrs().Index,
					MTU:         cfg.MTU,
					Flags:       net.FlagUp, // `ip link set ipv1 up`
				},
				Mode: netlink.IPVLAN_MODE_L2,
			})
			if err != nil {
				klog.Errorf(fmt.Sprintf("add slave ipvlan interface name %s err:%v", slaveLinkName, err))
				return err
			}
		}
	} else {
		if slaveLink.Attrs().MTU != cfg.MTU {
			if err = netlink.LinkSetMTU(slaveLink, cfg.MTU); err != nil {
				return err
			}
		}
	}
	if slaveLink.Attrs().Flags&unix.IFF_NOARP != 0 { // arp off
		if err = netlink.LinkSetARPOff(slaveLink); err != nil {
			return err
		}
	}
	addrs, err := netlink.AddrList(slaveLink, netlink.FAMILY_V4)
	if err != nil {
		return err
	}
	found := false
	for _, addr := range addrs {
		if addr.IPNet.String() == cfg.HostIPSet.IPv4.String() && addr.Scope == int(netlink.SCOPE_HOST) {
			found = true
			break
		}
	}
	if !found {
		err = netlink.AddrAdd(slaveLink, &netlink.Addr{
			IPNet: cfg.HostIPSet.IPv4,
			Scope: int(netlink.SCOPE_HOST),
		})
		if err != nil {
			return err
		}
	}

	if err = netlink.RouteAdd(&netlink.Route{ // add route to container
		LinkIndex: slaveLink.Attrs().Index,
		Scope:     netlink.SCOPE_LINK,
		Dst: &net.IPNet{ // podIP/32
			IP:   cfg.ContainerIPNet.IPv4.IP,
			Mask: net.CIDRMask(32, 32),
		},
	}); err != nil {
		return err
	}
	// (3.2) 确保有一个 clsact qdisc 排队规则，然后在该规则里创建 filter
	// tc clsact qdisc 为 tc 提供了一个加载 ebpf 的入口。确保 parentLink 必须有 clsact qdisc
	// 参考 ifb ingress/egress 限流例子：go/k8s/network/cilium/bpf/tc-bpf/tc/bandwidth/ifb_creator.go
	// add filter on host device to mirror traffic to ifb device
	// `tc qdisc add dev eth0 handle ffff: clsact`
	// `tc filter add dev eth0 parent ffff: protocol ip u32 match u32 0 0 action mirred egress redirect dev ifb0`
	qdiscs, _ := netlink.QdiscList(parentLink) // `tc qdisc show dev eth0`
	found = false
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
	// INFO: 只有 serviceCIDR/hostStackCIDR 网段才会重定向到 slaveLink ipvlan interface
	cidrs := append(cfg.HostStackCIDRs, cfg.ServiceCIDR.IPv4)
	ptype := uint16(unix.PACKET_HOST)
	for _, cidr := range cidrs { // `tc filter add dev eth0 parent ffff0000: protocol ip u32 match ip src 192.168.0.0/16 action mirred ingress redirect dev slaveLink`
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
			RedirIndex: slaveLink.Attrs().Index, // redirect dev slaveLink
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

// Teardown 删除路由和eth0网卡
func (ipvlan *IPvlanDriver) Teardown(cfg *types.TeardownCfg, netNS ns.NetNS) error {
	err := netNS.Do(func(netNS ns.NetNS) error {
		link, err := netlink.LinkByName(cfg.ContainerIfName)
		if err != nil {
			return err
		}

		return netlink.LinkDel(link)
	})
	if err != nil {
		return err
	}

	// delete route to pod
	routes, err := netlink.RouteListFiltered(netlink.FAMILY_V4, &netlink.Route{
		Dst: cfg.ContainerIPNet.IPv4,
	}, netlink.RT_FILTER_DST)
	if err != nil {
		return err
	}
	for _, route := range routes {
		if err = netlink.RouteDel(&route); err != nil {
			klog.Errorf(fmt.Sprintf("delelete route %s err:%v", route.String(), err))
			continue
		}
	}

	return nil
}

func (ipvlan *IPvlanDriver) Check(cfg *CheckConfig) error {
	panic("implement me")
}
