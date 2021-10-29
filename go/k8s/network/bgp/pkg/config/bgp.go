package config

import api "github.com/osrg/gobgp/api"

var AfiSafiTypeToIntMap = map[AfiSafiType]int{
	AFI_SAFI_TYPE_IPV4_UNICAST:          0,
	AFI_SAFI_TYPE_IPV6_UNICAST:          1,
	AFI_SAFI_TYPE_IPV4_LABELLED_UNICAST: 2,
	AFI_SAFI_TYPE_IPV6_LABELLED_UNICAST: 3,
	AFI_SAFI_TYPE_L3VPN_IPV4_UNICAST:    4,
	AFI_SAFI_TYPE_L3VPN_IPV6_UNICAST:    5,
	AFI_SAFI_TYPE_L3VPN_IPV4_MULTICAST:  6,
	AFI_SAFI_TYPE_L3VPN_IPV6_MULTICAST:  7,
	AFI_SAFI_TYPE_L2VPN_VPLS:            8,
	AFI_SAFI_TYPE_L2VPN_EVPN:            9,
	AFI_SAFI_TYPE_IPV4_MULTICAST:        10,
	AFI_SAFI_TYPE_IPV6_MULTICAST:        11,
	AFI_SAFI_TYPE_RTC:                   12,
	AFI_SAFI_TYPE_IPV4_ENCAP:            13,
	AFI_SAFI_TYPE_IPV6_ENCAP:            14,
	AFI_SAFI_TYPE_IPV4_FLOWSPEC:         15,
	AFI_SAFI_TYPE_L3VPN_IPV4_FLOWSPEC:   16,
	AFI_SAFI_TYPE_IPV6_FLOWSPEC:         17,
	AFI_SAFI_TYPE_L3VPN_IPV6_FLOWSPEC:   18,
	AFI_SAFI_TYPE_L2VPN_FLOWSPEC:        19,
	AFI_SAFI_TYPE_IPV4_SRPOLICY:         20,
	AFI_SAFI_TYPE_IPV6_SRPOLICY:         21,
	AFI_SAFI_TYPE_OPAQUE:                22,
	AFI_SAFI_TYPE_LS:                    23,
}

// struct for container bgp:config.
// Configuration parameters relating to the global BGP router.
type GlobalConfig struct {
	// original -> bgp:as
	// bgp:as's original type is inet:as-number.
	// Local autonomous system number of the router.  Uses
	// the 32-bit as-number type from the model in RFC 6991.
	As uint32 `mapstructure:"as" json:"as,omitempty"`
	// original -> bgp:router-id
	// bgp:router-id's original type is inet:ipv4-address.
	// Router id of the router, expressed as an
	// 32-bit value, IPv4 address.
	RouterId string `mapstructure:"router-id" json:"router-id,omitempty"`
	// original -> gobgp:port
	Port int32 `mapstructure:"port" json:"port,omitempty"`
	// original -> gobgp:local-address
	LocalAddressList []string `mapstructure:"local-address-list" json:"local-address-list,omitempty"`
}

// struct for container bgp:state.
// State information relating to the global BGP router.
type GlobalState struct {
	// original -> bgp:as
	// bgp:as's original type is inet:as-number.
	// Local autonomous system number of the router.  Uses
	// the 32-bit as-number type from the model in RFC 6991.
	As uint32 `mapstructure:"as" json:"as,omitempty"`
	// original -> bgp:router-id
	// bgp:router-id's original type is inet:ipv4-address.
	// Router id of the router, expressed as an
	// 32-bit value, IPv4 address.
	RouterId string `mapstructure:"router-id" json:"router-id,omitempty"`
	// original -> bgp-op:total-paths
	// Total number of BGP paths within the context.
	TotalPaths uint32 `mapstructure:"total-paths" json:"total-paths,omitempty"`
	// original -> bgp-op:total-prefixes
	// .
	TotalPrefixes uint32 `mapstructure:"total-prefixes" json:"total-prefixes,omitempty"`
	// original -> gobgp:port
	Port int32 `mapstructure:"port" json:"port,omitempty"`
	// original -> gobgp:local-address
	LocalAddressList []string `mapstructure:"local-address-list" json:"local-address-list,omitempty"`
}

// struct for container bgp:global.
// Global configuration for the BGP router.
type Global struct {
	// original -> bgp:global-config
	// Configuration parameters relating to the global BGP router.
	Config GlobalConfig `mapstructure:"config" json:"config,omitempty"`

	// original -> bgp:global-state
	// State information relating to the global BGP router.
	State GlobalState `mapstructure:"state" json:"state,omitempty"`
	// original -> bgp-mp:route-selection-options
	// Parameters relating to options for route selection.
	RouteSelectionOptions RouteSelectionOptions `mapstructure:"route-selection-options" json:"route-selection-options,omitempty"`
	// original -> bgp:default-route-distance
	// Administrative distance (or preference) assigned to
	// routes received from different sources
	// (external, internal, and local).
	//DefaultRouteDistance DefaultRouteDistance `mapstructure:"default-route-distance" json:"default-route-distance,omitempty"`
	// original -> bgp:confederation
	// Parameters indicating whether the local system acts as part
	// of a BGP confederation.
	//Confederation Confederation `mapstructure:"confederation" json:"confederation,omitempty"`
	// original -> bgp-mp:use-multiple-paths
	// Parameters related to the use of multiple paths for the
	// same NLRI.
	UseMultiplePaths UseMultiplePaths `mapstructure:"use-multiple-paths" json:"use-multiple-paths,omitempty"`
	// original -> bgp:graceful-restart
	// Parameters relating the graceful restart mechanism for BGP.
	GracefulRestart GracefulRestart `mapstructure:"graceful-restart" json:"graceful-restart,omitempty"`
	// original -> bgp:afi-safis
	// Address family specific configuration.
	AfiSafis []AfiSafi `mapstructure:"afi-safis" json:"afi-safis,omitempty"`
	// original -> rpol:apply-policy
	// Anchor point for routing policies in the model.
	// Import and export policies are with respect to the local
	// routing table, i.e., export (send) and import (receive),
	// depending on the context.
	ApplyPolicy ApplyPolicy `mapstructure:"apply-policy" json:"apply-policy,omitempty"`
}

