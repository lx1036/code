package main

import (
	"encoding/json"
	"fmt"
	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/plugins/pkg/ipam"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/vishvananda/netlink"
	"k8s.io/klog/v2"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/version"
	bv "github.com/containernetworking/plugins/pkg/utils/buildversion"
	"net"
	"runtime"
)

/* INFO: @see https://www.cni.dev/plugins/current/meta/sbr/
   https://github.com/containernetworking/plugins/blob/main/plugins/meta/sbr/main.go
   根据 ip source 来设置 ip rule 路由策略，然后 move default table routes to tableID table routes，这样就可以根据 ip source 来分流包
*/

const (
	firstTableID = 100
)

func init() {
	// this ensures that main runs only on main thread (thread group leader).
	// since namespace ops (unshare, setns) are done for a single thread, we
	// must ensure that the goroutine does not jump from OS thread to thread
	runtime.LockOSThread()
}

func main() {
	skel.PluginMain(cmdAdd, cmdCheck, cmdDel, version.All, bv.BuildString("src-routing"))
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
		result, err = current.NewResultFromResult(ipamResult)
		if err != nil {
			return err
		}

		if len(result.IPs) == 0 {
			return fmt.Errorf("IPAM plugin returned missing IP config")
		}
	}
	var ipConfigs []*current.IPConfig
	for _, ipConfig := range result.IPs {
		idx := *ipConfig.Interface
		if idx < 0 || idx >= len(result.Interfaces) || result.Interfaces[idx].Name != args.IfName {
			continue
		}
		ipConfigs = append(ipConfigs, ipConfig)
	}

	err = netNS.Do(func(_ ns.NetNS) error {
		rules, err := netlink.RuleList(netlink.FAMILY_ALL)
		if err != nil {
			return fmt.Errorf("Failed to list all rules: %v", err)
		}
		routes, err := netlink.RouteList(nil, netlink.FAMILY_ALL)
		if err != nil {
			return fmt.Errorf("Failed to list all routes: %v", err)
		}

		containerLink, err := netlink.LinkByName(args.IfName)
		if err != nil {
			return err
		}

		// Get all routes for the container interface in the default routing table
		routes, err = netlink.RouteList(containerLink, netlink.FAMILY_ALL)
		if err != nil {
			return err
		}

		// INFO: ipConfigs 有多个，且每一个在自己的 tableID 里，这里的逻辑更严谨
		tableID := getNextTableID(rules, routes, firstTableID)

		// setup source based rules and routes
		for _, ipConfig := range ipConfigs {
			// INFO: (1) 添加路由策略 `ip rule add from 204.153.191.1/32 table 200` 来自于 204.153.191.1/32 的包查询 table 200 里的路由
			rule := netlink.NewRule()
			rule.Table = tableID
			src := &net.IPNet{
				IP:   ipConfig.Address.IP,
				Mask: net.CIDRMask(32, 32),
			}
			rule.Src = src
			if err = netlink.RuleAdd(rule); err != nil { // `ip rule add from 204.153.191.1/32 table 200 prio 2048`
				klog.Errorf(fmt.Sprintf("add source %s ip rule err:%v", src.String(), err))
			}

			// Add a default route, since this may have been removed by previous plugin.
			if ipConfig.Gateway != nil {
				defaultRoute := &net.IPNet{
					IP:   net.IPv4zero,
					Mask: net.CIDRMask(0, 32),
				}
				if err = netlink.RouteAdd(&netlink.Route{ // default route `sudo ip netns exec net1 ip route add default dev ipvlan1 table 200`
					Dst:       defaultRoute,
					Gw:        ipConfig.Gateway,
					Table:     tableID,
					LinkIndex: containerLink.Attrs().Index,
				}); err != nil {
					return err
				}
			}

			// INFO: (2)move default table routes to tableID table routes
			for _, route := range routes { // p route add 192.168.1.0/24 dev ipvlan1 src 204.153.191.1 table 200
				if (route.Src == nil && route.Gw == nil) || ipConfig.Address.Contains(route.Src) || ipConfig.Address.Contains(route.Gw) {
					// (r.Src == nil && r.Gw == nil) is inferred as a generic route
					klog.Infof(fmt.Sprintf("copying route %s from table %d to %d", route.String(), route.Table, tableID))
					route.Table = tableID
					route.Flags = 0 // Reset the route flags since if it is dynamically created, adding it to the new table will fail with "invalid argument"
					// We use route replace in case the route already exists, which
					// is possible for the default gateway we added above.
					if err = netlink.RouteReplace(&route); err != nil {
						return fmt.Errorf("failed to add route: %v", err)
					}
				}
			}

			// Use a different table for each ipCfg
			tableID++
			tableID = getNextTableID(rules, routes, tableID)
		}

		for _, route := range routes { // delete routes in default table
			if err = netlink.RouteDel(&route); err != nil {
				klog.Errorf(fmt.Sprintf("delete route %s err:%v", route.String(), err))
				continue
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return types.PrintResult(result, conf.CNIVersion)
}

func cmdDel(args *skel.CmdArgs) error {
	_, err := loadConf(args.StdinData)
	if err != nil {
		return err
	}

	netNS, err := ns.GetNS(args.Netns)
	if err != nil {
		return fmt.Errorf("failed to open netns %q: %v", args.Netns, err)
	}
	defer netNS.Close()

	// delete ip rules for the deleted interface
	return netNS.Do(func(_ ns.NetNS) error {
		rules, err := netlink.RuleList(netlink.FAMILY_ALL)
		if err != nil {
			return fmt.Errorf("Failed to list all rules: %v", err)
		}
		containerLink, err := netlink.LinkByName(args.IfName)
		if err != nil {
			return fmt.Errorf("failed to get link %s: %v", args.IfName, err)
		}
		addrs, err := netlink.AddrList(containerLink, netlink.FAMILY_V4)
		if err != nil {
			return fmt.Errorf("failed to list all addrs: %v", err)
		}

		for _, rule := range rules {
			if rule.Src == nil {
				continue
			}

			for _, addr := range addrs {
				if rule.Src.IP.Equal(addr.IP) {
					_ = netlink.RuleDel(&rule)
					break
				}
			}
		}

		return nil
	})
}

func cmdCheck(args *skel.CmdArgs) error {
	return nil
}

type NetConf struct {
	types.NetConf
}

func loadConf(stdin []byte) (*NetConf, error) {
	conf := NetConf{}

	if err := json.Unmarshal(stdin, &conf); err != nil {
		return nil, fmt.Errorf("failed to parse network configuration: %v", err)
	}

	return &conf, nil
}

// INFO: pick next free table id. 从 rules 和 routes 中找出已有的 tableID，找出 next free tableID
func getNextTableID(rules []netlink.Rule, routes []netlink.Route, candidateID int) int {
	table := candidateID
	for {
		found := false
		for _, rule := range rules {
			if rule.Table == table {
				found = true
				break
			}
		}
		for _, route := range routes {
			if route.Table == table {
				found = true
				break
			}
		}
		if found {
			table++
		} else {
			break
		}
	}

	return table
}
