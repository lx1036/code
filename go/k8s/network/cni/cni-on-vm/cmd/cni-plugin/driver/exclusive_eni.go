package driver

import (
	"fmt"
	"github.com/containernetworking/plugins/pkg/ip"
	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/utils/types"
	"net"

	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/vishvananda/netlink"
)

const defaultVethForENI = "veth1"

type ExclusiveENI struct{}

func NewExclusiveENIDriver() *ExclusiveENI {
	return &ExclusiveENI{}
}

// Setup INFO: 网卡单 IP 配置还是比较简单的
func (eni *ExclusiveENI) Setup(cfg *types.SetupConfig, netNS ns.NetNS) error {
	// move eni link to container net namespace, add addr to eni link
	eniLink, err := netlink.LinkByIndex(cfg.ENIIndex)
	if err != nil {
		return fmt.Errorf("error get eni by index %d, %w", cfg.ENIIndex, err)
	}
	if err = netlink.LinkSetNsFd(eniLink, int(netNS.Fd())); err != nil {
		return fmt.Errorf(fmt.Sprintf("fail to move eni link to container net namespace %v", err))
	}

	hostNetNS, err := ns.GetCurrentNS()
	if err != nil {
		return fmt.Errorf("err get host net ns, %w", err)
	}
	defer hostNetNS.Close()

	err = netNS.Do(func(netNS ns.NetNS) error {
		containerLink, err := netlink.LinkByName(eniLink.Attrs().Name)
		if err != nil {
			return fmt.Errorf("error find link %s, %w", eniLink.Attrs().Name, err)
		}

		err = netlink.AddrAdd(containerLink, &netlink.Addr{ // `sudo ip netns exec net1 ip addr add 10.0.1.10/24 dev ipv1`
			//IPNet: cfg.ContainerIPNet.IPv4, // 这里是 10.0.1.10/24，不是 10.0.1.10/32
			IPNet: &net.IPNet{ // 好像应该是 10.0.1.10/32，即 podIP/32
				IP:   cfg.ContainerIPNet.IPv4.IP,
				Mask: net.CIDRMask(32, 32),
			},
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
		}
		for _, route := range routes {
			_ = netlink.RouteAdd(route)
		}

		// for now we only create slave link for eth0
		// INFO: veth pair 常用的使用 169.254.1.1/32(mac 地址为 hostVeth mac) 来吧走 serviceCIDR 的包转给 hostVeth 网卡，这样
		//  包会走宿主机侧的 iptables/ipvs 规则。
		if cfg.ContainerIfName == "eth0" {
			tmpName, _ := ip.RandomVethName()
			err := netlink.LinkAdd(&netlink.Veth{
				LinkAttrs: netlink.LinkAttrs{
					Name:      tmpName,
					MTU:       cfg.MTU,
					Flags:     net.FlagUp,                        // `ip link set ipv1 up`
					Namespace: netlink.NsFd(int(hostNetNS.Fd())), // tmpName 在 host 侧
				},
				PeerName: defaultVethForENI, // 该网卡在容器侧
			})
			if err != nil {
				return err
			}
			var mac net.HardwareAddr
			err = hostNetNS.Do(func(netNS ns.NetNS) error {
				tmpLink, err := netlink.LinkByName(tmpName)
				if err != nil {
					return err
				}
				err = netlink.LinkSetUp(tmpLink)
				if err != nil {
					return err
				}
				err = netlink.LinkSetName(tmpLink, cfg.HostVethIfName)
				if err != nil {
					return err
				}

				mac = tmpLink.Attrs().HardwareAddr
				return nil
			})
			if err != nil {
				return err
			}

			veth1Link, err := netlink.LinkByName(defaultVethForENI)
			if err != nil {
				return err
			}
			// INFO: 容器侧 eth0 已经绑定了 podIP/32，现在又给 veth pair eht1 也绑定了  podIP/32
			err = netlink.AddrAdd(veth1Link, &netlink.Addr{ // `sudo ip netns exec net1 ip addr add 10.0.1.10/24 dev ipv1`
				//IPNet: cfg.ContainerIPNet.IPv4, // 这里是 10.0.1.10/24，不是 10.0.1.10/32
				IPNet: &net.IPNet{ // 好像应该是 10.0.1.10/32，即 podIP/32
					IP:   cfg.ContainerIPNet.IPv4.IP,
					Mask: net.CIDRMask(32, 32),
				},
			})
			if err != nil {
				return err
			}
			routes := []*netlink.Route{
				{
					Dst:       cfg.ServiceCIDR.IPv4, // INFO: service cidr
					Gw:        Gateway.IP,
					LinkIndex: veth1Link.Attrs().Index,
					Scope:     netlink.SCOPE_UNIVERSE,
					Flags:     int(netlink.FLAG_ONLINK),
				},
				{
					Dst:       Gateway,
					LinkIndex: veth1Link.Attrs().Index,
					Scope:     netlink.SCOPE_LINK,
				},
			}
			for _, route := range routes {
				_ = netlink.RouteAdd(route)
			}
			err = netlink.NeighAdd(&netlink.Neigh{ // arp host IP to container eth0 mac
				LinkIndex:    veth1Link.Attrs().Index,
				IP:           Gateway.IP,
				HardwareAddr: mac, // host 侧的网卡
				State:        netlink.NUD_PERMANENT,
			})
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	hostVethLink, err := netlink.LinkByName(cfg.HostVethIfName)
	if err != nil {
		return err
	}
	// add route to container
	routes := []*netlink.Route{
		{
			LinkIndex: hostVethLink.Attrs().Index,
			Scope:     netlink.SCOPE_LINK,
			Dst: &net.IPNet{ // 好像应该是 10.0.1.10/32，即 podIP/32
				IP:   cfg.ContainerIPNet.IPv4.IP,
				Mask: net.CIDRMask(32, 32),
			},
		},
	}
	for _, route := range routes {
		_ = netlink.RouteAdd(route)
	}

	return nil
}
