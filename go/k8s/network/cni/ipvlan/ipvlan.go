package main

import (
	"encoding/json"
	"fmt"
	"github.com/containernetworking/cni/pkg/skel"
	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/containernetworking/plugins/pkg/ip"
	"github.com/containernetworking/plugins/pkg/ipam"
	"github.com/containernetworking/plugins/pkg/ns"
	bv "github.com/containernetworking/plugins/pkg/utils/buildversion"
	"github.com/vishvananda/netlink"
	"k8s.io/klog/v2"
	"net"
	"runtime"

	"github.com/containernetworking/cni/pkg/types"
)

/* INFO: @see https://www.cni.dev/plugins/current/main/ipvlan/
    https://github.com/containernetworking/plugins/blob/main/plugins/main/ipvlan/ipvlan.go
    以 eth0 为 parent link，创建一个 ipvlan link，然后从 ipam 获取 ip 并配置新建的 ipvlan 网卡
	{
	"name": "mynet",
	"type": "ipvlan",
	"master": "eth0",
	"ipam": {
		"type": "host-local",
		"subnet": "10.1.2.0/24"
		}
	}
*/

func init() {
	// this ensures that main runs only on main thread (thread group leader).
	// since namespace ops (unshare, setns) are done for a single thread, we
	// must ensure that the goroutine does not jump from OS thread to thread
	runtime.LockOSThread()
}

func main() {
	skel.PluginMain(cmdAdd, cmdCheck, cmdDel, version.All, bv.BuildString("ipvlan"))
}

func cmdAdd(args *skel.CmdArgs) error {
	conf, cniVersion, err := loadConf(args.StdinData, false)
	if err != nil {
		return err
	}

	netNS, err := ns.GetNS(args.Netns)
	if err != nil {
		return fmt.Errorf("failed to open netns %q: %v", args.Netns, err)
	}
	defer netNS.Close()

	mode, err := modeFromString(conf.Mode)
	if err != nil {
		return err
	}
	parentLink, err := netlink.LinkByName(conf.Master)
	if err != nil {
		return fmt.Errorf("failed to lookup master %q: %v", conf.Master, err)
	}
	// due to kernel bug we have to create with tmpname or it might
	// collide with the name on the host and error out
	tmpName, _ := ip.RandomVethName()
	if err = netlink.LinkAdd(&netlink.IPVlan{
		LinkAttrs: netlink.LinkAttrs{
			Name:        tmpName,
			ParentIndex: parentLink.Attrs().Index,
			MTU:         conf.MTU,
			Namespace:   netlink.NsFd(int(netNS.Fd())),
		},
		Mode: mode,
	}); err != nil {
		return fmt.Errorf("failed to create ipvlan: %v", err)
	}
	var ipvlanInterface *current.Interface
	if err = netNS.Do(func(_ ns.NetNS) error {
		containerLink, err := netlink.LinkByName(tmpName)
		if err != nil {
			return err
		}
		err = netlink.LinkSetName(containerLink, args.IfName)
		if err != nil {
			return err
		}
		containerLink, err = netlink.LinkByName(args.IfName)
		if err != nil {
			return err
		}
		err = netlink.LinkSetUp(containerLink)
		if err != nil {
			return err
		}

		ipvlanInterface = &current.Interface{
			Name:    containerLink.Attrs().Name,
			Mac:     containerLink.Attrs().HardwareAddr.String(),
			Sandbox: netNS.Path(),
		}
		return nil
	}); err != nil {
		return err
	}

	// (1)IP 从 PrevResult 中获取，或者 IP 从 IPAM 二进制中获取
	var result *current.Result
	haveResult := false
	if conf.IPAM.Type == "" && conf.PrevResult != nil {
		result, err = current.NewResultFromResult(conf.PrevResult)
		if err != nil {
			return err
		}
		if len(result.IPs) > 0 {
			haveResult = true
		}
	}
	if !haveResult {
		var err1 error
		ipamResult, err1 := ipam.ExecAdd(conf.IPAM.Type, args.StdinData)
		// Invoke ipam del if err to avoid ip leak
		defer func() {
			if err1 != nil {
				ipam.ExecDel(conf.IPAM.Type, args.StdinData)
			}
		}()
		if err1 != nil {
			return err1
		}

		// Convert whatever the IPAM result was into the current Result type
		result, err1 = current.NewResultFromResult(ipamResult)
		if err1 != nil {
			return err1
		}

		if len(result.IPs) == 0 {
			return fmt.Errorf("IPAM plugin returned missing IP config")
		}
	}

	result.Interfaces = []*current.Interface{ipvlanInterface}
	result.DNS = conf.DNS
	for _, ipc := range result.IPs {
		// All addresses belong to the ipvlan interface
		ipc.Interface = current.Int(0)
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

	return types.PrintResult(result, cniVersion)
}

func cmdDel(args *skel.CmdArgs) error {
	conf, _, err := loadConf(args.StdinData, false)
	if err != nil {
		return err
	}

	// On chained invocation, IPAM block can be empty
	if conf.IPAM.Type != "" {
		err = ipam.ExecDel(conf.IPAM.Type, args.StdinData)
		if err != nil {
			return err
		}
	}

	if args.Netns == "" {
		return nil
	}

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

		return netlink.LinkDel(containerLink)
	}); err != nil {
		return err
	}

	return nil
}

func cmdCheck(args *skel.CmdArgs) error {
	return nil
}

type NetConf struct {
	types.NetConf
	Master string `json:"master"`
	Mode   string `json:"mode"`
	MTU    int    `json:"mtu"`
}

func loadConf(bytes []byte, cmdCheck bool) (*NetConf, string, error) {
	conf := &NetConf{}
	if err := json.Unmarshal(bytes, conf); err != nil {
		return nil, "", fmt.Errorf("failed to load netconf: %v", err)
	}

	if cmdCheck {
		return conf, conf.CNIVersion, nil
	}

	return conf, conf.CNIVersion, nil
}

func modeFromString(s string) (netlink.IPVlanMode, error) {
	switch s {
	case "", "l2":
		return netlink.IPVLAN_MODE_L2, nil
	case "l3":
		return netlink.IPVLAN_MODE_L3, nil
	case "l3s":
		return netlink.IPVLAN_MODE_L3S, nil
	default:
		return 0, fmt.Errorf("unknown ipvlan mode: %q", s)
	}
}
