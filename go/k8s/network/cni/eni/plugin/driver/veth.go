package driver

import (
	"fmt"
	"k8s-lx1036/k8s/network/cni/eni/pkg/sysctl"
	"k8s-lx1036/k8s/network/cni/eni/types"

	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/vishvananda/netlink"
)

type VETHDriver struct {
	name string
	ipv4 bool
	ipv6 bool
}

func NewVETHDriver(ipv4, ipv6 bool) *VETHDriver {
	return &VETHDriver{
		name: "veth",
		ipv4: ipv4,
		ipv6: ipv6,
	}
}

// INFO:
func (driver *VETHDriver) Setup(cfg *SetupConfig, netNS ns.NetNS) error {
	prevHostLink, err := netlink.LinkByName(cfg.HostVETHName)
	if err == nil {
		if err = LinkDel(prevHostLink); err != nil {
			return fmt.Errorf("error del pre host link, %w", err)
		}
	}

	// disabled accept_ra if local forwarding is enabled
	// @see https://www.kernel.org/doc/Documentation/networking/ip-sysctl.txt
	if cfg.ENIIndex != 0 {
		parentLink, err := netlink.LinkByIndex(cfg.ENIIndex)
		if err != nil {
			return fmt.Errorf("error get eni by index %d, %w", cfg.ENIIndex, err)
		}

		if driver.ipv6 {
			_ = sysctl.WriteProcSys(fmt.Sprintf("/proc/sys/net/ipv6/conf/%s/accept_ra", parentLink.Attrs().Name), "0")
			_ = sysctl.WriteProcSys(fmt.Sprintf("/proc/sys/net/ipv6/conf/%s/forwarding", parentLink.Attrs().Name), "1")
		}
	}

	var hostVETH, containerVETH netlink.Link

	err = netNS.Do(func(_ ns.NetNS) error {
		if driver.ipv6 {
			err := EnableIPv6()
			if err != nil {
				return err
			}
		}

		// 1. create veth pair
		hostVETH, containerVETH, err = setupVETHPair(cfg.ContainerIfName, cfg.HostVETHName, cfg.MTU, hostNetNS)
		if err != nil {
			return fmt.Errorf("error create veth pair, %w", err)
		}

		// 2. add address for container interface
		containerLink, err := netlink.LinkByName(containerVETH.Attrs().Name)
		if err != nil {
			return fmt.Errorf("error find link %s in container, %w", containerVETH.Attrs().Name, err)
		}
		IPNetToMaxMask(cfg.ContainerIPNet)
		err = SetupLink(containerLink, cfg)
		if err != nil {
			return err
		}

		if driver.ipv6 {
			_ = sysctl.WriteProcSys(fmt.Sprintf("/proc/sys/net/ipv6/conf/%s/accept_ra", cfg.ContainerIfName), "0")
		}

		defaultGW := &types.IPSet{}
		if cfg.ContainerIPNet.IPv4 != nil {
			defaultGW.IPv4 = linkIPNet.IP
		}
		if cfg.ContainerIPNet.IPv6 != nil {
			defaultGW.IPv6 = linkIPNetv6.IP
		}
		_, err = EnsureDefaultRoute(contLink, defaultGW, unix.RT_TABLE_MAIN)
		if err != nil {
			return err
		}

		// 3. add route and neigh for container
		err = AddNeigh(contLink, hostVETH.Attrs().HardwareAddr, defaultGW)
		if err != nil {
			return err
		}

		if len(cfg.ExtraRoutes) != 0 {
			if driver.ipv4 {
				_, err = EnsureRoute(&netlink.Route{
					LinkIndex: containerLink.Attrs().Index,
					Scope:     netlink.SCOPE_LINK,
					Dst:       linkIPNet,
				})
				if err != nil {
					return fmt.Errorf("error add route for container veth, %w", err)
				}
			}
			if driver.ipv6 {
				_, err = EnsureRoute(&netlink.Route{
					LinkIndex: containerLink.Attrs().Index,
					Scope:     netlink.SCOPE_LINK,
					Dst:       linkIPNetv6,
				})
				if err != nil {
					return fmt.Errorf("error add route for container veth, %w", err)
				}
			}

			for _, extraRoute := range cfg.ExtraRoutes {
				err = RouteAdd(&netlink.Route{
					LinkIndex: containerLink.Attrs().Index,
					Scope:     netlink.SCOPE_UNIVERSE,
					Flags:     int(netlink.FLAG_ONLINK),
					Dst:       &extraRoute.Dst,
					Gw:        extraRoute.GW,
				})
				if err != nil {
					return fmt.Errorf("error add extra route for container veth, %w", err)
				}
			}
		}

		if cfg.Egress > 0 {
			return driver.setupTC(containerLink, cfg.Egress)
		}

		return nil
	})

	if err != nil {
		return err
	}

	// config in host netns
	hostVETHLink, err := netlink.LinkByName(hostVETH.Attrs().Name)
	if err != nil {
		return fmt.Errorf("error found link %s, %w", hostVETH.Attrs().Name, err)
	}

	_, err = EnsureLinkUp(hostVETHLink)
	if err != nil {
		return fmt.Errorf("error set link %s to up, %w", hostVETHLink.Attrs().Name, err)
	}

	// 1. config to container routes
	_, err = EnsureHostToContainerRoute(hostVETHLink, cfg.ContainerIPNet)
	if err != nil {
		return err
	}

	if len(cfg.ExtraRoutes) != 0 {
		if driver.ipv4 {
			err = AddrReplace(hostVETHLink, &netlink.Addr{
				IPNet: linkIPNet,
			})
			if err != nil {
				return fmt.Errorf("error add extra addr %s, %w", linkIPNet.String(), err)
			}
		}
		if driver.ipv6 {
			err = AddrReplace(hostVETHLink, &netlink.Addr{
				IPNet: linkIPNetv6,
			})
			if err != nil {
				return fmt.Errorf("error add extra addr %s, %w", linkIPNetv6.String(), err)
			}
		}
	}

	// 2. config from container routes
	if cfg.ENIIndex != 0 {
		parentLink, err := netlink.LinkByIndex(cfg.ENIIndex)
		if err != nil {
			return fmt.Errorf("error get eni by index %d, %w", cfg.ENIIndex, err)
		}

		tableID := getRouteTableID(parentLink.Attrs().Index)

		// ensure eni config
		err = driver.ensureENIConfig(parentLink, cfg.TrunkENI, cfg.MTU, tableID, cfg.GatewayIP, cfg.HostIPSet)
		if err != nil {
			return fmt.Errorf("error setup eni config, %w", err)
		}

		_, err = EnsureIPRule(hostVETHLink, cfg.ContainerIPNet, tableID)
		if err != nil {
			return err
		}
	}

	if cfg.Ingress > 0 {
		return driver.setupTC(hostVETHLink, cfg.Ingress)
	}

	return nil
}
