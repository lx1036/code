package config

import "github.com/osrg/gobgp/pkg/packet/bgp"

// https://www.juniper.net/documentation/cn/zh/software/junos/bgp/topics/topic-map/multiprotocol-bgp.html

// AFI: Address Family Identifiers 地址族指示符
// SAFI: Subsequent address family identifiers 后续地址族标识符

// typedef for typedef openconfig-types:percentage.
type Percentage uint8

// typedef for identity rpol:default-policy-type.
// type used to specify default route disposition in
// a policy chain.
type DefaultPolicyType string

const (
	DEFAULT_POLICY_TYPE_ACCEPT_ROUTE DefaultPolicyType = "accept-route"
	DEFAULT_POLICY_TYPE_REJECT_ROUTE DefaultPolicyType = "reject-route"
)

// typedef for identity bgp-types:afi-safi-type.
// Base identity type for AFI,SAFI tuples for BGP-4.
type AfiSafiType string

const (
	AFI_SAFI_TYPE_IPV4_UNICAST          AfiSafiType = "ipv4-unicast"
	AFI_SAFI_TYPE_IPV6_UNICAST          AfiSafiType = "ipv6-unicast"
	AFI_SAFI_TYPE_IPV4_LABELLED_UNICAST AfiSafiType = "ipv4-labelled-unicast"
	AFI_SAFI_TYPE_IPV6_LABELLED_UNICAST AfiSafiType = "ipv6-labelled-unicast"
	AFI_SAFI_TYPE_L3VPN_IPV4_UNICAST    AfiSafiType = "l3vpn-ipv4-unicast"
	AFI_SAFI_TYPE_L3VPN_IPV6_UNICAST    AfiSafiType = "l3vpn-ipv6-unicast"
	AFI_SAFI_TYPE_L3VPN_IPV4_MULTICAST  AfiSafiType = "l3vpn-ipv4-multicast"
	AFI_SAFI_TYPE_L3VPN_IPV6_MULTICAST  AfiSafiType = "l3vpn-ipv6-multicast"
	AFI_SAFI_TYPE_L2VPN_VPLS            AfiSafiType = "l2vpn-vpls"
	AFI_SAFI_TYPE_L2VPN_EVPN            AfiSafiType = "l2vpn-evpn"
	AFI_SAFI_TYPE_IPV4_MULTICAST        AfiSafiType = "ipv4-multicast"
	AFI_SAFI_TYPE_IPV6_MULTICAST        AfiSafiType = "ipv6-multicast"
	AFI_SAFI_TYPE_RTC                   AfiSafiType = "rtc"
	AFI_SAFI_TYPE_IPV4_ENCAP            AfiSafiType = "ipv4-encap"
	AFI_SAFI_TYPE_IPV6_ENCAP            AfiSafiType = "ipv6-encap"
	AFI_SAFI_TYPE_IPV4_FLOWSPEC         AfiSafiType = "ipv4-flowspec"
	AFI_SAFI_TYPE_L3VPN_IPV4_FLOWSPEC   AfiSafiType = "l3vpn-ipv4-flowspec"
	AFI_SAFI_TYPE_IPV6_FLOWSPEC         AfiSafiType = "ipv6-flowspec"
	AFI_SAFI_TYPE_L3VPN_IPV6_FLOWSPEC   AfiSafiType = "l3vpn-ipv6-flowspec"
	AFI_SAFI_TYPE_L2VPN_FLOWSPEC        AfiSafiType = "l2vpn-flowspec"
	AFI_SAFI_TYPE_IPV4_SRPOLICY         AfiSafiType = "ipv4-srpolicy"
	AFI_SAFI_TYPE_IPV6_SRPOLICY         AfiSafiType = "ipv6-srpolicy"
	AFI_SAFI_TYPE_OPAQUE                AfiSafiType = "opaque"
	AFI_SAFI_TYPE_LS                    AfiSafiType = "ls"
)

