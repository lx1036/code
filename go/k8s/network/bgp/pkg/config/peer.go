package config

// struct for container bgp:peer-group.
// List of BGP peer-groups configured on the local system -
// uniquely identified by peer-group name.
type PeerGroup struct {
	// original -> bgp:peer-group-name
	// original -> bgp:peer-group-config
	// Configuration parameters relating to the BGP neighbor or
	// group.
	Config PeerGroupConfig `mapstructure:"config" json:"config,omitempty"`
	// original -> bgp:peer-group-state
	// State information relating to the BGP neighbor or group.
	State PeerGroupState `mapstructure:"state" json:"state,omitempty"`
	// original -> bgp:timers
	// Timers related to a BGP neighbor or group.
	Timers Timers `mapstructure:"timers" json:"timers,omitempty"`
	// original -> bgp:transport
	// Transport session parameters for the BGP neighbor or group.
	Transport Transport `mapstructure:"transport" json:"transport,omitempty"`
	// original -> bgp:error-handling
	// Error handling parameters used for the BGP neighbor or
	// group.
	ErrorHandling ErrorHandling `mapstructure:"error-handling" json:"error-handling,omitempty"`
	// original -> bgp:logging-options
	// Logging options for events related to the BGP neighbor or
	// group.
	LoggingOptions LoggingOptions `mapstructure:"logging-options" json:"logging-options,omitempty"`
	// original -> bgp:ebgp-multihop
	// eBGP multi-hop parameters for the BGP neighbor or group.
	EbgpMultihop EbgpMultihop `mapstructure:"ebgp-multihop" json:"ebgp-multihop,omitempty"`
	// original -> bgp:route-reflector
	// Route reflector parameters for the BGP neighbor or group.
	RouteReflector RouteReflector `mapstructure:"route-reflector" json:"route-reflector,omitempty"`
	// original -> bgp:as-path-options
	// AS_PATH manipulation parameters for the BGP neighbor or
	// group.
	AsPathOptions AsPathOptions `mapstructure:"as-path-options" json:"as-path-options,omitempty"`
	// original -> bgp:add-paths
	// Parameters relating to the advertisement and receipt of
	// multiple paths for a single NLRI (add-paths).
	AddPaths AddPaths `mapstructure:"add-paths" json:"add-paths,omitempty"`
	// original -> bgp:afi-safis
	// Per-address-family configuration parameters associated with
	// the neighbor or group.
	AfiSafis []AfiSafi `mapstructure:"afi-safis" json:"afi-safis,omitempty"`
	// original -> bgp:graceful-restart
	// Parameters relating the graceful restart mechanism for BGP.
	GracefulRestart GracefulRestart `mapstructure:"graceful-restart" json:"graceful-restart,omitempty"`
	// original -> rpol:apply-policy
	// Anchor point for routing policies in the model.
	// Import and export policies are with respect to the local
	// routing table, i.e., export (send) and import (receive),
	// depending on the context.
	ApplyPolicy ApplyPolicy `mapstructure:"apply-policy" json:"apply-policy,omitempty"`
	// original -> bgp-mp:use-multiple-paths
	// Parameters related to the use of multiple paths for the
	// same NLRI.
	UseMultiplePaths UseMultiplePaths `mapstructure:"use-multiple-paths" json:"use-multiple-paths,omitempty"`
	// original -> gobgp:route-server
	// Configure the local router as a route server.
	RouteServer RouteServer `mapstructure:"route-server" json:"route-server,omitempty"`
	// original -> gobgp:ttl-security
	// Configure TTL Security feature.
	TtlSecurity TtlSecurity `mapstructure:"ttl-security" json:"ttl-security,omitempty"`
}

func (lhs *PeerGroup) Equal(rhs *PeerGroup) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if !lhs.Config.Equal(&(rhs.Config)) {
		return false
	}
	if !lhs.Timers.Equal(&(rhs.Timers)) {
		return false
	}
	if !lhs.Transport.Equal(&(rhs.Transport)) {
		return false
	}
	if !lhs.ErrorHandling.Equal(&(rhs.ErrorHandling)) {
		return false
	}
	if !lhs.LoggingOptions.Equal(&(rhs.LoggingOptions)) {
		return false
	}
	if !lhs.EbgpMultihop.Equal(&(rhs.EbgpMultihop)) {
		return false
	}
	if !lhs.RouteReflector.Equal(&(rhs.RouteReflector)) {
		return false
	}
	if !lhs.AsPathOptions.Equal(&(rhs.AsPathOptions)) {
		return false
	}
	if !lhs.AddPaths.Equal(&(rhs.AddPaths)) {
		return false
	}
	if len(lhs.AfiSafis) != len(rhs.AfiSafis) {
		return false
	}
	{
		lmap := make(map[string]*AfiSafi)
		for i, l := range lhs.AfiSafis {
			lmap[mapkey(i, string(l.Config.AfiSafiName))] = &lhs.AfiSafis[i]
		}
		for i, r := range rhs.AfiSafis {
			if l, y := lmap[mapkey(i, string(r.Config.AfiSafiName))]; !y {
				return false
			} else if !r.Equal(l) {
				return false
			}
		}
	}
	if !lhs.GracefulRestart.Equal(&(rhs.GracefulRestart)) {
		return false
	}
	if !lhs.ApplyPolicy.Equal(&(rhs.ApplyPolicy)) {
		return false
	}
	if !lhs.UseMultiplePaths.Equal(&(rhs.UseMultiplePaths)) {
		return false
	}
	if !lhs.RouteServer.Equal(&(rhs.RouteServer)) {
		return false
	}
	if !lhs.TtlSecurity.Equal(&(rhs.TtlSecurity)) {
		return false
	}
	return true
}

