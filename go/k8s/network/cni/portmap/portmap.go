package main

import (
	"encoding/json"
	"fmt"
	"net"
	"runtime"
	"strconv"
	"strings"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/containernetworking/plugins/pkg/utils"
	bv "github.com/containernetworking/plugins/pkg/utils/buildversion"
	"github.com/containernetworking/plugins/pkg/utils/sysctl"
	iptablesCmd "github.com/coreos/go-iptables/iptables"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
	"k8s.io/kubernetes/pkg/util/iptables"
	"k8s.io/utils/exec"
)

const (
	// SNAT
	SetMarkChainName  = "CNI-HOSTPORT-SETMARK"
	MarkMasqChainName = "CNI-HOSTPORT-MASQ"

	// DNAT
	TopLevelDNATChainName = "CNI-HOSTPORT-DNAT"
)

// INFO: portmap plugin 根据主机侧的 port 来 forward traffic to --to-destination podIP:podPort
//  @see https://www.cni.dev/plugins/current/meta/portmap/

func init() {
	// this ensures that main runs only on main thread (thread group leader).
	// since namespace ops (unshare, setns) are done for a single thread, we
	// must ensure that the goroutine does not jump from OS thread to thread
	runtime.LockOSThread()
}

func main() {
	skel.PluginMain(cmdAdd, cmdCheck, cmdDel, version.All, bv.BuildString("src-routing"))
}