// struct for container bgp-mp:afi-safi.
// AFI,SAFI configuration available for the
// neighbour or group.
type AfiSafi struct {
	// original -> bgp-mp:afi-safi-name
	// original -> bgp-mp:mp-graceful-restart
	// Parameters relating to BGP graceful-restart.
	MpGracefulRestart MpGracefulRestart `mapstructure:"mp-graceful-restart" json:"mp-graceful-restart,omitempty"`
	// original -> bgp-mp:afi-safi-config
	// Configuration parameters for the AFI-SAFI.
	Config AfiSafiConfig `mapstructure:"config" json:"config,omitempty"`
	// original -> bgp-mp:afi-safi-state
	// State information relating to the AFI-SAFI.
	State AfiSafiState `mapstructure:"state" json:"state,omitempty"`
	// original -> rpol:apply-policy
	// Anchor point for routing policies in the model.
	// Import and export policies are with respect to the local
	// routing table, i.e., export (send) and import (receive),
	// depending on the context.
	ApplyPolicy ApplyPolicy `mapstructure:"apply-policy" json:"apply-policy,omitempty"`
	// original -> bgp-mp:ipv4-unicast
	// IPv4 unicast configuration options.
	Ipv4Unicast Ipv4Unicast `mapstructure:"ipv4-unicast" json:"ipv4-unicast,omitempty"`
	// original -> bgp-mp:ipv6-unicast
	// IPv6 unicast configuration options.
	Ipv6Unicast Ipv6Unicast `mapstructure:"ipv6-unicast" json:"ipv6-unicast,omitempty"`
	// original -> bgp-mp:ipv4-labelled-unicast
	// IPv4 Labelled Unicast configuration options.
	Ipv4LabelledUnicast Ipv4LabelledUnicast `mapstructure:"ipv4-labelled-unicast" json:"ipv4-labelled-unicast,omitempty"`
	// original -> bgp-mp:ipv6-labelled-unicast
	// IPv6 Labelled Unicast configuration options.
	Ipv6LabelledUnicast Ipv6LabelledUnicast `mapstructure:"ipv6-labelled-unicast" json:"ipv6-labelled-unicast,omitempty"`
	// original -> bgp-mp:l3vpn-ipv4-unicast
	// Unicast IPv4 L3VPN configuration options.
	L3vpnIpv4Unicast L3vpnIpv4Unicast `mapstructure:"l3vpn-ipv4-unicast" json:"l3vpn-ipv4-unicast,omitempty"`
	// original -> bgp-mp:l3vpn-ipv6-unicast
	// Unicast IPv6 L3VPN configuration options.
	L3vpnIpv6Unicast L3vpnIpv6Unicast `mapstructure:"l3vpn-ipv6-unicast" json:"l3vpn-ipv6-unicast,omitempty"`
	// original -> bgp-mp:l3vpn-ipv4-multicast
	// Multicast IPv4 L3VPN configuration options.
	L3vpnIpv4Multicast L3vpnIpv4Multicast `mapstructure:"l3vpn-ipv4-multicast" json:"l3vpn-ipv4-multicast,omitempty"`
	// original -> bgp-mp:l3vpn-ipv6-multicast
	// Multicast IPv6 L3VPN configuration options.
	L3vpnIpv6Multicast L3vpnIpv6Multicast `mapstructure:"l3vpn-ipv6-multicast" json:"l3vpn-ipv6-multicast,omitempty"`
	// original -> bgp-mp:l2vpn-vpls
	// BGP-signalled VPLS configuration options.
	L2vpnVpls L2vpnVpls `mapstructure:"l2vpn-vpls" json:"l2vpn-vpls,omitempty"`
	// original -> bgp-mp:l2vpn-evpn
	// BGP EVPN configuration options.
	L2vpnEvpn L2vpnEvpn `mapstructure:"l2vpn-evpn" json:"l2vpn-evpn,omitempty"`
	// original -> bgp-mp:route-selection-options
	// Parameters relating to options for route selection.
	RouteSelectionOptions RouteSelectionOptions `mapstructure:"route-selection-options" json:"route-selection-options,omitempty"`
	// original -> bgp-mp:use-multiple-paths
	// Parameters related to the use of multiple paths for the
	// same NLRI.
	UseMultiplePaths UseMultiplePaths `mapstructure:"use-multiple-paths" json:"use-multiple-paths,omitempty"`
	// original -> bgp-mp:prefix-limit
	// Configure the maximum number of prefixes that will be
	// accepted from a peer.
	PrefixLimit PrefixLimit `mapstructure:"prefix-limit" json:"prefix-limit,omitempty"`
	// original -> gobgp:route-target-membership
	RouteTargetMembership RouteTargetMembership `mapstructure:"route-target-membership" json:"route-target-membership,omitempty"`
	// original -> gobgp:long-lived-graceful-restart
	LongLivedGracefulRestart LongLivedGracefulRestart `mapstructure:"long-lived-graceful-restart" json:"long-lived-graceful-restart,omitempty"`
	// original -> gobgp:add-paths
	// add-paths configuration options related to a particular AFI-SAFI.
	AddPaths AddPaths `mapstructure:"add-paths" json:"add-paths,omitempty"`
}

// struct for container bgp-mp:state.
// State information for BGP graceful-restart.
type MpGracefulRestartState struct {
	// original -> bgp-mp:enabled
	// bgp-mp:enabled's original type is boolean.
	// This leaf indicates whether graceful-restart is enabled for
	// this AFI-SAFI.
	Enabled bool `mapstructure:"enabled" json:"enabled,omitempty"`
	// original -> bgp-op:received
	// bgp-op:received's original type is boolean.
	// This leaf indicates whether the neighbor advertised the
	// ability to support graceful-restart for this AFI-SAFI.
	Received bool `mapstructure:"received" json:"received,omitempty"`
	// original -> bgp-op:advertised
	// bgp-op:advertised's original type is boolean.
	// This leaf indicates whether the ability to support
	// graceful-restart has been advertised to the peer.
	Advertised bool `mapstructure:"advertised" json:"advertised,omitempty"`
	// original -> gobgp:end-of-rib-received
	// gobgp:end-of-rib-received's original type is boolean.
	EndOfRibReceived bool `mapstructure:"end-of-rib-received" json:"end-of-rib-received,omitempty"`
	// original -> gobgp:end-of-rib-sent
	// gobgp:end-of-rib-sent's original type is boolean.
	EndOfRibSent bool `mapstructure:"end-of-rib-sent" json:"end-of-rib-sent,omitempty"`
}

// struct for container bgp-mp:config.
// Configuration options for BGP graceful-restart.
type MpGracefulRestartConfig struct {
	// original -> bgp-mp:enabled
	// bgp-mp:enabled's original type is boolean.
	// This leaf indicates whether graceful-restart is enabled for
	// this AFI-SAFI.
	Enabled bool `mapstructure:"enabled" json:"enabled,omitempty"`
}

func (lhs *MpGracefulRestartConfig) Equal(rhs *MpGracefulRestartConfig) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if lhs.Enabled != rhs.Enabled {
		return false
	}
	return true
}

// struct for container bgp-mp:graceful-restart.
// Parameters relating to BGP graceful-restart.
type MpGracefulRestart struct {
	// original -> bgp-mp:mp-graceful-restart-config
	// Configuration options for BGP graceful-restart.
	Config MpGracefulRestartConfig `mapstructure:"config" json:"config,omitempty"`
	// original -> bgp-mp:mp-graceful-restart-state
	// State information for BGP graceful-restart.
	State MpGracefulRestartState `mapstructure:"state" json:"state,omitempty"`
}

func (lhs *MpGracefulRestart) Equal(rhs *MpGracefulRestart) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if !lhs.Config.Equal(&(rhs.Config)) {
		return false
	}
	return true
}