// struct for container bgp:state.
// State information relating to the BGP neighbor or group.
type PeerGroupState struct {
	// original -> bgp:peer-as
	// bgp:peer-as's original type is inet:as-number.
	// AS number of the peer.
	PeerAs uint32 `mapstructure:"peer-as" json:"peer-as,omitempty"`
	// original -> bgp:local-as
	// bgp:local-as's original type is inet:as-number.
	// The local autonomous system number that is to be used
	// when establishing sessions with the remote peer or peer
	// group, if this differs from the global BGP router
	// autonomous system number.
	LocalAs uint32 `mapstructure:"local-as" json:"local-as,omitempty"`
	// original -> bgp:peer-type
	// Explicitly designate the peer or peer group as internal
	// (iBGP) or external (eBGP).
	PeerType PeerType `mapstructure:"peer-type" json:"peer-type,omitempty"`
	// original -> bgp:auth-password
	// Configures an MD5 authentication password for use with
	// neighboring devices.
	AuthPassword string `mapstructure:"auth-password" json:"auth-password,omitempty"`
	// original -> bgp:remove-private-as
	// Remove private AS numbers from updates sent to peers.
	RemovePrivateAs RemovePrivateAsOption `mapstructure:"remove-private-as" json:"remove-private-as,omitempty"`
	// original -> bgp:route-flap-damping
	// bgp:route-flap-damping's original type is boolean.
	// Enable route flap damping.
	RouteFlapDamping bool `mapstructure:"route-flap-damping" json:"route-flap-damping,omitempty"`
	// original -> bgp:send-community
	// Specify which types of community should be sent to the
	// neighbor or group. The default is to not send the
	// community attribute.
	SendCommunity CommunityType `mapstructure:"send-community" json:"send-community,omitempty"`
	// original -> bgp:description
	// An optional textual description (intended primarily for use
	// with a peer or group.
	Description string `mapstructure:"description" json:"description,omitempty"`
	// original -> bgp:peer-group-name
	// Name of the BGP peer-group.
	PeerGroupName string `mapstructure:"peer-group-name" json:"peer-group-name,omitempty"`
	// original -> bgp-op:total-paths
	// Total number of BGP paths within the context.
	TotalPaths uint32 `mapstructure:"total-paths" json:"total-paths,omitempty"`
	// original -> bgp-op:total-prefixes
	// .
	TotalPrefixes uint32 `mapstructure:"total-prefixes" json:"total-prefixes,omitempty"`
}

// struct for container bgp:config.
// Configuration parameters relating to the BGP neighbor or
// group.
type PeerGroupConfig struct {
	// original -> bgp:peer-as
	// bgp:peer-as's original type is inet:as-number.
	// AS number of the peer.
	PeerAs uint32 `mapstructure:"peer-as" json:"peer-as,omitempty"`
	// original -> bgp:local-as
	// bgp:local-as's original type is inet:as-number.
	// The local autonomous system number that is to be used
	// when establishing sessions with the remote peer or peer
	// group, if this differs from the global BGP router
	// autonomous system number.
	LocalAs uint32 `mapstructure:"local-as" json:"local-as,omitempty"`
	// original -> bgp:peer-type
	// Explicitly designate the peer or peer group as internal
	// (iBGP) or external (eBGP).
	PeerType PeerType `mapstructure:"peer-type" json:"peer-type,omitempty"`
	// original -> bgp:auth-password
	// Configures an MD5 authentication password for use with
	// neighboring devices.
	AuthPassword string `mapstructure:"auth-password" json:"auth-password,omitempty"`
	// original -> bgp:remove-private-as
	// Remove private AS numbers from updates sent to peers.
	RemovePrivateAs RemovePrivateAsOption `mapstructure:"remove-private-as" json:"remove-private-as,omitempty"`
	// original -> bgp:route-flap-damping
	// bgp:route-flap-damping's original type is boolean.
	// Enable route flap damping.
	RouteFlapDamping bool `mapstructure:"route-flap-damping" json:"route-flap-damping,omitempty"`
	// original -> bgp:send-community
	// Specify which types of community should be sent to the
	// neighbor or group. The default is to not send the
	// community attribute.
	SendCommunity CommunityType `mapstructure:"send-community" json:"send-community,omitempty"`
	// original -> bgp:description
	// An optional textual description (intended primarily for use
	// with a peer or group.
	Description string `mapstructure:"description" json:"description,omitempty"`
	// original -> bgp:peer-group-name
	// Name of the BGP peer-group.
	PeerGroupName string `mapstructure:"peer-group-name" json:"peer-group-name,omitempty"`
}

func (lhs *PeerGroupConfig) Equal(rhs *PeerGroupConfig) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if lhs.PeerAs != rhs.PeerAs {
		return false
	}
	if lhs.LocalAs != rhs.LocalAs {
		return false
	}
	if lhs.PeerType != rhs.PeerType {
		return false
	}
	if lhs.AuthPassword != rhs.AuthPassword {
		return false
	}
	if lhs.RemovePrivateAs != rhs.RemovePrivateAs {
		return false
	}
	if lhs.RouteFlapDamping != rhs.RouteFlapDamping {
		return false
	}
	if lhs.SendCommunity != rhs.SendCommunity {
		return false
	}
	if lhs.Description != rhs.Description {
		return false
	}
	if lhs.PeerGroupName != rhs.PeerGroupName {
		return false
	}
	return true
}