// portmap 必须是 chain plugin
// INFO:
//
//	DNAT:
//	 PREROUTING, OUTPUT: --dst-type local -j CNI-HOSTPORT-DNAT
//	 CNI-HOSTPORT-DNAT: -m multiport --destination-ports 8080,8081 -j CNI-DN-abcd123
//	 CNI-HOSTPORT-SETMARK: -j MARK --set-xmark 0x2000/0x2000
//	 CNI-DN-abcd123: -p tcp -s 172.16.30.0/24 --dport 8080 -j CNI-HOSTPORT-SETMARK
//	 CNI-DN-abcd123: -p tcp -s 127.0.0.1 --dport 8080 -j CNI-HOSTPORT-SETMARK
//	 CNI-DN-abcd123: -p tcp --dport 8080 -j DNAT --to-destination 172.16.30.2:8080
//	SNAT:
//	 POSTROUTING: -j CNI-HOSTPORT-MASQ
//	 CNI-HOSTPORT-MASQ: --mark 0x2000 -j MASQUERADE
func cmdAdd(args *skel.CmdArgs) error {
	netConf, err := loadConf(args.StdinData, args.IfName)
	if err != nil {
		return err
	}

	if netConf.PrevResult == nil {
		return fmt.Errorf("must be called as chained plugin")
	}
	if len(netConf.RuntimeConfig.PortMaps) == 0 {
		return types.PrintResult(netConf.PrevResult, netConf.CNIVersion)
	}

	netConf.ContainerID = args.ContainerID

	if netConf.ContainerIPv4.IP != nil {
		iptablesCmdHandler := iptables.New(exec.New(), iptables.ProtocolIPv4)

		// INFO: (1) SNAT Masquerade https://www.cni.dev/plugins/current/meta/portmap/#snat-masquerade
		//  Enable masquerading for traffic as necessary. The DNAT chain sets a mark bit for traffic that needs masq:
		//  Some packets also need to have the source address rewritten:
		//  [1] connections from localhost
		//  [2] Hairpin traffic back to the container
		if *netConf.SNAT {
			if netConf.ExternalSetMarkChain == nil { // 选择默认的 mark chain
				_, err = iptablesCmdHandler.EnsureChain(iptables.TableNAT, SetMarkChainName)
				if err != nil {
					return err
				}
				_, err = iptablesCmdHandler.EnsureChain(iptables.TableNAT, MarkMasqChainName)
				if err != nil {
					return err
				}

				markValue := 1 << uint(*netConf.MarkMasqBit)
				markDef := fmt.Sprintf("%#x/%#x", markValue, markValue) // 1<<13 的 16进制：0x2000/0x2000

				// iptables -t nat -A CNI-HOSTPORT-SETMARK -m comment --comment xxx -j MARK --set-xmark 0x2000/0x2000 (先打码)
				markArgs := []string{"-m", "comment", "--comment", "CNI port forward masquerade mark", "-j", "MARK", "--set-xmark", markDef}
				_, _ = iptablesCmdHandler.EnsureRule(iptables.Append, iptables.TableNAT, SetMarkChainName, markArgs...)

				// iptables -t nat -I POSTROUTING -m comment --comment xxx -j CNI-HOSTPORT-MASQ (从 POSTROUTING 跳到 CNI-HOSTPORT-MASQ)
				markArgs = []string{"-m", "comment", "--comment", "CNI port forward requiring masquerade", "-j", MarkMasqChainName}
				_, _ = iptablesCmdHandler.EnsureRule(iptables.Prepend, iptables.TableNAT, iptables.ChainPostrouting, markArgs...)
				// iptables -t nat -A CNI-HOSTPORT-MASQ -m mark --mark 0x2000/0x2000 -j MASQUERADE (对于打码的 packet 则 IP 伪装)
				markArgs = []string{"-m", "mark", "--mark", markDef, "-j", "MASQUERADE"}
				_, _ = iptablesCmdHandler.EnsureRule(iptables.Append, iptables.TableNAT, MarkMasqChainName, markArgs...)
			}

			// Set the route_localnet bit on the host interface, so that
			// 127/8 can cross a routing boundary. Only for ipv4
			hostIfName := getRoutableHostIF(netConf.ContainerIPv4.IP)
			if hostIfName != "" {
				if err := enableLocalnetRouting(hostIfName); err != nil {
					return fmt.Errorf("unable to enable route_localnet: %v", err)
				}
			}
		}

		// INFO: (2) DNAT https://www.cni.dev/plugins/current/meta/portmap/#dnat
		//  PREROUTING,OUTPUT -> CNI-HOSTPORT-DNAT -> CONTAINER-DNAT-xxxxxx -> 172.16.30.2:80
		_, err = iptablesCmdHandler.EnsureChain(iptables.TableNAT, TopLevelDNATChainName)
		if err != nil {
			return err
		}
		// iptables -t nat -A PREROUTING -m addrtype --dst-type LOCAL -j CNI-HOSTPORT-DNAT
		markArgs := []string{"-m", "addrtype", "--dst-type", "LOCAL", "-j", TopLevelDNATChainName}
		_, _ = iptablesCmdHandler.EnsureRule(iptables.Append, iptables.TableNAT, iptables.ChainPrerouting, markArgs...)
		// iptables -t nat -A OUTPUT -m addrtype --dst-type LOCAL -j CNI-HOSTPORT-DNAT
		_, _ = iptablesCmdHandler.EnsureRule(iptables.Append, iptables.TableNAT, iptables.ChainOutput, markArgs...)

		containerDnatChain := getContainerDnatChain(netConf.Name, netConf.ContainerID) // CONTAINER-DNAT-xxxxxx chain
		_, err = iptablesCmdHandler.EnsureChain(iptables.TableNAT, iptables.Chain(containerDnatChain))
		if err != nil {
			return err
		}

		// iptables -t nat -A CNI-HOSTPORT-DNAT ${ConditionsV4/6} -m comment --comment xxx -m multiport -p protocol --destination-ports 8080,8043 -j CONTAINER-DNAT-xxxxxx
		var protocolPorts map[string][]string
		for _, portMap := range netConf.RuntimeConfig.PortMaps {
			protocolPorts[portMap.Protocol] = append(protocolPorts[portMap.Protocol], strconv.Itoa(portMap.ContainerPort))
		}
		for protocol, ports := range protocolPorts {
			var containerDnatArgs []string
			if len(*netConf.ConditionsV4) != 0 {
				containerDnatArgs = append(containerDnatArgs, *netConf.ConditionsV4...)
			}
			containerDnatArgs = append(containerDnatArgs, "-m", "comment", "--comment", "forward to specified container dnat chain",
				"-m", "multiport", "-p", protocol, "--destination-ports", strings.Join(ports, ","), "-j", containerDnatChain)
			_, _ = iptablesCmdHandler.EnsureRule(iptables.Append, iptables.TableNAT, TopLevelDNATChainName, containerDnatArgs...)
		}

		// For every entry, generate 3 rules:
		// - mark hairpin for masq
		// - mark localhost for masq (for v4)
		// - do dnat
		// the ordering is important here; the mark rules must be first.
		var rules [][]string
		setMarkChainName := SetMarkChainName
		if netConf.ExternalSetMarkChain != nil {
			setMarkChainName = *netConf.ExternalSetMarkChain
		}
		for _, portMap := range netConf.RuntimeConfig.PortMaps {
			rulePrefix := []string{"-p", portMap.Protocol, "--dport", strconv.Itoa(portMap.HostPort)}
			if len(portMap.HostIP) != 0 {
				rulePrefix = append(rulePrefix, "-d", portMap.HostIP)
			}

			// Add mark-to-masquerade rules for hairpin and localhost
			if *netConf.SNAT {
				// iptables -t nat -A CONTAINER-DNAT-xxxxxx -p tcp -s 172.16.30.0/24 --dport 8080 -j CNI-HOSTPORT-SETMARK
				tmp := append(rulePrefix, "-s", netConf.ContainerIPv4.String(), "-j", setMarkChainName)
				rules = append(rules, tmp)
				// iptables -t nat -A CONTAINER-DNAT-xxxxxx -p tcp -s 127.0.0.1 --dport 8080 -j CNI-HOSTPORT-SETMARK
				tmp2 := append(rulePrefix, "-s", "127.0.0.1", "-j", setMarkChainName) // 只能是 ipv4 才可以
				rules = append(rules, tmp2)
			}
			// iptables -t nat -A CONTAINER-DNAT-xxxxxx -p tcp --dport 8080 -j DNAT --to-destination 172.16.30.2:80
			tmp3 := append(rulePrefix, "-j", "DNAT", "--to-destination", fmt.Sprintf("%s:%d", netConf.ContainerIPv4.IP, portMap.ContainerPort))
			rules = append(rules, tmp3)
		}
		for _, rule := range rules {
			_, _ = iptablesCmdHandler.EnsureRule(iptables.Append, iptables.TableNAT, iptables.Chain(containerDnatChain), rule...)
		}

		// Delete conntrack entries for UDP to avoid conntrack blackholing traffic
		// due to stale connections. We do that after the iptables rules are set, so
		// the new traffic uses them. Failures are informative only.
		for _, portMap := range netConf.RuntimeConfig.PortMaps {
			if strings.ToLower(portMap.Protocol) != "udp" {
				continue
			}

			filter := &netlink.ConntrackFilter{}
			_ = filter.AddPort(netlink.ConntrackOrigDstPort, uint16(portMap.HostPort))
			_ = filter.AddProtocol(utils.PROTOCOL_UDP)
			_, _ = netlink.ConntrackDeleteFilter(netlink.ConntrackTable, unix.AF_INET, filter)
		}
	}

	return types.PrintResult(netConf.PrevResult, netConf.CNIVersion)
}