// struct for container bgp-mp:ipv4-unicast.
// IPv4 unicast configuration options.
type Ipv4Unicast struct {
	// original -> bgp-mp:prefix-limit
	// Configure the maximum number of prefixes that will be
	// accepted from a peer.
	PrefixLimit PrefixLimit `mapstructure:"prefix-limit" json:"prefix-limit,omitempty"`
	// original -> bgp-mp:ipv4-unicast-config
	// Configuration parameters for common IPv4 and IPv6 unicast
	// AFI-SAFI options.
	Config Ipv4UnicastConfig `mapstructure:"config" json:"config,omitempty"`
	// original -> bgp-mp:ipv4-unicast-state
	// State information for common IPv4 and IPv6 unicast
	// parameters.
	State Ipv4UnicastState `mapstructure:"state" json:"state,omitempty"`
}

func (lhs *Ipv4Unicast) Equal(rhs *Ipv4Unicast) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if !lhs.PrefixLimit.Equal(&(rhs.PrefixLimit)) {
		return false
	}
	if !lhs.Config.Equal(&(rhs.Config)) {
		return false
	}
	return true
}

// struct for container rpol:state.
// Operational state for routing policy.
type ApplyPolicyState struct {
	// original -> rpol:import-policy
	// list of policy names in sequence to be applied on
	// receiving a routing update in the current context, e.g.,
	// for the current peer group, neighbor, address family,
	// etc.
	ImportPolicyList []string `mapstructure:"import-policy-list" json:"import-policy-list,omitempty"`
	// original -> rpol:default-import-policy
	// explicitly set a default policy if no policy definition
	// in the import policy chain is satisfied.
	DefaultImportPolicy DefaultPolicyType `mapstructure:"default-import-policy" json:"default-import-policy,omitempty"`
	// original -> rpol:export-policy
	// list of policy names in sequence to be applied on
	// sending a routing update in the current context, e.g.,
	// for the current peer group, neighbor, address family,
	// etc.
	ExportPolicyList []string `mapstructure:"export-policy-list" json:"export-policy-list,omitempty"`
	// original -> rpol:default-export-policy
	// explicitly set a default policy if no policy definition
	// in the export policy chain is satisfied.
	DefaultExportPolicy DefaultPolicyType `mapstructure:"default-export-policy" json:"default-export-policy,omitempty"`
	// original -> gobgp:in-policy
	// list of policy names in sequence to be applied on
	// sending a routing update in the current context, e.g.,
	// for the current other route server clients.
	InPolicyList []string `mapstructure:"in-policy-list" json:"in-policy-list,omitempty"`
	// original -> gobgp:default-in-policy
	// explicitly set a default policy if no policy definition
	// in the in-policy chain is satisfied.
	DefaultInPolicy DefaultPolicyType `mapstructure:"default-in-policy" json:"default-in-policy,omitempty"`
}

// struct for container rpol:config.
// Policy configuration data.
type ApplyPolicyConfig struct {
	// original -> rpol:import-policy
	// list of policy names in sequence to be applied on
	// receiving a routing update in the current context, e.g.,
	// for the current peer group, neighbor, address family,
	// etc.
	ImportPolicyList []string `mapstructure:"import-policy-list" json:"import-policy-list,omitempty"`
	// original -> rpol:default-import-policy
	// explicitly set a default policy if no policy definition
	// in the import policy chain is satisfied.
	DefaultImportPolicy DefaultPolicyType `mapstructure:"default-import-policy" json:"default-import-policy,omitempty"`
	// original -> rpol:export-policy
	// list of policy names in sequence to be applied on
	// sending a routing update in the current context, e.g.,
	// for the current peer group, neighbor, address family,
	// etc.
	ExportPolicyList []string `mapstructure:"export-policy-list" json:"export-policy-list,omitempty"`
	// original -> rpol:default-export-policy
	// explicitly set a default policy if no policy definition
	// in the export policy chain is satisfied.
	DefaultExportPolicy DefaultPolicyType `mapstructure:"default-export-policy" json:"default-export-policy,omitempty"`
	// original -> gobgp:in-policy
	// list of policy names in sequence to be applied on
	// sending a routing update in the current context, e.g.,
	// for the current other route server clients.
	InPolicyList []string `mapstructure:"in-policy-list" json:"in-policy-list,omitempty"`
	// original -> gobgp:default-in-policy
	// explicitly set a default policy if no policy definition
	// in the in-policy chain is satisfied.
	DefaultInPolicy DefaultPolicyType `mapstructure:"default-in-policy" json:"default-in-policy,omitempty"`
}

func (lhs *ApplyPolicyConfig) Equal(rhs *ApplyPolicyConfig) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if len(lhs.ImportPolicyList) != len(rhs.ImportPolicyList) {
		return false
	}
	for idx, l := range lhs.ImportPolicyList {
		if l != rhs.ImportPolicyList[idx] {
			return false
		}
	}
	if lhs.DefaultImportPolicy != rhs.DefaultImportPolicy {
		return false
	}
	if len(lhs.ExportPolicyList) != len(rhs.ExportPolicyList) {
		return false
	}
	for idx, l := range lhs.ExportPolicyList {
		if l != rhs.ExportPolicyList[idx] {
			return false
		}
	}
	if lhs.DefaultExportPolicy != rhs.DefaultExportPolicy {
		return false
	}
	if len(lhs.InPolicyList) != len(rhs.InPolicyList) {
		return false
	}
	for idx, l := range lhs.InPolicyList {
		if l != rhs.InPolicyList[idx] {
			return false
		}
	}
	if lhs.DefaultInPolicy != rhs.DefaultInPolicy {
		return false
	}
	return true
}

// struct for container rpol:apply-policy.
// Anchor point for routing policies in the model.
// Import and export policies are with respect to the local
// routing table, i.e., export (send) and import (receive),
// depending on the context.
type ApplyPolicy struct {
	// original -> rpol:apply-policy-config
	// Policy configuration data.
	Config ApplyPolicyConfig `mapstructure:"config" json:"config,omitempty"`
	// original -> rpol:apply-policy-state
	// Operational state for routing policy.
	State ApplyPolicyState `mapstructure:"state" json:"state,omitempty"`
}

func (lhs *ApplyPolicy) Equal(rhs *ApplyPolicy) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if !lhs.Config.Equal(&(rhs.Config)) {
		return false
	}
	return true
}