func NewGlobalFromConfigStruct(c *Global) *api.Global {
	families := make([]uint32, 0, len(c.AfiSafis))
	for _, f := range c.AfiSafis {
		families = append(families, uint32(AfiSafiTypeToIntMap[f.Config.AfiSafiName]))
	}

	applyPolicy := newApplyPolicyFromConfigStruct(&c.ApplyPolicy)

	return &api.Global{
		As:               c.Config.As,
		RouterId:         c.Config.RouterId,
		ListenPort:       c.Config.Port,
		ListenAddresses:  c.Config.LocalAddressList,
		Families:         families,
		UseMultiplePaths: c.UseMultiplePaths.Config.Enabled,
		RouteSelectionOptions: &api.RouteSelectionOptionsConfig{
			AlwaysCompareMed:         c.RouteSelectionOptions.Config.AlwaysCompareMed,
			IgnoreAsPathLength:       c.RouteSelectionOptions.Config.IgnoreAsPathLength,
			ExternalCompareRouterId:  c.RouteSelectionOptions.Config.ExternalCompareRouterId,
			AdvertiseInactiveRoutes:  c.RouteSelectionOptions.Config.AdvertiseInactiveRoutes,
			EnableAigp:               c.RouteSelectionOptions.Config.EnableAigp,
			IgnoreNextHopIgpMetric:   c.RouteSelectionOptions.Config.IgnoreNextHopIgpMetric,
			DisableBestPathSelection: c.RouteSelectionOptions.Config.DisableBestPathSelection,
		},
		DefaultRouteDistance: &api.DefaultRouteDistance{
			//ExternalRouteDistance: uint32(c.DefaultRouteDistance.Config.ExternalRouteDistance),
			//InternalRouteDistance: uint32(c.DefaultRouteDistance.Config.InternalRouteDistance),
		},
		Confederation: &api.Confederation{
			//Enabled:      c.Confederation.Config.Enabled,
			//Identifier:   c.Confederation.Config.Identifier,
			//MemberAsList: c.Confederation.Config.MemberAsList,
		},
		GracefulRestart: &api.GracefulRestart{
			Enabled:             c.GracefulRestart.Config.Enabled,
			RestartTime:         uint32(c.GracefulRestart.Config.RestartTime),
			StaleRoutesTime:     uint32(c.GracefulRestart.Config.StaleRoutesTime),
			HelperOnly:          c.GracefulRestart.Config.HelperOnly,
			DeferralTime:        uint32(c.GracefulRestart.Config.DeferralTime),
			NotificationEnabled: c.GracefulRestart.Config.NotificationEnabled,
			LonglivedEnabled:    c.GracefulRestart.Config.LongLivedEnabled,
		},
		ApplyPolicy: applyPolicy,
	}
}

// struct for container bgp:bgp.
// Top-level configuration and state for the BGP router.
type Bgp struct {
	// original -> bgp:global
	// Global configuration for the BGP router.
	Global Global `mapstructure:"global" json:"global,omitempty"`
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
// State information relating to eBGP multipath.
type EbgpState struct {
	// original -> bgp-mp:allow-multiple-as
	// bgp-mp:allow-multiple-as's original type is boolean.
	// Allow multipath to use paths from different neighbouring
	// ASes.  The default is to only consider multiple paths from
	// the same neighbouring AS.
	AllowMultipleAs bool `mapstructure:"allow-multiple-as" json:"allow-multiple-as,omitempty"`
	// original -> bgp-mp:maximum-paths
	// Maximum number of parallel paths to consider when using
	// BGP multipath. The default is use a single path.
	MaximumPaths uint32 `mapstructure:"maximum-paths" json:"maximum-paths,omitempty"`
}

// struct for container bgp-mp:config.
// Configuration parameters relating to eBGP multipath.
type EbgpConfig struct {
	// original -> bgp-mp:allow-multiple-as
	// bgp-mp:allow-multiple-as's original type is boolean.
	// Allow multipath to use paths from different neighbouring
	// ASes.  The default is to only consider multiple paths from
	// the same neighbouring AS.
	AllowMultipleAs bool `mapstructure:"allow-multiple-as" json:"allow-multiple-as,omitempty"`
	// original -> bgp-mp:maximum-paths
	// Maximum number of parallel paths to consider when using
	// BGP multipath. The default is use a single path.
	MaximumPaths uint32 `mapstructure:"maximum-paths" json:"maximum-paths,omitempty"`
}

func (lhs *EbgpConfig) Equal(rhs *EbgpConfig) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if lhs.AllowMultipleAs != rhs.AllowMultipleAs {
		return false
	}
	if lhs.MaximumPaths != rhs.MaximumPaths {
		return false
	}
	return true
}