func cmdDel(args *skel.CmdArgs) error {
	netConf, err := loadConf(args.StdinData, args.IfName)
	if err != nil {
		return err
	}

	if len(netConf.RuntimeConfig.PortMaps) == 0 {
		return nil
	}

	netConf.ContainerID = args.ContainerID

	// We don't need to parse out whether or not we're using v6 or snat, deletion is idempotent幂等的
	iptablesCmdHandler := iptables.New(exec.New(), iptables.ProtocolIPv4)
	containerDnatChain := getContainerDnatChain(netConf.Name, netConf.ContainerID)
	if err = iptablesCmdHandler.FlushChain(iptables.TableNAT, iptables.Chain(containerDnatChain)); err != nil {
		return err
	}

	iptablesHandler, _ := iptablesCmd.NewWithProtocol(iptablesCmd.ProtocolIPv4)
	rules, err := iptablesHandler.List(string(iptables.TableNAT), TopLevelDNATChainName)
	if err != nil {
		return err
	}
	for _, rule := range rules {
		if strings.Contains(rule, fmt.Sprintf("-j %s", containerDnatChain)) {
			// List results always include an "-A CHAINNAME -d 169.254.20.10/32 -p tcp -m tcp --dport 53 -j ACCEPT"
			matchRule := strings.Split(strings.TrimSpace(rule), " ")[2:]
			if err = iptablesHandler.Delete(string(iptables.TableNAT), TopLevelDNATChainName, matchRule...); err != nil {
				return err
			}
		}
	}

	return nil
}

func cmdCheck(args *skel.CmdArgs) error {
	return nil
}