// struct for container bgp-mp:state.
// State information relating to the AFI-SAFI.
type AfiSafiState struct {
	// original -> bgp-mp:afi-safi-name
	// AFI,SAFI.
	AfiSafiName AfiSafiType `mapstructure:"afi-safi-name" json:"afi-safi-name,omitempty"`
	// original -> bgp-mp:enabled
	// bgp-mp:enabled's original type is boolean.
	// This leaf indicates whether the IPv4 Unicast AFI,SAFI is
	// enabled for the neighbour or group.
	Enabled bool `mapstructure:"enabled" json:"enabled,omitempty"`
	// original -> bgp-op:total-paths
	// Total number of BGP paths within the context.
	TotalPaths uint32 `mapstructure:"total-paths" json:"total-paths,omitempty"`
	// original -> bgp-op:total-prefixes
	// .
	TotalPrefixes uint32 `mapstructure:"total-prefixes" json:"total-prefixes,omitempty"`
	// original -> gobgp:family
	// gobgp:family's original type is route-family.
	// Address family value of AFI-SAFI pair translated from afi-safi-name.
	Family bgp.RouteFamily `mapstructure:"family" json:"family,omitempty"`
}

// struct for container bgp-mp:config.
// Configuration parameters for the AFI-SAFI.
type AfiSafiConfig struct {
	// original -> bgp-mp:afi-safi-name
	// AFI,SAFI.
	AfiSafiName AfiSafiType `mapstructure:"afi-safi-name" json:"afi-safi-name,omitempty"`
	// original -> bgp-mp:enabled
	// bgp-mp:enabled's original type is boolean.
	// This leaf indicates whether the IPv4 Unicast AFI,SAFI is
	// enabled for the neighbour or group.
	Enabled bool `mapstructure:"enabled" json:"enabled,omitempty"`
}

func (lhs *AfiSafiConfig) Equal(rhs *AfiSafiConfig) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if lhs.AfiSafiName != rhs.AfiSafiName {
		return false
	}
	if lhs.Enabled != rhs.Enabled {
		return false
	}
	return true
}

// struct for container bgp-mp:l2vpn-evpn.
// BGP EVPN configuration options.
type L2vpnEvpn struct {
	// original -> bgp-mp:prefix-limit
	// Configure the maximum number of prefixes that will be
	// accepted from a peer.
	PrefixLimit PrefixLimit `mapstructure:"prefix-limit" json:"prefix-limit,omitempty"`
}

func (lhs *L2vpnEvpn) Equal(rhs *L2vpnEvpn) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if !lhs.PrefixLimit.Equal(&(rhs.PrefixLimit)) {
		return false
	}
	return true
}

// struct for container bgp-mp:l2vpn-vpls.
// BGP-signalled VPLS configuration options.
type L2vpnVpls struct {
	// original -> bgp-mp:prefix-limit
	// Configure the maximum number of prefixes that will be
	// accepted from a peer.
	PrefixLimit PrefixLimit `mapstructure:"prefix-limit" json:"prefix-limit,omitempty"`
}

func (lhs *L2vpnVpls) Equal(rhs *L2vpnVpls) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if !lhs.PrefixLimit.Equal(&(rhs.PrefixLimit)) {
		return false
	}
	return true
}

// struct for container bgp-mp:l3vpn-ipv6-multicast.
// Multicast IPv6 L3VPN configuration options.
type L3vpnIpv6Multicast struct {
	// original -> bgp-mp:prefix-limit
	// Configure the maximum number of prefixes that will be
	// accepted from a peer.
	PrefixLimit PrefixLimit `mapstructure:"prefix-limit" json:"prefix-limit,omitempty"`
}

func (lhs *L3vpnIpv6Multicast) Equal(rhs *L3vpnIpv6Multicast) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if !lhs.PrefixLimit.Equal(&(rhs.PrefixLimit)) {
		return false
	}
	return true
}

// struct for container bgp-mp:l3vpn-ipv4-multicast.
// Multicast IPv4 L3VPN configuration options.
type L3vpnIpv4Multicast struct {
	// original -> bgp-mp:prefix-limit
	// Configure the maximum number of prefixes that will be
	// accepted from a peer.
	PrefixLimit PrefixLimit `mapstructure:"prefix-limit" json:"prefix-limit,omitempty"`
}

func (lhs *L3vpnIpv4Multicast) Equal(rhs *L3vpnIpv4Multicast) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if !lhs.PrefixLimit.Equal(&(rhs.PrefixLimit)) {
		return false
	}
	return true
}

// struct for container bgp-mp:l3vpn-ipv6-unicast.
// Unicast IPv6 L3VPN configuration options.
type L3vpnIpv6Unicast struct {
	// original -> bgp-mp:prefix-limit
	// Configure the maximum number of prefixes that will be
	// accepted from a peer.
	PrefixLimit PrefixLimit `mapstructure:"prefix-limit" json:"prefix-limit,omitempty"`
}

func (lhs *L3vpnIpv6Unicast) Equal(rhs *L3vpnIpv6Unicast) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if !lhs.PrefixLimit.Equal(&(rhs.PrefixLimit)) {
		return false
	}
	return true
}

// struct for container bgp-mp:l3vpn-ipv4-unicast.
// Unicast IPv4 L3VPN configuration options.
type L3vpnIpv4Unicast struct {
	// original -> bgp-mp:prefix-limit
	// Configure the maximum number of prefixes that will be
	// accepted from a peer.
	PrefixLimit PrefixLimit `mapstructure:"prefix-limit" json:"prefix-limit,omitempty"`
}

func (lhs *L3vpnIpv4Unicast) Equal(rhs *L3vpnIpv4Unicast) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if !lhs.PrefixLimit.Equal(&(rhs.PrefixLimit)) {
		return false
	}
	return true
}

// struct for container bgp-mp:ipv6-labelled-unicast.
// IPv6 Labelled Unicast configuration options.
type Ipv6LabelledUnicast struct {
	// original -> bgp-mp:prefix-limit
	// Configure the maximum number of prefixes that will be
	// accepted from a peer.
	PrefixLimit PrefixLimit `mapstructure:"prefix-limit" json:"prefix-limit,omitempty"`
}

