package config

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/cilium/cilium/pkg/datapath/iptables"
	"github.com/cilium/cilium/pkg/datapath/link"
	"github.com/cilium/cilium/pkg/identity"
	"github.com/cilium/cilium/pkg/labels"
	"github.com/cilium/cilium/pkg/maps/ctmap"
	"github.com/cilium/cilium/pkg/maps/encrypt"
	"github.com/cilium/cilium/pkg/maps/eppolicymap"
	"github.com/cilium/cilium/pkg/maps/eventsmap"
	"github.com/cilium/cilium/pkg/maps/fragmap"
	"github.com/cilium/cilium/pkg/maps/ipmasq"
	"github.com/cilium/cilium/pkg/maps/lbmap"
	"github.com/cilium/cilium/pkg/maps/metricsmap"
	"github.com/cilium/cilium/pkg/maps/nat"
	"github.com/cilium/cilium/pkg/maps/neighborsmap"
	"github.com/cilium/cilium/pkg/maps/policymap"
	"github.com/cilium/cilium/pkg/maps/signalmap"
	"github.com/cilium/cilium/pkg/maps/sockmap"
	"github.com/cilium/cilium/pkg/maps/tunnel"
	"github.com/cilium/cilium/pkg/node"
	"github.com/cilium/cilium/pkg/option"
	"io"
	"reflect"

	"k8s-lx1036/k8s/network/cilium/cilium/pkg/config/defaults"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/datapath"
)

// HeaderfileWriter is a wrapper type which implements datapath.ConfigWriter.
// It manages writing of configuration of datapath program headerfiles.
type HeaderfileWriter struct{}

