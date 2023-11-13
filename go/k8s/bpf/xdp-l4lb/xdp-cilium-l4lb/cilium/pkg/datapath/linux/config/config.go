package config

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	datapathOption "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/datapath/option"

	"github.com/cilium/cilium/pkg/datapath/iptables"
	"github.com/cilium/cilium/pkg/datapath/link"
	"github.com/cilium/cilium/pkg/identity"
	"github.com/cilium/cilium/pkg/labels"
	"github.com/cilium/cilium/pkg/maps/bwmap"
	"github.com/cilium/cilium/pkg/maps/ctmap"
	"github.com/cilium/cilium/pkg/maps/egressmap"
	"github.com/cilium/cilium/pkg/maps/encrypt"
	"github.com/cilium/cilium/pkg/maps/eppolicymap"
	"github.com/cilium/cilium/pkg/maps/eventsmap"
	"github.com/cilium/cilium/pkg/maps/ipmasq"
	"github.com/cilium/cilium/pkg/maps/nat"
	"github.com/cilium/cilium/pkg/maps/neighborsmap"
	"github.com/cilium/cilium/pkg/maps/recorder"
	"github.com/cilium/cilium/pkg/maps/signalmap"
	"github.com/cilium/cilium/pkg/maps/tunnel"
	"github.com/cilium/cilium/pkg/node"
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"io"
	"reflect"
	"sort"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/datapath"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/maglev"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/maps/lbmap"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/maps/lxcmap"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/option"
)

type HeaderfileWriter struct{}

func writeIncludes(w io.Writer) (int, error) {
	return fmt.Fprintf(w, "#include \"lib/utils.h\"\n\n")
}