func (lhs *Ipv6LabelledUnicast) Equal(rhs *Ipv6LabelledUnicast) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if !lhs.PrefixLimit.Equal(&(rhs.PrefixLimit)) {
		return false
	}
	return true
}

// struct for container bgp-mp:ipv4-labelled-unicast.
// IPv4 Labelled Unicast configuration options.
type Ipv4LabelledUnicast struct {
	// original -> bgp-mp:prefix-limit
	// Configure the maximum number of prefixes that will be
	// accepted from a peer.
	PrefixLimit PrefixLimit `mapstructure:"prefix-limit" json:"prefix-limit,omitempty"`
}

func (lhs *Ipv4LabelledUnicast) Equal(rhs *Ipv4LabelledUnicast) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if !lhs.PrefixLimit.Equal(&(rhs.PrefixLimit)) {
		return false
	}
	return true
}

// struct for container bgp-mp:state.
// State information for common IPv4 and IPv6 unicast
// parameters.
type Ipv6UnicastState struct {
	// original -> bgp-mp:send-default-route
	// bgp-mp:send-default-route's original type is boolean.
	// If set to true, send the default-route to the neighbour(s).
	SendDefaultRoute bool `mapstructure:"send-default-route" json:"send-default-route,omitempty"`
}

// struct for container bgp-mp:config.
// Configuration parameters for common IPv4 and IPv6 unicast
// AFI-SAFI options.
type Ipv6UnicastConfig struct {
	// original -> bgp-mp:send-default-route
	// bgp-mp:send-default-route's original type is boolean.
	// If set to true, send the default-route to the neighbour(s).
	SendDefaultRoute bool `mapstructure:"send-default-route" json:"send-default-route,omitempty"`
}

func (lhs *Ipv6UnicastConfig) Equal(rhs *Ipv6UnicastConfig) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if lhs.SendDefaultRoute != rhs.SendDefaultRoute {
		return false
	}
	return true
}

// struct for container bgp-mp:ipv6-unicast.
// IPv6 unicast configuration options.
type Ipv6Unicast struct {
	// original -> bgp-mp:prefix-limit
	// Configure the maximum number of prefixes that will be
	// accepted from a peer.
	PrefixLimit PrefixLimit `mapstructure:"prefix-limit" json:"prefix-limit,omitempty"`
	// original -> bgp-mp:ipv6-unicast-config
	// Configuration parameters for common IPv4 and IPv6 unicast
	// AFI-SAFI options.
	Config Ipv6UnicastConfig `mapstructure:"config" json:"config,omitempty"`
	// original -> bgp-mp:ipv6-unicast-state
	// State information for common IPv4 and IPv6 unicast
	// parameters.
	State Ipv6UnicastState `mapstructure:"state" json:"state,omitempty"`
}

func (lhs *Ipv6Unicast) Equal(rhs *Ipv6Unicast) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if !lhs.PrefixLimit.Equal(&(rhs.PrefixLimit)) {
		return false
	}
	if !lhs.Config.Equal(&(rhs.Config)) {
		return false
	}
	return true
}

// struct for container bgp-mp:state.
// State information for common IPv4 and IPv6 unicast
// parameters.
type Ipv4UnicastState struct {
	// original -> bgp-mp:send-default-route
	// bgp-mp:send-default-route's original type is boolean.
	// If set to true, send the default-route to the neighbour(s).
	SendDefaultRoute bool `mapstructure:"send-default-route" json:"send-default-route,omitempty"`
}

// struct for container bgp-mp:config.
// Configuration parameters for common IPv4 and IPv6 unicast
// AFI-SAFI options.
type Ipv4UnicastConfig struct {
	// original -> bgp-mp:send-default-route
	// bgp-mp:send-default-route's original type is boolean.
	// If set to true, send the default-route to the neighbour(s).
	SendDefaultRoute bool `mapstructure:"send-default-route" json:"send-default-route,omitempty"`
}

func (lhs *Ipv4UnicastConfig) Equal(rhs *Ipv4UnicastConfig) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if lhs.SendDefaultRoute != rhs.SendDefaultRoute {
		return false
	}
	return true
}

// struct for container bgp-mp:state.
// State information relating to the prefix-limit for the
// AFI-SAFI.
type PrefixLimitState struct {
	// original -> bgp-mp:max-prefixes
	// Maximum number of prefixes that will be accepted
	// from the neighbour.
	MaxPrefixes uint32 `mapstructure:"max-prefixes" json:"max-prefixes,omitempty"`
	// original -> bgp-mp:shutdown-threshold-pct
	// Threshold on number of prefixes that can be received
	// from a neighbour before generation of warning messages
	// or log entries. Expressed as a percentage of
	// max-prefixes.
	ShutdownThresholdPct Percentage `mapstructure:"shutdown-threshold-pct" json:"shutdown-threshold-pct,omitempty"`
	// original -> bgp-mp:restart-timer
	// bgp-mp:restart-timer's original type is decimal64.
	// Time interval in seconds after which the BGP session
	// is re-established after being torn down due to exceeding
	// the max-prefix limit.
	RestartTimer float64 `mapstructure:"restart-timer" json:"restart-timer,omitempty"`
}

// struct for container bgp-mp:config.
// Configuration parameters relating to the prefix
// limit for the AFI-SAFI.
type PrefixLimitConfig struct {
	// original -> bgp-mp:max-prefixes
	// Maximum number of prefixes that will be accepted
	// from the neighbour.
	MaxPrefixes uint32 `mapstructure:"max-prefixes" json:"max-prefixes,omitempty"`
	// original -> bgp-mp:shutdown-threshold-pct
	// Threshold on number of prefixes that can be received
	// from a neighbour before generation of warning messages
	// or log entries. Expressed as a percentage of
	// max-prefixes.
	ShutdownThresholdPct Percentage `mapstructure:"shutdown-threshold-pct" json:"shutdown-threshold-pct,omitempty"`
	// original -> bgp-mp:restart-timer
	// bgp-mp:restart-timer's original type is decimal64.
	// Time interval in seconds after which the BGP session
	// is re-established after being torn down due to exceeding
	// the max-prefix limit.
	RestartTimer float64 `mapstructure:"restart-timer" json:"restart-timer,omitempty"`
}