// WriteNodeConfig writes the local node configuration to the specified writer.
// INFO: @see tools/state/globals/node_config.h
func (h *HeaderfileWriter) WriteNodeConfig(w io.Writer, cfg *datapath.LocalNodeConfiguration) error {
	cDefinesMap := make(map[string]string)

	fw := bufio.NewWriter(w)

	writeIncludes(w)

	fmt.Fprintf(fw, "/*\n")
	fmt.Fprintf(fw, " cilium.v4.external.str %s\n", node.GetExternalIPv4().String())
	fmt.Fprintf(fw, " cilium.v4.internal.str %s\n", node.GetInternalIPv4().String())
	fmt.Fprintf(fw, " cilium.v4.nodeport.str %s\n", node.GetNodePortIPv4Addrs())
	fmt.Fprintf(fw, "\n")
	fw.WriteString(dumpRaw(defaults.RestoreV4Addr, node.GetInternalIPv4()))
	fmt.Fprintf(fw, " */\n\n")

	cDefinesMap["KERNEL_HZ"] = fmt.Sprintf("%d", option.Config.KernelHz)
	ipv4GW := node.GetInternalIPv4()
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
	if nat46Range := option.Config.NAT46Prefix; nat46Range != nil {
		fw.WriteString(FmtDefineAddress("NAT46_PREFIX", nat46Range.IP))
	}
	cDefinesMap["HOST_ID"] = fmt.Sprintf("%d", identity.GetReservedID(labels.IDNameHost))
	cDefinesMap["WORLD_ID"] = fmt.Sprintf("%d", identity.GetReservedID(labels.IDNameWorld))
	cDefinesMap["HEALTH_ID"] = fmt.Sprintf("%d", identity.GetReservedID(labels.IDNameHealth))
	cDefinesMap["UNMANAGED_ID"] = fmt.Sprintf("%d", identity.GetReservedID(labels.IDNameUnmanaged))
	cDefinesMap["INIT_ID"] = fmt.Sprintf("%d", identity.GetReservedID(labels.IDNameInit))
	cDefinesMap["REMOTE_NODE_ID"] = fmt.Sprintf("%d", identity.GetReservedID(labels.IDNameRemoteNode))
	cDefinesMap["CILIUM_LB_MAP_MAX_ENTRIES"] = fmt.Sprintf("%d", lbmap.MaxEntries)
	cDefinesMap["TUNNEL_MAP"] = tunnel.MapName
	cDefinesMap["TUNNEL_ENDPOINT_MAP_SIZE"] = fmt.Sprintf("%d", tunnel.MaxEntries)
	cDefinesMap["ENDPOINTS_MAP"] = lxcmap.MapName
	cDefinesMap["ENDPOINTS_MAP_SIZE"] = fmt.Sprintf("%d", lxcmap.MaxEntries)
	cDefinesMap["METRICS_MAP"] = metricsmap.MapName
	cDefinesMap["METRICS_MAP_SIZE"] = fmt.Sprintf("%d", metricsmap.MaxEntries)
	cDefinesMap["POLICY_MAP_SIZE"] = fmt.Sprintf("%d", policymap.MaxEntries)
	cDefinesMap["IPCACHE_MAP"] = ipcachemap.Name
	cDefinesMap["IPCACHE_MAP_SIZE"] = fmt.Sprintf("%d", ipcachemap.MaxEntries)
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
		cDefinesMap["ENABLE_SECCTX_FROM_IPCACHE"] = "1"
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
	}
	cDefinesMap["TRACE_PAYLOAD_LEN"] = fmt.Sprintf("%dULL", option.Config.TracePayloadlen)
	cDefinesMap["MTU"] = fmt.Sprintf("%d", cfg.MtuConfig.GetDeviceMTU())
	if option.Config.EnableIPv4 {
		cDefinesMap["ENABLE_IPV4"] = "1"
	}
	if option.Config.EnableIPv6 {
		cDefinesMap["ENABLE_IPV6"] = "1"
	}
	if option.Config.EnableIPSec {
		cDefinesMap["ENABLE_IPSEC"] = "1"
	}
	if option.Config.InstallIptRules || iptables.KernelHasNetfilter() {
		cDefinesMap["NO_REDIRECT"] = "1"
	}
	if option.Config.EncryptNode {
		cDefinesMap["ENCRYPT_NODE"] = "1"
	}
	if option.Config.DevicePreFilter != "undefined" {
		cDefinesMap["ENABLE_PREFILTER"] = "1"
	}
	if !option.Config.DisableK8sServices {
		cDefinesMap["ENABLE_SERVICES"] = "1"
	}
	if option.Config.EnableHostReachableServices {
		if option.Config.EnableHostServicesTCP {
			cDefinesMap["ENABLE_HOST_SERVICES_TCP"] = "1"
		}
		if option.Config.EnableHostServicesUDP {
			cDefinesMap["ENABLE_HOST_SERVICES_UDP"] = "1"
		}
		if option.Config.EnableHostServicesTCP && option.Config.EnableHostServicesUDP {
			cDefinesMap["ENABLE_HOST_SERVICES_FULL"] = "1"
		}
		if option.Config.EnableHostServicesPeer {
			cDefinesMap["ENABLE_HOST_SERVICES_PEER"] = "1"
		}
	}
	if option.Config.EnableNodePort {
		cDefinesMap["ENABLE_NODEPORT"] = "1"
		cDefinesMap["ENABLE_LOADBALANCER"] = "1"

		if option.Config.EnableIPv4 {
			cDefinesMap["NODEPORT_NEIGH4"] = neighborsmap.Map4Name
			cDefinesMap["NODEPORT_NEIGH4_SIZE"] = fmt.Sprintf("%d", option.Config.NeighMapEntriesGlobal)
		}
		if option.Config.EnableIPv6 {
			cDefinesMap["NODEPORT_NEIGH6"] = neighborsmap.Map6Name
			cDefinesMap["NODEPORT_NEIGH6_SIZE"] = fmt.Sprintf("%d", option.Config.NeighMapEntriesGlobal)
		}
		if option.Config.NodePortMode == option.NodePortModeDSR ||
			option.Config.NodePortMode == option.NodePortModeHybrid {
			cDefinesMap["ENABLE_DSR"] = "1"
			if option.Config.NodePortMode == option.NodePortModeHybrid {
				cDefinesMap["ENABLE_DSR_HYBRID"] = "1"
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

		cDefinesMap["NODEPORT_PORT_MIN"] = fmt.Sprintf("%d", option.Config.NodePortMin)
		cDefinesMap["NODEPORT_PORT_MAX"] = fmt.Sprintf("%d", option.Config.NodePortMax)
		cDefinesMap["NODEPORT_PORT_MIN_NAT"] = fmt.Sprintf("%d", option.Config.NodePortMax+1)
		cDefinesMap["NODEPORT_PORT_MAX_NAT"] = "65535"

		directRoutingIface := option.Config.DirectRoutingDevice
		directRoutingIfIndex, err := link.GetIfIndex(directRoutingIface)
		if err != nil {
			return err
		}
		cDefinesMap["DIRECT_ROUTING_DEV_IFINDEX"] = fmt.Sprintf("%d", directRoutingIfIndex)
		if option.Config.EnableIPv4 {
			nodePortIPv4Addrs := node.GetNodePortIPv4AddrsWithDevices()
			ipv4 := byteorder.HostSliceToNetwork(nodePortIPv4Addrs[directRoutingIface], reflect.Uint32).(uint32)
			cDefinesMap["IPV4_DIRECT_ROUTING"] = fmt.Sprintf("%d", ipv4)
		}
	} else {
		cDefinesMap["DIRECT_ROUTING_DEV_IFINDEX"] = "0"
		if option.Config.EnableIPv4 {
			cDefinesMap["IPV4_DIRECT_ROUTING"] = "0"
		}
	}
	if option.Config.EnableHostFirewall {
		cDefinesMap["ENABLE_HOST_FIREWALL"] = "1"
	}
	if option.Config.IsPodSubnetsDefined() {
		cDefinesMap["IP_POOLS"] = "1"
	}
	if option.Config.EnableNodePort { // nat
		if option.Config.EnableIPv4 {
			cDefinesMap["SNAT_MAPPING_IPV4"] = nat.MapNameSnat4Global
			cDefinesMap["SNAT_MAPPING_IPV4_SIZE"] = fmt.Sprintf("%d", option.Config.NATMapEntriesGlobal)
		}

		if option.Config.EnableBPFMasquerade && option.Config.EnableIPv4 {
			cDefinesMap["ENABLE_MASQUERADE"] = "1"
			cidr := datapath.RemoteSNATDstAddrExclusionCIDR()
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
	if option.Config.PolicyAuditMode {
		cDefinesMap["POLICY_AUDIT_MODE"] = "1"
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

	keys := []string{}
	for key := range cDefinesMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Fprintf(fw, "#define %s %s\n", key, cDefinesMap[key])
	}

	jsonBytes, err := json.Marshal(cDefinesMap)
	if err == nil {
		// We don't care if some error occurs while marshaling the map.
		// In such cases we skip embedding the base64 encoded JSON configuration
		// to the writer.
		encodedConfig := base64.StdEncoding.EncodeToString(jsonBytes)
		fmt.Fprintf(fw, "\n// JSON_OUTPUT: %s\n", encodedConfig)
	}

	return fw.Flush()
}

func writeIncludes(w io.Writer) (int, error) {
	return fmt.Fprintf(w, "#include \"lib/utils.h\"\n\n")
}