// DefaultMarkBit The default mark bit to signal that masquerading is required
// Kubernetes uses 14 and 15, Calico uses 20-31.
const DefaultMarkBit = 13

type PortMapEntry struct {
	HostIP        string `json:"hostIP,omitempty"`
	HostPort      int    `json:"hostPort"`
	ContainerPort int    `json:"containerPort"`
	Protocol      string `json:"protocol"` // tcp or udp
}

type NetConf struct {
	types.NetConf

	SNAT         *bool     `json:"snat,omitempty"`
	ConditionsV4 *[]string `json:"conditionsV4"`
	MarkMasqBit  *int      `json:"markMasqBit"`
	//  If you already have a Masquerade mark chain (e.g. Kubernetes), specify it here. This will use that instead of creating a separate chain. When this is set, markMasqBit must be unspecified
	ExternalSetMarkChain *string `json:"externalSetMarkChain"`
	RuntimeConfig        struct {
		PortMaps []PortMapEntry `json:"portMappings,omitempty"`
	} `json:"runtimeConfig,omitempty"`

	// These are fields parsed out of the config or the environment;
	// included here for convenience
	ContainerID   string    `json:"-"`
	ContainerIPv4 net.IPNet `json:"-"`
}

func loadConf(stdin []byte, ifName string) (*NetConf, error) {
	conf := NetConf{}
	if err := json.Unmarshal(stdin, &conf); err != nil {
		return nil, fmt.Errorf("failed to parse network configuration: %v", err)
	}
	if conf.SNAT == nil {
		tvar := true
		conf.SNAT = &tvar
	}
	if conf.MarkMasqBit != nil && conf.ExternalSetMarkChain != nil {
		return nil, fmt.Errorf("cannot specify externalSetMarkChain and markMasqBit")
	}
	if conf.MarkMasqBit == nil {
		bvar := DefaultMarkBit // go constants are "special"
		conf.MarkMasqBit = &bvar
	}
	if *conf.MarkMasqBit < 0 || *conf.MarkMasqBit > 31 {
		return nil, fmt.Errorf("MasqMarkBit must be between 0 and 31")
	}
	for _, pm := range conf.RuntimeConfig.PortMaps {
		if pm.ContainerPort <= 0 {
			return nil, fmt.Errorf("Invalid container port number: %d", pm.ContainerPort)
		}
		if pm.HostPort <= 0 {
			return nil, fmt.Errorf("Invalid host port number: %d", pm.HostPort)
		}
	}

	if conf.PrevResult != nil {
		result, err := current.NewResultFromResult(conf.PrevResult)
		if err != nil {
			return nil, fmt.Errorf("could not convert result to current version: %v", err)
		}
		for _, ipConfig := range result.IPs {
			isIPV4 := ipConfig.Address.IP.To4() != nil
			if isIPV4 && conf.ContainerIPv4.IP != nil {
				continue
			}
			// Skip known non-sandbox interfaces
			if ipConfig.Interface != nil {
				index := *ipConfig.Interface
				if index < 0 || index >= len(result.Interfaces) || result.Interfaces[index].Name != ifName ||
					len(result.Interfaces[index].Sandbox) == 0 {
					continue
				}
			}

			conf.ContainerIPv4 = ipConfig.Address // INFO: 这样 conf.ContainerIPv4 就未必是最后一个 ips[l-1].address
		}
	}

	return &conf, nil
}

// enableLocalnetRouting tells the kernel not to treat 127/8 as a martian,
// so that connections with a source ip of 127/8 can cross a routing boundary.
func enableLocalnetRouting(ifName string) error {
	routeLocalnetPath := "net/ipv4/conf/" + ifName + "/route_localnet"
	_, err := sysctl.Sysctl(routeLocalnetPath, "1")
	return err
}

func getRoutableHostIF(containerIP net.IP) string {
	routes, err := netlink.RouteGet(containerIP)
	if err != nil {
		return ""
	}

	for _, route := range routes {
		link, err := netlink.LinkByIndex(route.LinkIndex)
		if err != nil {
			continue
		}

		return link.Attrs().Name
	}

	return ""
}

func getContainerDnatChain(name, id string) string {
	return utils.MustFormatChainNameWithPrefix(name, id, "CONTAINER-DNAT-")
}