func (lhs *PrefixLimitConfig) Equal(rhs *PrefixLimitConfig) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if lhs.MaxPrefixes != rhs.MaxPrefixes {
		return false
	}
	if lhs.ShutdownThresholdPct != rhs.ShutdownThresholdPct {
		return false
	}
	if lhs.RestartTimer != rhs.RestartTimer {
		return false
	}
	return true
}

// struct for container bgp-mp:prefix-limit.
// Configure the maximum number of prefixes that will be
// accepted from a peer.
type PrefixLimit struct {
	// original -> bgp-mp:prefix-limit-config
	// Configuration parameters relating to the prefix
	// limit for the AFI-SAFI.
	Config PrefixLimitConfig `mapstructure:"config" json:"config,omitempty"`
	// original -> bgp-mp:prefix-limit-state
	// State information relating to the prefix-limit for the
	// AFI-SAFI.
	State PrefixLimitState `mapstructure:"state" json:"state,omitempty"`
}

func (lhs *PrefixLimit) Equal(rhs *PrefixLimit) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if !lhs.Config.Equal(&(rhs.Config)) {
		return false
	}
	return true
}

// struct for container gobgp:state.
type LongLivedGracefulRestartState struct {
	// original -> gobgp:enabled
	// gobgp:enabled's original type is boolean.
	Enabled bool `mapstructure:"enabled" json:"enabled,omitempty"`
	// original -> gobgp:received
	// gobgp:received's original type is boolean.
	Received bool `mapstructure:"received" json:"received,omitempty"`
	// original -> gobgp:advertised
	// gobgp:advertised's original type is boolean.
	Advertised bool `mapstructure:"advertised" json:"advertised,omitempty"`
	// original -> gobgp:peer-restart-time
	PeerRestartTime uint32 `mapstructure:"peer-restart-time" json:"peer-restart-time,omitempty"`
	// original -> gobgp:peer-restart-timer-expired
	// gobgp:peer-restart-timer-expired's original type is boolean.
	PeerRestartTimerExpired bool `mapstructure:"peer-restart-timer-expired" json:"peer-restart-timer-expired,omitempty"`
}

// struct for container gobgp:config.
type LongLivedGracefulRestartConfig struct {
	// original -> gobgp:enabled
	// gobgp:enabled's original type is boolean.
	Enabled bool `mapstructure:"enabled" json:"enabled,omitempty"`
	// original -> gobgp:restart-time
	RestartTime uint32 `mapstructure:"restart-time" json:"restart-time,omitempty"`
}

func (lhs *LongLivedGracefulRestartConfig) Equal(rhs *LongLivedGracefulRestartConfig) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if lhs.Enabled != rhs.Enabled {
		return false
	}
	if lhs.RestartTime != rhs.RestartTime {
		return false
	}
	return true
}

// struct for container gobgp:long-lived-graceful-restart.
type LongLivedGracefulRestart struct {
	// original -> gobgp:long-lived-graceful-restart-config
	Config LongLivedGracefulRestartConfig `mapstructure:"config" json:"config,omitempty"`
	// original -> gobgp:long-lived-graceful-restart-state
	State LongLivedGracefulRestartState `mapstructure:"state" json:"state,omitempty"`
}

func (lhs *LongLivedGracefulRestart) Equal(rhs *LongLivedGracefulRestart) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if !lhs.Config.Equal(&(rhs.Config)) {
		return false
	}
	return true
}

// struct for container gobgp:state.
type RouteTargetMembershipState struct {
	// original -> gobgp:deferral-time
	DeferralTime uint16 `mapstructure:"deferral-time" json:"deferral-time,omitempty"`
}

// struct for container gobgp:config.
type RouteTargetMembershipConfig struct {
	// original -> gobgp:deferral-time
	DeferralTime uint16 `mapstructure:"deferral-time" json:"deferral-time,omitempty"`
}

func (lhs *RouteTargetMembershipConfig) Equal(rhs *RouteTargetMembershipConfig) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if lhs.DeferralTime != rhs.DeferralTime {
		return false
	}
	return true
}

// struct for container gobgp:route-target-membership.
type RouteTargetMembership struct {
	// original -> gobgp:route-target-membership-config
	Config RouteTargetMembershipConfig `mapstructure:"config" json:"config,omitempty"`
	// original -> gobgp:route-target-membership-state
	State RouteTargetMembershipState `mapstructure:"state" json:"state,omitempty"`
}

func (lhs *RouteTargetMembership) Equal(rhs *RouteTargetMembership) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if !lhs.Config.Equal(&(rhs.Config)) {
		return false
	}
	return true
}

// struct for container bgp-mp:route-selection-options.
// Parameters relating to options for route selection.
type RouteSelectionOptions struct {
	// original -> bgp-mp:route-selection-options-config
	// Configuration parameters relating to route selection
	// options.
	Config RouteSelectionOptionsConfig `mapstructure:"config" json:"config,omitempty"`
	// original -> bgp-mp:route-selection-options-state
	// State information for the route selection options.
	State RouteSelectionOptionsState `mapstructure:"state" json:"state,omitempty"`
}

func (lhs *RouteSelectionOptions) Equal(rhs *RouteSelectionOptions) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if !lhs.Config.Equal(&(rhs.Config)) {
		return false
	}
	return true
}

