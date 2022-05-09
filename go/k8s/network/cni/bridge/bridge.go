package main

import (
	"encoding/json"
	"fmt"
	"net"
	"runtime"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/containernetworking/plugins/pkg/ip"
	"github.com/containernetworking/plugins/pkg/ipam"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/containernetworking/plugins/pkg/utils"
	bv "github.com/containernetworking/plugins/pkg/utils/buildversion"
	"github.com/vishvananda/netlink"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/util/iptables"
	"k8s.io/utils/exec"
)

func init() {
	// this ensures that main runs only on main thread (thread group leader).
	// since namespace ops (unshare, setns) are done for a single thread, we
	// must ensure that the goroutine does not jump from OS thread to thread
	runtime.LockOSThread()
}

func main() {
	skel.PluginMain(cmdAdd, cmdCheck, cmdDel, version.All, bv.BuildString("bridge"))
}

func cmdAdd(args *skel.CmdArgs) error {
	conf, err := loadConf(args.StdinData)
	if err != nil {
		return err
	}

	netNS, err := ns.GetNS(args.Netns)
	if err != nil {
		return fmt.Errorf("failed to open netns %q: %v", args.Netns, err)
	}
	defer netNS.Close()

	if conf.HairpinMode && conf.PromiscMode {
		return fmt.Errorf("cannot set hairpin mode and promiscuous mode at the same time")
	}

	// (1) setup bridge link
	vlanFiltering := conf.Vlan != 0
	br := &netlink.Bridge{
		LinkAttrs: netlink.LinkAttrs{
			Name: conf.BrName,
			MTU:  conf.MTU,
			// Let kernel use default txqueuelen; leaving it unset
			// means 0, and a zero-length TX queue messes up FIFO
			// traffic shapers which use TX queue length as the
			// default packet limit
			TxQLen: -1,
		},
		VlanFiltering: &vlanFiltering,
	}
	if err = netlink.LinkAdd(br); err != nil {
		return err
	}
	if conf.PromiscMode {
		if err = netlink.SetPromiscOn(br); err != nil {
			return err
		}
	}
	brLink, err := netlink.LinkByName(conf.BrName)
	if err != nil {
		return err
	}
	if err = netlink.LinkSetUp(brLink); err != nil {
		return err
	}
	brInterface := &current.Interface{ // 宿主机侧不需要写 Sandbox
		Name: brLink.Attrs().Name,
		Mac:  brLink.Attrs().HardwareAddr.String(),
	}

	// (2) setup veth pair, set host veth up to bridge
	var (
		hostInterface      *current.Interface
		containerInterface *current.Interface
	)
	err = netNS.Do(func(hostNS ns.NetNS) error {
		peerName, _ := ip.RandomVethName()
		err = netlink.LinkAdd(&netlink.Veth{
			LinkAttrs: netlink.LinkAttrs{
				Name:  args.IfName,
				MTU:   conf.MTU,
				Flags: net.FlagUp,
			},
			PeerName:      peerName,
			PeerNamespace: netlink.NsFd(int(hostNS.Fd())), // peer veth 在宿主机一侧
		})
		containerLink, err := netlink.LinkByName(args.IfName)
		if err != nil {
			return err
		}
		if err = netlink.LinkSetUp(containerLink); err != nil {
			return err
		}
		containerInterface = &current.Interface{
			Name:    containerLink.Attrs().Name,
			Mac:     containerLink.Attrs().HardwareAddr.String(),
			Sandbox: netNS.Path(),
		}

		err = hostNS.Do(func(_ ns.NetNS) error {
			hostLink, err := netlink.LinkByName(peerName)
			if err != nil {
				return err
			}
			if err = netlink.LinkSetUp(hostLink); err != nil {
				return err
			}
			hostInterface = &current.Interface{ // 宿主机侧不需要写 Sandbox
				Name: hostLink.Attrs().Name,
				Mac:  hostLink.Attrs().HardwareAddr.String(),
			}
			// INFO: set hostLink master bridge, 宿主机侧的 veth link 放置在 bridge
			_ = netlink.LinkSetMaster(hostLink, brLink)
			_ = netlink.LinkSetHairpin(hostLink, conf.HairpinMode)
			if conf.Vlan != 0 {
				_ = netlink.BridgeVlanAdd(hostLink, uint16(conf.Vlan), true, true, false, true)
			}

			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	// Assume L2 interface only
	result := &current.Result{
		CNIVersion: current.ImplementedSpecVersion,
		Interfaces: []*current.Interface{
			brInterface,
			hostInterface,
			containerInterface,
		},
		DNS: conf.DNS,
	}

	isLayer3 := conf.IPAM.Type != ""
	if isLayer3 {
		ipamResult, err := ipam.ExecAdd(conf.IPAM.Type, args.StdinData)
		// Invoke ipam del if err to avoid ip leak
		defer func() {
			if err != nil {
				ipam.ExecDel(conf.IPAM.Type, args.StdinData)
			}
		}()
		if err != nil {
			return err
		}
		// Convert whatever the IPAM result was into the current Result type
		r, err := current.NewResultFromResult(ipamResult)
		if err != nil {
			return err
		}
		result.IPs = r.IPs
		result.Routes = r.Routes
		if len(result.IPs) == 0 {
			return fmt.Errorf("IPAM plugin returned missing IP config")
		}

		// configure container IP and routes
		err = netNS.Do(func(_ ns.NetNS) error {
			containerLink, err := netlink.LinkByName(args.IfName)
			if err != nil {
				return err
			}

			var gw net.IP
			for _, ipConfig := range result.IPs {
				if result.Interfaces[*ipConfig.Interface].Name != args.IfName {
					continue
				}

				gw = ipConfig.Gateway
				err = netlink.AddrAdd(containerLink, &netlink.Addr{ // sudo ip netns exec net1 ip addr add 10.0.1.10/24 dev ipvlan1
					IPNet: &ipConfig.Address, // 192.168.1.0/24
				})
				if err != nil {
					klog.Errorf(fmt.Sprintf("config addr %s for interface %s err:%v", ipConfig.Address.String(),
						containerLink.Attrs().Name, err))
					continue
				}

				if gw != nil {
					defaultRoute := &net.IPNet{
						IP:   net.IPv4zero,
						Mask: net.CIDRMask(0, 32),
					}
					if err = netlink.RouteAdd(&netlink.Route{ // default route `sudo ip netns exec net1 ip route add default dev ipvlan1`
						Dst:       defaultRoute,
						Gw:        gw,
						LinkIndex: containerLink.Attrs().Index,
						Scope:     netlink.SCOPE_UNIVERSE,
						Flags:     int(netlink.FLAG_ONLINK),
					}); err != nil {
						return err
					}
				}
			}

			for _, r := range result.Routes { // extra routes
				var gateway net.IP
				if r.GW == nil {
					gateway = gw
				} else {
					gateway = r.GW
				}
				if err = netlink.RouteAddEcmp(&netlink.Route{
					Dst:       &r.Dst,
					Gw:        gateway,
					LinkIndex: containerLink.Attrs().Index,
				}); err != nil {
					klog.Errorf(fmt.Sprintf("failed to add route '%s via %s dev %s' err: %v", r.Dst.String(), gw.String(), args.IfName, err))
					continue
				}
			}

			return nil
		})
		if err != nil {
			return err
		}

		// TODO: assign an IP address to the bridge, makes the assigned IP the default route

		// INFO: IP地址伪装，在该 result.IPs IP 段内做 SNAT，source ip 地址选择本机机器出口网卡 eth0 的地址。
		//  访问 pod 时候，目标地址是 pod ip，但是 src ip 选择 IP 伪装的机器地址
		if conf.IPMasq {
			chain := utils.FormatChainName(conf.Name, args.ContainerID)
			comment := utils.FormatComment(conf.Name, args.ContainerID)
			multicastNet := "224.0.0.0/4"
			iptablesCmdHandler := iptables.New(exec.New(), iptables.ProtocolIPv4)
			_, err = iptablesCmdHandler.EnsureChain(iptables.TableNAT, iptables.Chain(chain))
			if err != nil {
				return err
			}
			for _, ipConfig := range result.IPs {
				markArgs := []string{"-s", ipConfig.Address.IP.String(), "-j", chain, "-m", "comment", "--comment", comment}
				_, _ = iptablesCmdHandler.EnsureRule(iptables.Append, iptables.TableNAT, iptables.ChainPostrouting, markArgs...)

				markArgsSlice := [][]string{
					{"-d", ipConfig.Address.String(), "-j", "ACCEPT", "-m", "comment", "--comment", comment},
					{"!", "-d", multicastNet, "-j", "MASQUERADE", "-m", "comment", "--comment", comment}, // IP 伪装
				}
				for _, value := range markArgsSlice {
					_, _ = iptablesCmdHandler.EnsureRule(iptables.Append, iptables.TableNAT, iptables.Chain(chain), value...)
				}
			}
		}
	}

	return types.PrintResult(result, conf.CNIVersion)
}

func cmdDel(args *skel.CmdArgs) error {
	conf, err := loadConf(args.StdinData)
	if err != nil {
		return err
	}

	isLayer3 := conf.IPAM.Type != ""
	if isLayer3 {
		err = ipam.ExecDel(conf.IPAM.Type, args.StdinData)
		if err != nil {
			return err
		}
	}

	if args.Netns == "" {
		return nil
	}

	var ipnets []*net.IPNet
	netNS, err := ns.GetNS(args.Netns)
	if err != nil {
		return fmt.Errorf("failed to open netns %q: %v", args.Netns, err)
	}
	defer netNS.Close()
	if err = netNS.Do(func(_ ns.NetNS) error {
		containerLink, err := netlink.LinkByName(args.IfName)
		if err != nil {
			return err
		}

		addrs, err := netlink.AddrList(containerLink, netlink.FAMILY_V4)
		if err != nil {
			return err
		}
		for _, addr := range addrs {
			ipnets = append(ipnets, addr.IPNet)
		}

		return netlink.LinkDel(containerLink)
	}); err != nil {
		return err
	}

	// 清除容器网卡 IP 相关的 iptable rules
	if isLayer3 && conf.IPMasq {
		chain := utils.FormatChainName(conf.Name, args.ContainerID)
		comment := utils.FormatComment(conf.Name, args.ContainerID)
		iptablesCmdHandler := iptables.New(exec.New(), iptables.ProtocolIPv4)
		_ = iptablesCmdHandler.DeleteChain(iptables.TableNAT, iptables.Chain(chain))
		for _, ipnet := range ipnets {
			markArgs := []string{"-s", ipnet.String(), "-j", chain, "-m", "comment", "--comment", comment}
			_ = iptablesCmdHandler.DeleteRule(iptables.TableNAT, iptables.ChainPostrouting, markArgs...)
		}
	}

	return nil
}

func cmdCheck(args *skel.CmdArgs) error {
	return nil
}

type NetConf struct {
	types.NetConf

	BrName string `json:"bridge"`

	Vlan int `json:"vlan"`
	MTU  int `json:"mtu"`

	HairpinMode bool `json:"hairpinMode"`
	PromiscMode bool `json:"promiscMode"`

	IPMasq bool `json:"ipMasq"`
}

func loadConf(stdin []byte) (*NetConf, error) {
	conf := NetConf{}

	if err := json.Unmarshal(stdin, &conf); err != nil {
		return nil, fmt.Errorf("failed to parse network configuration: %v", err)
	}

	if conf.Vlan < 0 || conf.Vlan > 4094 {
		return nil, fmt.Errorf("invalid VLAN ID %d (must be between 0 and 4094)", conf.Vlan)
	}

	return &conf, nil
}