// WriteNodeConfig 初始化时重新写字节内容 "/var/run/cilium/state/globals/node_config.h"
func (h *HeaderfileWriter) WriteNodeConfig(w io.Writer, configuration *datapath.LocalNodeConfiguration) error {
	cDefinesMap := make(map[string]string)

	// 4k buffer writer
	fw := bufio.NewWriter(w)
	writeIncludes(fw)

	fmt.Fprintf(fw, "/*\n")
	fmt.Fprintf(fw, " cilium.v4.external.str %s\n", node.GetIPv4().String())
	fmt.Fprintf(fw, " cilium.v4.internal.str %s\n", node.GetInternalIPv4Router().String())
	fmt.Fprintf(fw, " cilium.v4.nodeport.str %s\n", node.GetNodePortIPv4Addrs())
	fmt.Fprintf(fw, "\n")
	fw.WriteString(dumpRaw(defaults.RestoreV4Addr, node.GetInternalIPv4Router()))
	fmt.Fprintf(fw, " */\n\n")

	cDefinesMap["KERNEL_HZ"] = fmt.Sprintf("%d", option.Config.KernelHz)
	if option.Config.EnableIPv4 {
		ipv4GW := node.GetInternalIPv4Router()
		loopbackIPv4 := node.GetIPv4Loopback()
		ipv4Range := node.GetIPv4AllocRange()
		cDefinesMap["IPV4_GATEWAY"] = fmt.Sprintf("%#x", byteorder.HostSliceToNetwork(ipv4GW, reflect.Uint32).(uint32))
		cDefinesMap["IPV4_LOOPBACK"] = fmt.Sprintf("%#x", byteorder.HostSliceToNetwork(loopbackIPv4, reflect.Uint32).(uint32))
		cDefinesMap["IPV4_MASK"] = fmt.Sprintf("%#x", byteorder.HostSliceToNetwork(ipv4Range.Mask, reflect.Uint32).(uint32))

		if option.Config.EnableIPv4FragmentsTracking {
			cDefinesMap["ENABLE_IPV4_FRAGMENTS"] = "1"
			cDefinesMap["IPV4_FRAG_DATAGRAMS_MAP"] = fragmap.MapName
			cDefinesMap["CILIUM_IPV4_FRAG_MAP_MAX_ENTRIES"] = fmt.Sprintf("%d", option.Config.FragmentsMapEntries)
		}
	}
	if nat46Range := option.Config.NAT46Prefix; nat46Range != nil {
		fw.WriteString(FmtDefineAddress("NAT46_PREFIX", nat46Range.IP))
	}
	cDefinesMap["HOST_ID"] = fmt.Sprintf("%d", identity.GetReservedID(labels.IDNameHost))
	cDefinesMap["WORLD_ID"] = fmt.Sprintf("%d", identity.GetReservedID(labels.IDNameWorld))
	cDefinesMap["HEALTH_ID"] = fmt.Sprintf("%d", identity.GetReservedID(labels.IDNameHealth))
	cDefinesMap["UNMANAGED_ID"] = fmt.Sprintf("%d", identity.GetReservedID(labels.IDNameUnmanaged))
	cDefinesMap["INIT_ID"] = fmt.Sprintf("%d", identity.GetReservedID(labels.IDNameInit))
	cDefinesMap["LOCAL_NODE_ID"] = fmt.Sprintf("%d", identity.GetLocalNodeID())
	cDefinesMap["REMOTE_NODE_ID"] = fmt.Sprintf("%d", identity.GetReservedID(labels.IDNameRemoteNode))
	cDefinesMap["CILIUM_LB_MAP_MAX_ENTRIES"] = fmt.Sprintf("%d", lbmap.MaxEntries)
	cDefinesMap["TUNNEL_MAP"] = tunnel.MapName
	cDefinesMap["TUNNEL_ENDPOINT_MAP_SIZE"] = fmt.Sprintf("%d", tunnel.MaxEntries)
	cDefinesMap["ENDPOINTS_MAP"] = lxcmap.MapName // __lookup_ip4_endpoint endpoints.h
	cDefinesMap["ENDPOINTS_MAP_SIZE"] = fmt.Sprintf("%d", lxcmap.MaxEntries)
	cDefinesMap["METRICS_MAP"] = metricsmap.MapName
	cDefinesMap["METRICS_MAP_SIZE"] = fmt.Sprintf("%d", metricsmap.MaxEntries)
	cDefinesMap["POLICY_MAP_SIZE"] = fmt.Sprintf("%d", policymap.MaxEntries)
	cDefinesMap["IPCACHE_MAP"] = ipcachemap.Name
	cDefinesMap["IPCACHE_MAP_SIZE"] = fmt.Sprintf("%d", ipcachemap.MaxEntries)
	// TODO(anfernee): Update Documentation/concepts/ebpf/maps.rst when egress gateway support is merged.
	cDefinesMap["EGRESS_POLICY_MAP"] = egressmap.PolicyMapName
	cDefinesMap["EGRESS_POLICY_MAP_SIZE"] = fmt.Sprintf("%d", egressmap.MaxPolicyEntries)
	cDefinesMap["POLICY_PROG_MAP_SIZE"] = fmt.Sprintf("%d", policymap.PolicyCallMaxEntries)
	cDefinesMap["SOCKOPS_MAP_SIZE"] = fmt.Sprintf("%d", sockmap.MaxEntries)
	cDefinesMap["ENCRYPT_MAP"] = encrypt.MapName
	cDefinesMap["CT_CONNECTION_LIFETIME_TCP"] = fmt.Sprintf("%d", int64(option.Config.CTMapEntriesTimeoutTCP.Seconds()))
	cDefinesMap["CT_CONNECTION_LIFETIME_NONTCP"] = fmt.Sprintf("%d", int64(option.Config.CTMapEntriesTimeoutAny.Seconds()))
	cDefinesMap["CT_SERVICE_LIFETIME_TCP"] = fmt.Sprintf("%d", int64(option.Config.CTMapEntriesTimeoutSVCTCP.Seconds()))
	cDefinesMap["CT_SERVICE_LIFETIME_NONTCP"] = fmt.Sprintf("%d", int64(option.Config.CTMapEntriesTimeoutSVCAny.Seconds()))
	cDefinesMap["CT_SYN_TIMEOUT"] = fmt.Sprintf("%d", int64(option.Config.CTMapEntriesTimeoutSYN.Seconds()))
	cDefinesMap["CT_CLOSE_TIMEOUT"] = fmt.Sprintf("%d", int64(option.Config.CTMapEntriesTimeoutFIN.Seconds()))
	cDefinesMap["CT_REPORT_INTERVAL"] = fmt.Sprintf("%d", int64(option.Config.MonitorAggregationInterval.Seconds()))
	cDefinesMap["CT_REPORT_FLAGS"] = fmt.Sprintf("%#04x", int64(option.Config.MonitorAggregationFlags))
	if option.Config.DatapathMode == datapathOption.DatapathModeIpvlan {
		cDefinesMap["ENABLE_EXTRA_HOST_DEV"] = "1"
	}
	if option.Config.PreAllocateMaps {
		cDefinesMap["PREALLOCATE_MAPS"] = "1"
	}
	cDefinesMap["EVENTS_MAP"] = eventsmap.MapName
	cDefinesMap["SIGNAL_MAP"] = signalmap.MapName
	cDefinesMap["POLICY_CALL_MAP"] = policymap.PolicyCallMapName
	cDefinesMap["EP_POLICY_MAP"] = eppolicymap.MapName
	cDefinesMap["LB6_REVERSE_NAT_MAP"] = "cilium_lb6_reverse_nat"
	cDefinesMap["LB6_SERVICES_MAP_V2"] = "cilium_lb6_services_v2"
	cDefinesMap["LB6_BACKEND_MAP"] = "cilium_lb6_backends"
	cDefinesMap["LB6_REVERSE_NAT_SK_MAP"] = lbmap.SockRevNat6MapName
	cDefinesMap["LB6_REVERSE_NAT_SK_MAP_SIZE"] = fmt.Sprintf("%d", lbmap.MaxSockRevNat6MapEntries)
	cDefinesMap["LB4_REVERSE_NAT_MAP"] = "cilium_lb4_reverse_nat"
	cDefinesMap["LB4_SERVICES_MAP_V2"] = "cilium_lb4_services_v2"
	cDefinesMap["LB4_BACKEND_MAP"] = "cilium_lb4_backends"
	cDefinesMap["LB4_REVERSE_NAT_SK_MAP"] = lbmap.SockRevNat4MapName
	cDefinesMap["LB4_REVERSE_NAT_SK_MAP_SIZE"] = fmt.Sprintf("%d", lbmap.MaxSockRevNat4MapEntries)
	if option.Config.EnableSessionAffinity {
		cDefinesMap["ENABLE_SESSION_AFFINITY"] = "1"
		cDefinesMap["LB_AFFINITY_MATCH_MAP"] = lbmap.AffinityMatchMapName
		if option.Config.EnableIPv4 {
			cDefinesMap["LB4_AFFINITY_MAP"] = lbmap.Affinity4MapName
		}
		//if option.Config.EnableIPv6 {
		//	cDefinesMap["LB6_AFFINITY_MAP"] = lbmap.Affinity6MapName
		//}
	}
	cDefinesMap["TRACE_PAYLOAD_LEN"] = fmt.Sprintf("%dULL", option.Config.TracePayloadlen)
	cDefinesMap["MTU"] = fmt.Sprintf("%d", cfg.MtuConfig.GetDeviceMTU())
	if option.Config.EnableIPv4 {
		cDefinesMap["ENABLE_IPV4"] = "1"
	}
	if option.Config.EnableIPSec {
		cDefinesMap["ENABLE_IPSEC"] = "1"
	}
	if option.Config.EnableWireguard {
		cDefinesMap["ENABLE_WIREGUARD"] = "1"
	}
	if option.Config.InstallIptRules || iptables.KernelHasNetfilter() {
		cDefinesMap["NO_REDIRECT"] = "1"
	}
	if option.Config.EnableBPFTProxy {
		cDefinesMap["ENABLE_TPROXY"] = "1"
	}
	if option.Config.EncryptNode {
		cDefinesMap["ENCRYPT_NODE"] = "1"
	}
	if option.Config.EnableXDPPrefilter {
		cDefinesMap["ENABLE_PREFILTER"] = "1"
	}
	if option.Config.EnableEgressGateway {
		cDefinesMap["ENABLE_EGRESS_GATEWAY"] = "1"
	}
	if option.Config.EnableEndpointRoutes {
		cDefinesMap["ENABLE_ENDPOINT_ROUTES"] = "1"
	}
	if option.Config.EnableHostReachableServices {
		if option.Config.EnableHostServicesTCP {
			cDefinesMap["ENABLE_HOST_SERVICES_TCP"] = "1"
		}
		if option.Config.EnableHostServicesUDP {
			cDefinesMap["ENABLE_HOST_SERVICES_UDP"] = "1"
		}
		if option.Config.EnableHostServicesTCP && option.Config.EnableHostServicesUDP && !option.Config.BPFSocketLBHostnsOnly {
			cDefinesMap["ENABLE_HOST_SERVICES_FULL"] = "1"
		}
		if option.Config.EnableHostServicesPeer {
			cDefinesMap["ENABLE_HOST_SERVICES_PEER"] = "1"
		}

		if option.Config.BPFSocketLBHostnsOnly {
			cDefinesMap["ENABLE_SOCKET_LB_HOST_ONLY"] = "1"
		}
	}
	if option.Config.EnableNodePort {
		if option.Config.EnableHealthDatapath {
			cDefinesMap["ENABLE_HEALTH_CHECK"] = "1"
		}
		if option.Config.EnableMKE && option.Config.EnableHostReachableServices {
			cDefinesMap["ENABLE_MKE"] = "1"
			cDefinesMap["MKE_HOST"] = fmt.Sprintf("%d", option.HostExtensionMKE)
		}
		if option.Config.EnableRecorder {
			cDefinesMap["ENABLE_CAPTURE"] = "1"
			if option.Config.EnableIPv4 {
				cDefinesMap["CAPTURE4_RULES"] = recorder.MapNameWcard4
				cDefinesMap["CAPTURE4_SIZE"] = fmt.Sprintf("%d", recorder.MapSize)
			}
			if option.Config.EnableIPv6 {
				cDefinesMap["CAPTURE6_RULES"] = recorder.MapNameWcard6
				cDefinesMap["CAPTURE6_SIZE"] = fmt.Sprintf("%d", recorder.MapSize)
			}
		}
		cDefinesMap["ENABLE_NODEPORT"] = "1"
		cDefinesMap["ENABLE_LOADBALANCER"] = "1"
		if option.Config.EnableIPv4 {
			cDefinesMap["NODEPORT_NEIGH4"] = neighborsmap.Map4Name
			cDefinesMap["NODEPORT_NEIGH4_SIZE"] = fmt.Sprintf("%d", option.Config.NeighMapEntriesGlobal)
			if option.Config.EnableHealthDatapath {
				cDefinesMap["LB4_HEALTH_MAP"] = lbmap.HealthProbe4MapName
			}
		}
		if option.Config.EnableIPv6 {
			cDefinesMap["NODEPORT_NEIGH6"] = neighborsmap.Map6Name
			cDefinesMap["NODEPORT_NEIGH6_SIZE"] = fmt.Sprintf("%d", option.Config.NeighMapEntriesGlobal)
			if option.Config.EnableHealthDatapath {
				cDefinesMap["LB6_HEALTH_MAP"] = lbmap.HealthProbe6MapName
			}
		}
		const (
			dsrEncapInv = iota
			dsrEncapNone
			dsrEncapIPIP
		)
		const (
			dsrL4XlateInv = iota
			dsrL4XlateFrontend
			dsrL4XlateBackend
		)
		cDefinesMap["DSR_ENCAP_IPIP"] = fmt.Sprintf("%d", dsrEncapIPIP)
		cDefinesMap["DSR_ENCAP_NONE"] = fmt.Sprintf("%d", dsrEncapNone)
		cDefinesMap["DSR_XLATE_FRONTEND"] = fmt.Sprintf("%d", dsrL4XlateFrontend)
		cDefinesMap["DSR_XLATE_BACKEND"] = fmt.Sprintf("%d", dsrL4XlateBackend)
		if option.Config.NodePortMode == option.NodePortModeDSR ||
			option.Config.NodePortMode == option.NodePortModeHybrid {
			cDefinesMap["ENABLE_DSR"] = "1"
			if option.Config.LoadBalancerPMTUDiscovery {
				cDefinesMap["ENABLE_DSR_ICMP_ERRORS"] = "1"
			}
			if option.Config.NodePortMode == option.NodePortModeHybrid {
				cDefinesMap["ENABLE_DSR_HYBRID"] = "1"
			}
			if option.Config.LoadBalancerDSRDispatch == option.DSRDispatchOption {
				cDefinesMap["DSR_ENCAP_MODE"] = fmt.Sprintf("%d", dsrEncapNone)
			} else if option.Config.LoadBalancerDSRDispatch == option.DSRDispatchIPIP {
				cDefinesMap["DSR_ENCAP_MODE"] = fmt.Sprintf("%d", dsrEncapIPIP)
			}
			if option.Config.LoadBalancerDSRDispatch == option.DSRDispatchIPIP {
				if option.Config.LoadBalancerDSRL4Xlate == option.DSRL4XlateFrontend {
					cDefinesMap["DSR_XLATE_MODE"] = fmt.Sprintf("%d", dsrL4XlateFrontend)
				} else if option.Config.LoadBalancerDSRL4Xlate == option.DSRL4XlateBackend {
					cDefinesMap["DSR_XLATE_MODE"] = fmt.Sprintf("%d", dsrL4XlateBackend)
				}
			} else {
				cDefinesMap["DSR_XLATE_MODE"] = fmt.Sprintf("%d", dsrL4XlateInv)
			}
		} else {
			cDefinesMap["DSR_ENCAP_MODE"] = fmt.Sprintf("%d", dsrEncapInv)
			cDefinesMap["DSR_XLATE_MODE"] = fmt.Sprintf("%d", dsrL4XlateInv)
		}
		if option.Config.EnableIPv4 {
			if option.Config.LoadBalancerRSSv4CIDR != "" {
				ipv4 := byteorder.HostSliceToNetwork(option.Config.LoadBalancerRSSv4.IP, reflect.Uint32).(uint32)
				ones, _ := option.Config.LoadBalancerRSSv4.Mask.Size()
				cDefinesMap["IPV4_RSS_PREFIX"] = fmt.Sprintf("%d", ipv4)
				cDefinesMap["IPV4_RSS_PREFIX_BITS"] = fmt.Sprintf("%d", ones)
			} else {
				cDefinesMap["IPV4_RSS_PREFIX"] = "IPV4_DIRECT_ROUTING"
				cDefinesMap["IPV4_RSS_PREFIX_BITS"] = "32"
			}
		}
		if option.Config.NodePortAcceleration != option.NodePortAccelerationDisabled {
			cDefinesMap["ENABLE_NODEPORT_ACCELERATION"] = "1"
		}
		if option.Config.NodePortHairpin {
			cDefinesMap["ENABLE_NODEPORT_HAIRPIN"] = "1"
		}
		if option.Config.EnableExternalIPs {
			cDefinesMap["ENABLE_EXTERNAL_IP"] = "1"
		}
		if option.Config.EnableHostPort {
			cDefinesMap["ENABLE_HOSTPORT"] = "1"
		}
		if !option.Config.EnableHostLegacyRouting {
			cDefinesMap["ENABLE_REDIRECT_FAST"] = "1"
		}
		if option.Config.EnableSVCSourceRangeCheck {
			cDefinesMap["ENABLE_SRC_RANGE_CHECK"] = "1"
			if option.Config.EnableIPv4 {
				cDefinesMap["LB4_SRC_RANGE_MAP"] = lbmap.SourceRange4MapName
				cDefinesMap["LB4_SRC_RANGE_MAP_SIZE"] =
					fmt.Sprintf("%d", lbmap.SourceRange4Map.MapInfo.MaxEntries)
			}
			if option.Config.EnableIPv6 {
				cDefinesMap["LB6_SRC_RANGE_MAP"] = lbmap.SourceRange6MapName
				cDefinesMap["LB6_SRC_RANGE_MAP_SIZE"] =
					fmt.Sprintf("%d", lbmap.SourceRange6Map.MapInfo.MaxEntries)
			}
		}
		if option.Config.EnableBPFBypassFIBLookup {
			cDefinesMap["ENABLE_FIB_LOOKUP_BYPASS"] = "1"
		}

		cDefinesMap["NODEPORT_PORT_MIN"] = fmt.Sprintf("%d", option.Config.NodePortMin)
		cDefinesMap["NODEPORT_PORT_MAX"] = fmt.Sprintf("%d", option.Config.NodePortMax)
		cDefinesMap["NODEPORT_PORT_MIN_NAT"] = fmt.Sprintf("%d", option.Config.NodePortMax+1)
		cDefinesMap["NODEPORT_PORT_MAX_NAT"] = "65535"

		macByIfIndexMacro, isL3DevMacro, err := devMacros()
		if err != nil {
			return err
		}
		cDefinesMap["NATIVE_DEV_MAC_BY_IFINDEX(IFINDEX)"] = macByIfIndexMacro
		cDefinesMap["IS_L3_DEV(ifindex)"] = isL3DevMacro
	}
	const (
		selectionRandom = iota + 1
		selectionMaglev
	)
	cDefinesMap["LB_SELECTION_RANDOM"] = fmt.Sprintf("%d", selectionRandom)
	cDefinesMap["LB_SELECTION_MAGLEV"] = fmt.Sprintf("%d", selectionMaglev)
	if option.Config.NodePortAlg == option.NodePortAlgRandom {
		cDefinesMap["LB_SELECTION"] = fmt.Sprintf("%d", selectionRandom)
	} else if option.Config.NodePortAlg == option.NodePortAlgMaglev {
		cDefinesMap["LB_SELECTION"] = fmt.Sprintf("%d", selectionMaglev)
		cDefinesMap["LB_MAGLEV_LUT_SIZE"] = fmt.Sprintf("%d", option.Config.MaglevTableSize)
		if option.Config.EnableIPv6 {
			cDefinesMap["LB6_MAGLEV_MAP_INNER"] = lbmap.MaglevInner6MapName
			cDefinesMap["LB6_MAGLEV_MAP_OUTER"] = lbmap.MaglevOuter6MapName
		}
		if option.Config.EnableIPv4 {
			cDefinesMap["LB4_MAGLEV_MAP_INNER"] = lbmap.MaglevInner4MapName
			cDefinesMap["LB4_MAGLEV_MAP_OUTER"] = lbmap.MaglevOuter4MapName
		}
	}
	cDefinesMap["HASH_INIT4_SEED"] = fmt.Sprintf("%d", maglev.SeedJhash0)
	cDefinesMap["HASH_INIT6_SEED"] = fmt.Sprintf("%d", maglev.SeedJhash1)
	if option.Config.EnableNodePort {
		directRoutingIface := option.Config.DirectRoutingDevice
		directRoutingIfIndex, err := link.GetIfIndex(directRoutingIface)
		if err != nil {
			return err
		}
		cDefinesMap["DIRECT_ROUTING_DEV_IFINDEX"] = fmt.Sprintf("%d", directRoutingIfIndex)

		if option.Config.EnableIPv4 {
			ip, ok := node.GetNodePortIPv4AddrsWithDevices()[directRoutingIface]
			if !ok {
				log.WithFields(logrus.Fields{
					"directRoutingIface": directRoutingIface,
				}).Fatal("NodePort enabled but direct routing device's IPv4 address not found")
			}

			ipv4 := byteorder.HostSliceToNetwork(ip, reflect.Uint32).(uint32)
			cDefinesMap["IPV4_DIRECT_ROUTING"] = fmt.Sprintf("%d", ipv4)
		}

		if option.Config.LoadBalancerPreserveWorldID {
			cDefinesMap["PRESERVE_WORLD_ID"] = "1"
		}
	} else {
		cDefinesMap["DIRECT_ROUTING_DEV_IFINDEX"] = "0"
		if option.Config.EnableIPv4 {
			cDefinesMap["IPV4_DIRECT_ROUTING"] = "0"
		}
	}
	if option.Config.ResetQueueMapping {
		cDefinesMap["RESET_QUEUES"] = "1"
	}
	if option.Config.EnableBandwidthManager {
		cDefinesMap["ENABLE_BANDWIDTH_MANAGER"] = "1"
		cDefinesMap["THROTTLE_MAP"] = bwmap.MapName
		cDefinesMap["THROTTLE_MAP_SIZE"] = fmt.Sprintf("%d", bwmap.MapSize)
	}
	if option.Config.EnableHostFirewall {
		cDefinesMap["ENABLE_HOST_FIREWALL"] = "1"
	}
	if option.Config.EnableIPSec {
		a := byteorder.HostSliceToNetwork(node.GetIPv4(), reflect.Uint32).(uint32)
		cDefinesMap["IPV4_ENCRYPT_IFACE"] = fmt.Sprintf("%d", a)
		if iface := option.Config.EncryptInterface; len(iface) != 0 {
			link, err := netlink.LinkByName(iface[0])
			if err == nil {
				cDefinesMap["ENCRYPT_IFACE"] = fmt.Sprintf("%d", link.Attrs().Index)
			}
		}
		// If we are using EKS or AKS IPAM modes, we should use IP_POOLS
		// datapath as the pod subnets will be auto-discovered later at
		// runtime.
		if option.Config.IPAM == ipamOption.IPAMENI ||
			option.Config.IPAM == ipamOption.IPAMAzure ||
			option.Config.IsPodSubnetsDefined() {
			cDefinesMap["IP_POOLS"] = "1"
		}
	}
	if option.Config.EnableNodePort {
		if option.Config.EnableIPv4 {
			cDefinesMap["SNAT_MAPPING_IPV4"] = nat.MapNameSnat4Global
			cDefinesMap["SNAT_MAPPING_IPV4_SIZE"] = fmt.Sprintf("%d", option.Config.NATMapEntriesGlobal)
		}

		if option.Config.EnableIPv6 {
			cDefinesMap["SNAT_MAPPING_IPV6"] = nat.MapNameSnat6Global
			cDefinesMap["SNAT_MAPPING_IPV6_SIZE"] = fmt.Sprintf("%d", option.Config.NATMapEntriesGlobal)
		}

		if option.Config.EnableIPv4Masquerade && option.Config.EnableBPFMasquerade {
			cDefinesMap["ENABLE_MASQUERADE"] = "1"
			cidr := datapath.RemoteSNATDstAddrExclusionCIDRv4()
			cDefinesMap["IPV4_SNAT_EXCLUSION_DST_CIDR"] =
				fmt.Sprintf("%#x", byteorder.HostSliceToNetwork(cidr.IP, reflect.Uint32).(uint32))
			ones, _ := cidr.Mask.Size()
			cDefinesMap["IPV4_SNAT_EXCLUSION_DST_CIDR_LEN"] = fmt.Sprintf("%d", ones)

			// ip-masq-agent depends on bpf-masq
			if option.Config.EnableIPMasqAgent {
				cDefinesMap["ENABLE_IP_MASQ_AGENT"] = "1"
				cDefinesMap["IP_MASQ_AGENT_IPV4"] = ipmasq.MapName
			}
		}

		ctmap.WriteBPFMacros(fw, nil)
	}
	if option.Config.AllowICMPFragNeeded {
		cDefinesMap["ALLOW_ICMP_FRAG_NEEDED"] = "1"
	}
	if option.Config.ClockSource == option.ClockSourceJiffies {
		cDefinesMap["ENABLE_JIFFIES"] = "1"
	}
	if option.Config.EnableIdentityMark {
		cDefinesMap["ENABLE_IDENTITY_MARK"] = "1"
	}
	if option.Config.EnableCustomCalls {
		cDefinesMap["ENABLE_CUSTOM_CALLS"] = "1"
	}

	// sort keys 方便 base64
	var keys []string
	for key := range cDefinesMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Fprintf(fw, "#define %s %s\n", key, cDefinesMap[key])
	}

	jsonBytes, err := json.Marshal(cDefinesMap)
	if err == nil {
		encodedConfig := base64.StdEncoding.EncodeToString(jsonBytes)
		fmt.Fprintf(fw, "\n// JSON_OUTPUT: %s\n", encodedConfig)
	}

	return fw.Flush()
}

func (h *HeaderfileWriter) WriteNetdevConfig(writer io.Writer, configuration datapath.DeviceConfiguration) error {
	//TODO implement me
	panic("implement me")
}

func (h *HeaderfileWriter) WriteTemplateConfig(w io.Writer, cfg datapath.EndpointConfiguration) error {
	//TODO implement me
	panic("implement me")
}

func (h *HeaderfileWriter) WriteEndpointConfig(w io.Writer, cfg datapath.EndpointConfiguration) error {
	//TODO implement me
	panic("implement me")
}