// struct for container bgp-mp:config.
// Configuration parameters relating to route selection
// options.
type RouteSelectionOptionsConfig struct {
	// original -> bgp-mp:always-compare-med
	// bgp-mp:always-compare-med's original type is boolean.
	// Compare multi-exit discriminator (MED) value from
	// different ASes when selecting the best route.  The
	// default behavior is to only compare MEDs for paths
	// received from the same AS.
	AlwaysCompareMed bool `mapstructure:"always-compare-med" json:"always-compare-med,omitempty"`
	// original -> bgp-mp:ignore-as-path-length
	// bgp-mp:ignore-as-path-length's original type is boolean.
	// Ignore the AS path length when selecting the best path.
	// The default is to use the AS path length and prefer paths
	// with shorter length.
	IgnoreAsPathLength bool `mapstructure:"ignore-as-path-length" json:"ignore-as-path-length,omitempty"`
	// original -> bgp-mp:external-compare-router-id
	// bgp-mp:external-compare-router-id's original type is boolean.
	// When comparing similar routes received from external
	// BGP peers, use the router-id as a criterion to select
	// the active path.
	ExternalCompareRouterId bool `mapstructure:"external-compare-router-id" json:"external-compare-router-id,omitempty"`
	// original -> bgp-mp:advertise-inactive-routes
	// bgp-mp:advertise-inactive-routes's original type is boolean.
	// Advertise inactive routes to external peers.  The
	// default is to only advertise active routes.
	AdvertiseInactiveRoutes bool `mapstructure:"advertise-inactive-routes" json:"advertise-inactive-routes,omitempty"`
	// original -> bgp-mp:enable-aigp
	// bgp-mp:enable-aigp's original type is boolean.
	// Flag to enable sending / receiving accumulated IGP
	// attribute in routing updates.
	EnableAigp bool `mapstructure:"enable-aigp" json:"enable-aigp,omitempty"`
	// original -> bgp-mp:ignore-next-hop-igp-metric
	// bgp-mp:ignore-next-hop-igp-metric's original type is boolean.
	// Ignore the IGP metric to the next-hop when calculating
	// BGP best-path. The default is to select the route for
	// which the metric to the next-hop is lowest.
	IgnoreNextHopIgpMetric bool `mapstructure:"ignore-next-hop-igp-metric" json:"ignore-next-hop-igp-metric,omitempty"`
	// original -> gobgp:disable-best-path-selection
	// gobgp:disable-best-path-selection's original type is boolean.
	// Disables best path selection process.
	DisableBestPathSelection bool `mapstructure:"disable-best-path-selection" json:"disable-best-path-selection,omitempty"`
}

func (lhs *RouteSelectionOptionsConfig) Equal(rhs *RouteSelectionOptionsConfig) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if lhs.AlwaysCompareMed != rhs.AlwaysCompareMed {
		return false
	}
	if lhs.IgnoreAsPathLength != rhs.IgnoreAsPathLength {
		return false
	}
	if lhs.ExternalCompareRouterId != rhs.ExternalCompareRouterId {
		return false
	}
	if lhs.AdvertiseInactiveRoutes != rhs.AdvertiseInactiveRoutes {
		return false
	}
	if lhs.EnableAigp != rhs.EnableAigp {
		return false
	}
	if lhs.IgnoreNextHopIgpMetric != rhs.IgnoreNextHopIgpMetric {
		return false
	}
	if lhs.DisableBestPathSelection != rhs.DisableBestPathSelection {
		return false
	}
	return true
}

// struct for container bgp-mp:state.
// State information for the route selection options.
type RouteSelectionOptionsState struct {
	// original -> bgp-mp:always-compare-med
	// bgp-mp:always-compare-med's original type is boolean.
	// Compare multi-exit discriminator (MED) value from
	// different ASes when selecting the best route.  The
	// default behavior is to only compare MEDs for paths
	// received from the same AS.
	AlwaysCompareMed bool `mapstructure:"always-compare-med" json:"always-compare-med,omitempty"`
	// original -> bgp-mp:ignore-as-path-length
	// bgp-mp:ignore-as-path-length's original type is boolean.
	// Ignore the AS path length when selecting the best path.
	// The default is to use the AS path length and prefer paths
	// with shorter length.
	IgnoreAsPathLength bool `mapstructure:"ignore-as-path-length" json:"ignore-as-path-length,omitempty"`
	// original -> bgp-mp:external-compare-router-id
	// bgp-mp:external-compare-router-id's original type is boolean.
	// When comparing similar routes received from external
	// BGP peers, use the router-id as a criterion to select
	// the active path.
	ExternalCompareRouterId bool `mapstructure:"external-compare-router-id" json:"external-compare-router-id,omitempty"`
	// original -> bgp-mp:advertise-inactive-routes
	// bgp-mp:advertise-inactive-routes's original type is boolean.
	// Advertise inactive routes to external peers.  The
	// default is to only advertise active routes.
	AdvertiseInactiveRoutes bool `mapstructure:"advertise-inactive-routes" json:"advertise-inactive-routes,omitempty"`
	// original -> bgp-mp:enable-aigp
	// bgp-mp:enable-aigp's original type is boolean.
	// Flag to enable sending / receiving accumulated IGP
	// attribute in routing updates.
	EnableAigp bool `mapstructure:"enable-aigp" json:"enable-aigp,omitempty"`
	// original -> bgp-mp:ignore-next-hop-igp-metric
	// bgp-mp:ignore-next-hop-igp-metric's original type is boolean.
	// Ignore the IGP metric to the next-hop when calculating
	// BGP best-path. The default is to select the route for
	// which the metric to the next-hop is lowest.
	IgnoreNextHopIgpMetric bool `mapstructure:"ignore-next-hop-igp-metric" json:"ignore-next-hop-igp-metric,omitempty"`
	// original -> gobgp:disable-best-path-selection
	// gobgp:disable-best-path-selection's original type is boolean.
	// Disables best path selection process.
	DisableBestPathSelection bool `mapstructure:"disable-best-path-selection" json:"disable-best-path-selection,omitempty"`
}

// struct for container bgp-mp:use-multiple-paths.
// Parameters related to the use of multiple paths for the
// same NLRI.
type UseMultiplePaths struct {
	// original -> bgp-mp:use-multiple-paths-config
	// Configuration parameters relating to multipath.
	Config UseMultiplePathsConfig `mapstructure:"config" json:"config,omitempty"`
	// original -> bgp-mp:use-multiple-paths-state
	// State parameters relating to multipath.
	State UseMultiplePathsState `mapstructure:"state" json:"state,omitempty"`
	// original -> bgp-mp:ebgp
	// Multipath parameters for eBGP.
	Ebgp Ebgp `mapstructure:"ebgp" json:"ebgp,omitempty"`
	// original -> bgp-mp:ibgp
	// Multipath parameters for iBGP.
	Ibgp Ibgp `mapstructure:"ibgp" json:"ibgp,omitempty"`
}

func (lhs *UseMultiplePaths) Equal(rhs *UseMultiplePaths) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if !lhs.Config.Equal(&(rhs.Config)) {
		return false
	}
	if !lhs.Ebgp.Equal(&(rhs.Ebgp)) {
		return false
	}
	if !lhs.Ibgp.Equal(&(rhs.Ibgp)) {
		return false
	}
	return true
}

// struct for container bgp-mp:ebgp.
// Multipath parameters for eBGP.
type Ebgp struct {
	// original -> bgp-mp:ebgp-config
	// Configuration parameters relating to eBGP multipath.
	Config EbgpConfig `mapstructure:"config" json:"config,omitempty"`
	// original -> bgp-mp:ebgp-state
	// State information relating to eBGP multipath.
	State EbgpState `mapstructure:"state" json:"state,omitempty"`
}

func (lhs *Ebgp) Equal(rhs *Ebgp) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if !lhs.Config.Equal(&(rhs.Config)) {
		return false
	}
	return true
}

// struct for container bgp-mp:state.
// State parameters relating to multipath.
type UseMultiplePathsState struct {
	// original -> bgp-mp:enabled
	// bgp-mp:enabled's original type is boolean.
	// Whether the use of multiple paths for the same NLRI is
	// enabled for the neighbor. This value is overridden by
	// any more specific configuration value.
	Enabled bool `mapstructure:"enabled" json:"enabled,omitempty"`
}

// struct for container bgp-mp:config.
// Configuration parameters relating to multipath.
type UseMultiplePathsConfig struct {
	// original -> bgp-mp:enabled
	// bgp-mp:enabled's original type is boolean.
	// Whether the use of multiple paths for the same NLRI is
	// enabled for the neighbor. This value is overridden by
	// any more specific configuration value.
	Enabled bool `mapstructure:"enabled" json:"enabled,omitempty"`
}

func (lhs *UseMultiplePathsConfig) Equal(rhs *UseMultiplePathsConfig) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if lhs.Enabled != rhs.Enabled {
		return false
	}
	return true
}

// struct for container bgp-mp:ibgp.
// Multipath parameters for iBGP.
type Ibgp struct {
	// original -> bgp-mp:ibgp-config
	// Configuration parameters relating to iBGP multipath.
	Config IbgpConfig `mapstructure:"config" json:"config,omitempty"`
	// original -> bgp-mp:ibgp-state
	// State information relating to iBGP multipath.
	State IbgpState `mapstructure:"state" json:"state,omitempty"`
}

func (lhs *Ibgp) Equal(rhs *Ibgp) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if !lhs.Config.Equal(&(rhs.Config)) {
		return false
	}
	return true
}

// struct for container bgp-mp:state.
// State information relating to iBGP multipath.
type IbgpState struct {
	// original -> bgp-mp:maximum-paths
	// Maximum number of parallel paths to consider when using
	// iBGP multipath. The default is to use a single path.
	MaximumPaths uint32 `mapstructure:"maximum-paths" json:"maximum-paths,omitempty"`
}

// struct for container bgp-mp:config.
// Configuration parameters relating to iBGP multipath.
type IbgpConfig struct {
	// original -> bgp-mp:maximum-paths
	// Maximum number of parallel paths to consider when using
	// iBGP multipath. The default is to use a single path.
	MaximumPaths uint32 `mapstructure:"maximum-paths" json:"maximum-paths,omitempty"`
}

func (lhs *IbgpConfig) Equal(rhs *IbgpConfig) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if lhs.MaximumPaths != rhs.MaximumPaths {
		return false
	}
	return true
}

// struct for container bgp:state.
// State information associated with ADD_PATHS.
type AddPathsState struct {
	// original -> bgp:receive
	// bgp:receive's original type is boolean.
	// Enable ability to receive multiple path advertisements
	// for an NLRI from the neighbor or group.
	Receive bool `mapstructure:"receive" json:"receive,omitempty"`
	// original -> bgp:send-max
	// The maximum number of paths to advertise to neighbors
	// for a single NLRI.
	SendMax uint8 `mapstructure:"send-max" json:"send-max,omitempty"`
}

// struct for container bgp:config.
// Configuration parameters relating to ADD_PATHS.
type AddPathsConfig struct {
	// original -> bgp:receive
	// bgp:receive's original type is boolean.
	// Enable ability to receive multiple path advertisements
	// for an NLRI from the neighbor or group.
	Receive bool `mapstructure:"receive" json:"receive,omitempty"`
	// original -> bgp:send-max
	// The maximum number of paths to advertise to neighbors
	// for a single NLRI.
	SendMax uint8 `mapstructure:"send-max" json:"send-max,omitempty"`
}

func (lhs *AddPathsConfig) Equal(rhs *AddPathsConfig) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if lhs.Receive != rhs.Receive {
		return false
	}
	if lhs.SendMax != rhs.SendMax {
		return false
	}
	return true
}

// struct for container bgp:add-paths.
// Parameters relating to the advertisement and receipt of
// multiple paths for a single NLRI (add-paths).
type AddPaths struct {
	// original -> bgp:add-paths-config
	// Configuration parameters relating to ADD_PATHS.
	Config AddPathsConfig `mapstructure:"config" json:"config,omitempty"`
	// original -> bgp:add-paths-state
	// State information associated with ADD_PATHS.
	State AddPathsState `mapstructure:"state" json:"state,omitempty"`
}

func (lhs *AddPaths) Equal(rhs *AddPaths) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if !lhs.Config.Equal(&(rhs.Config)) {
		return false
	}
	return true
}
