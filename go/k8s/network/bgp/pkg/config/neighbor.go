package config

import (
	"fmt"

	"github.com/osrg/gobgp/pkg/packet/bgp"
)

// typedef for typedef bgp-types:rr-cluster-id-type.
type RrClusterIdType string

// typedef for identity bgp-types:peer-type.
// labels a peer or peer group as explicitly internal or
// external.
type PeerType string

const (
	PEER_TYPE_INTERNAL PeerType = "internal"
	PEER_TYPE_EXTERNAL PeerType = "external"
)

// typedef for identity bgp-types:remove-private-as-option.
// set of options for configuring how private AS path numbers
// are removed from advertisements.
type RemovePrivateAsOption string

const (
	REMOVE_PRIVATE_AS_OPTION_ALL     RemovePrivateAsOption = "all"
	REMOVE_PRIVATE_AS_OPTION_REPLACE RemovePrivateAsOption = "replace"
)

// typedef for identity bgp-types:community-type.
// type describing variations of community attributes:
// STANDARD: standard BGP community [rfc1997]
// EXTENDED: extended BGP community [rfc4360]
// BOTH: both standard and extended community.
type CommunityType string

const (
	COMMUNITY_TYPE_STANDARD CommunityType = "standard"
	COMMUNITY_TYPE_EXTENDED CommunityType = "extended"
	COMMUNITY_TYPE_BOTH     CommunityType = "both"
	COMMUNITY_TYPE_NONE     CommunityType = "none"
)

// typedef for identity bgp:session-state.
// Operational state of the BGP peer.
type SessionState string

const (
	SESSION_STATE_IDLE        SessionState = "idle"
	SESSION_STATE_CONNECT     SessionState = "connect"
	SESSION_STATE_ACTIVE      SessionState = "active"
	SESSION_STATE_OPENSENT    SessionState = "opensent"
	SESSION_STATE_OPENCONFIRM SessionState = "openconfirm"
	SESSION_STATE_ESTABLISHED SessionState = "established"
)

var SessionStateToIntMap = map[SessionState]int{
	SESSION_STATE_IDLE:        0,
	SESSION_STATE_CONNECT:     1,
	SESSION_STATE_ACTIVE:      2,
	SESSION_STATE_OPENSENT:    3,
	SESSION_STATE_OPENCONFIRM: 4,
	SESSION_STATE_ESTABLISHED: 5,
}

var IntToSessionStateMap = map[int]SessionState{
	0: SESSION_STATE_IDLE,
	1: SESSION_STATE_CONNECT,
	2: SESSION_STATE_ACTIVE,
	3: SESSION_STATE_OPENSENT,
	4: SESSION_STATE_OPENCONFIRM,
	5: SESSION_STATE_ESTABLISHED,
}

func (v SessionState) Validate() error {
	if _, ok := SessionStateToIntMap[v]; !ok {
		return fmt.Errorf("invalid SessionState: %s", v)
	}
	return nil
}

func (v SessionState) ToInt() int {
	i, ok := SessionStateToIntMap[v]
	if !ok {
		return -1
	}
	return i
}

// typedef for identity bgp-types:bgp-capability.
// Base identity for a BGP capability.
type BgpCapability string

const (
	BGP_CAPABILITY_MPBGP            BgpCapability = "mpbgp"
	BGP_CAPABILITY_ROUTE_REFRESH    BgpCapability = "route-refresh"
	BGP_CAPABILITY_ASN32            BgpCapability = "asn32"
	BGP_CAPABILITY_GRACEFUL_RESTART BgpCapability = "graceful-restart"
	BGP_CAPABILITY_ADD_PATHS        BgpCapability = "add-paths"
)

var BgpCapabilityToIntMap = map[BgpCapability]int{
	BGP_CAPABILITY_MPBGP:            0,
	BGP_CAPABILITY_ROUTE_REFRESH:    1,
	BGP_CAPABILITY_ASN32:            2,
	BGP_CAPABILITY_GRACEFUL_RESTART: 3,
	BGP_CAPABILITY_ADD_PATHS:        4,
}

var IntToBgpCapabilityMap = map[int]BgpCapability{
	0: BGP_CAPABILITY_MPBGP,
	1: BGP_CAPABILITY_ROUTE_REFRESH,
	2: BGP_CAPABILITY_ASN32,
	3: BGP_CAPABILITY_GRACEFUL_RESTART,
	4: BGP_CAPABILITY_ADD_PATHS,
}

func (v BgpCapability) Validate() error {
	if _, ok := BgpCapabilityToIntMap[v]; !ok {
		return fmt.Errorf("invalid BgpCapability: %s", v)
	}
	return nil
}

func (v BgpCapability) ToInt() int {
	i, ok := BgpCapabilityToIntMap[v]
	if !ok {
		return -1
	}
	return i
}

// typedef for identity bgp:mode.
// Ths leaf indicates the mode of operation of BGP graceful
// restart with the peer.
type Mode string

const (
	MODE_HELPER_ONLY   Mode = "helper-only"
	MODE_BILATERAL     Mode = "bilateral"
	MODE_REMOTE_HELPER Mode = "remote-helper"
)

var ModeToIntMap = map[Mode]int{
	MODE_HELPER_ONLY:   0,
	MODE_BILATERAL:     1,
	MODE_REMOTE_HELPER: 2,
}

var IntToModeMap = map[int]Mode{
	0: MODE_HELPER_ONLY,
	1: MODE_BILATERAL,
	2: MODE_REMOTE_HELPER,
}

func (v Mode) Validate() error {
	if _, ok := ModeToIntMap[v]; !ok {
		return fmt.Errorf("invalid Mode: %s", v)
	}
	return nil
}

func (v Mode) ToInt() int {
	i, ok := ModeToIntMap[v]
	if !ok {
		return -1
	}
	return i
}

// typedef for identity bgp:admin-state.
type AdminState string

const (
	ADMIN_STATE_UP     AdminState = "up"
	ADMIN_STATE_DOWN   AdminState = "down"
	ADMIN_STATE_PFX_CT AdminState = "pfx_ct"
)

var AdminStateToIntMap = map[AdminState]int{
	ADMIN_STATE_UP:     0,
	ADMIN_STATE_DOWN:   1,
	ADMIN_STATE_PFX_CT: 2,
}

var IntToAdminStateMap = map[int]AdminState{
	0: ADMIN_STATE_UP,
	1: ADMIN_STATE_DOWN,
	2: ADMIN_STATE_PFX_CT,
}

func (v AdminState) Validate() error {
	if _, ok := AdminStateToIntMap[v]; !ok {
		return fmt.Errorf("invalid AdminState: %s", v)
	}
	return nil
}

func (v AdminState) ToInt() int {
	i, ok := AdminStateToIntMap[v]
	if !ok {
		return -1
	}
	return i
}

// struct for container bgp:neighbor.
// List of BGP neighbors configured on the local system,
// uniquely identified by peer IPv[46] address.
type Neighbor struct {
	// original -> bgp:neighbor-address
	// original -> bgp:neighbor-config
	// Configuration parameters relating to the BGP neighbor or
	// group.
	Config NeighborConfig `mapstructure:"config" json:"config,omitempty"`
	// original -> bgp:neighbor-state
	// State information relating to the BGP neighbor or group.
	State NeighborState `mapstructure:"state" json:"state,omitempty"`
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
	// Parameters related to the use of multiple-paths for the same
	// NLRI when they are received only from this neighbor.
	UseMultiplePaths UseMultiplePaths `mapstructure:"use-multiple-paths" json:"use-multiple-paths,omitempty"`
	// original -> gobgp:route-server
	// Configure the local router as a route server.
	RouteServer RouteServer `mapstructure:"route-server" json:"route-server,omitempty"`
	// original -> gobgp:ttl-security
	// Configure TTL Security feature.
	TtlSecurity TtlSecurity `mapstructure:"ttl-security" json:"ttl-security,omitempty"`
}

// struct for container bgp:state.
// State information relating to the BGP neighbor or group.
type NeighborState struct {
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
	// original -> bgp:peer-group
	// The peer-group with which this neighbor is associated.
	PeerGroup string `mapstructure:"peer-group" json:"peer-group,omitempty"`
	// original -> bgp:neighbor-address
	// bgp:neighbor-address's original type is inet:ip-address.
	// Address of the BGP peer, either in IPv4 or IPv6.
	NeighborAddress string `mapstructure:"neighbor-address" json:"neighbor-address,omitempty"`
	// original -> bgp-op:session-state
	// Operational state of the BGP peer.
	SessionState SessionState `mapstructure:"session-state" json:"session-state,omitempty"`
	// original -> bgp-op:supported-capabilities
	// BGP capabilities negotiated as supported with the peer.
	SupportedCapabilitiesList []BgpCapability `mapstructure:"supported-capabilities-list" json:"supported-capabilities-list,omitempty"`
	// original -> bgp:messages
	// Counters for BGP messages sent and received from the
	// neighbor.
	Messages Messages `mapstructure:"messages" json:"messages,omitempty"`
	// original -> bgp:queues
	// Counters related to queued messages associated with the
	// BGP neighbor.
	Queues Queues `mapstructure:"queues" json:"queues,omitempty"`
	// original -> gobgp:adj-table
	AdjTable AdjTable `mapstructure:"adj-table" json:"adj-table,omitempty"`
	// original -> gobgp:remote-capability
	// original type is list of bgp-capability
	RemoteCapabilityList []bgp.ParameterCapabilityInterface `mapstructure:"remote-capability-list" json:"remote-capability-list,omitempty"`
	// original -> gobgp:local-capability
	// original type is list of bgp-capability
	LocalCapabilityList []bgp.ParameterCapabilityInterface `mapstructure:"local-capability-list" json:"local-capability-list,omitempty"`
	// original -> gobgp:received-open-message
	// gobgp:received-open-message's original type is bgp-open-message.
	ReceivedOpenMessage *bgp.BGPMessage `mapstructure:"received-open-message" json:"received-open-message,omitempty"`
	// original -> gobgp:admin-down
	// gobgp:admin-down's original type is boolean.
	// The state of administrative operation. If the state is true, it indicates the neighbor is disabled by the administrator.
	AdminDown bool `mapstructure:"admin-down" json:"admin-down,omitempty"`
	// original -> gobgp:admin-state
	AdminState AdminState `mapstructure:"admin-state" json:"admin-state,omitempty"`
	// original -> gobgp:established-count
	// The number of how many the peer became established state.
	EstablishedCount uint32 `mapstructure:"established-count" json:"established-count,omitempty"`
	// original -> gobgp:flops
	// The number of flip-flops.
	Flops uint32 `mapstructure:"flops" json:"flops,omitempty"`
	// original -> gobgp:neighbor-interface
	NeighborInterface string `mapstructure:"neighbor-interface" json:"neighbor-interface,omitempty"`
	// original -> gobgp:vrf
	Vrf string `mapstructure:"vrf" json:"vrf,omitempty"`
	// original -> gobgp:remote-router-id
	RemoteRouterId string `mapstructure:"remote-router-id" json:"remote-router-id,omitempty"`
}

// struct for container bgp:config.
// Configuration parameters relating to the BGP neighbor or
// group.
type NeighborConfig struct {
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
	// original -> bgp:peer-group
	// The peer-group with which this neighbor is associated.
	PeerGroup string `mapstructure:"peer-group" json:"peer-group,omitempty"`
	// original -> bgp:neighbor-address
	// bgp:neighbor-address's original type is inet:ip-address.
	// Address of the BGP peer, either in IPv4 or IPv6.
	NeighborAddress string `mapstructure:"neighbor-address" json:"neighbor-address,omitempty"`
	// original -> gobgp:admin-down
	// gobgp:admin-down's original type is boolean.
	// The config of administrative operation. If state, indicates the neighbor is disabled by the administrator.
	AdminDown bool `mapstructure:"admin-down" json:"admin-down,omitempty"`
	// original -> gobgp:neighbor-interface
	NeighborInterface string `mapstructure:"neighbor-interface" json:"neighbor-interface,omitempty"`
	// original -> gobgp:vrf
	Vrf string `mapstructure:"vrf" json:"vrf,omitempty"`
}

func (lhs *NeighborConfig) Equal(rhs *NeighborConfig) bool {
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
	if lhs.PeerGroup != rhs.PeerGroup {
		return false
	}
	if lhs.NeighborAddress != rhs.NeighborAddress {
		return false
	}
	if lhs.AdminDown != rhs.AdminDown {
		return false
	}
	if lhs.NeighborInterface != rhs.NeighborInterface {
		return false
	}
	if lhs.Vrf != rhs.Vrf {
		return false
	}
	return true
}

// struct for container bgp:state.
// State information relating to enhanced error handling
// mechanisms for the BGP neighbor or group.
type ErrorHandlingState struct {
	// original -> bgp:treat-as-withdraw
	// bgp:treat-as-withdraw's original type is boolean.
	// Specify whether erroneous UPDATE messages for which the
	// NLRI can be extracted are reated as though the NLRI is
	// withdrawn - avoiding session reset.
	TreatAsWithdraw bool `mapstructure:"treat-as-withdraw" json:"treat-as-withdraw,omitempty"`
	// original -> bgp-op:erroneous-update-messages
	// The number of BGP UPDATE messages for which the
	// treat-as-withdraw mechanism has been applied based
	// on erroneous message contents.
	ErroneousUpdateMessages uint32 `mapstructure:"erroneous-update-messages" json:"erroneous-update-messages,omitempty"`
}

// struct for container bgp:config.
// Configuration parameters enabling or modifying the
// behavior or enhanced error handling mechanisms for the BGP
// neighbor or group.
type ErrorHandlingConfig struct {
	// original -> bgp:treat-as-withdraw
	// bgp:treat-as-withdraw's original type is boolean.
	// Specify whether erroneous UPDATE messages for which the
	// NLRI can be extracted are reated as though the NLRI is
	// withdrawn - avoiding session reset.
	TreatAsWithdraw bool `mapstructure:"treat-as-withdraw" json:"treat-as-withdraw,omitempty"`
}

func (lhs *ErrorHandlingConfig) Equal(rhs *ErrorHandlingConfig) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if lhs.TreatAsWithdraw != rhs.TreatAsWithdraw {
		return false
	}
	return true
}

// struct for container bgp:error-handling.
// Error handling parameters used for the BGP neighbor or
// group.
type ErrorHandling struct {
	// original -> bgp:error-handling-config
	// Configuration parameters enabling or modifying the
	// behavior or enhanced error handling mechanisms for the BGP
	// neighbor or group.
	Config ErrorHandlingConfig `mapstructure:"config" json:"config,omitempty"`
	// original -> bgp:error-handling-state
	// State information relating to enhanced error handling
	// mechanisms for the BGP neighbor or group.
	State ErrorHandlingState `mapstructure:"state" json:"state,omitempty"`
}

func (lhs *ErrorHandling) Equal(rhs *ErrorHandling) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if !lhs.Config.Equal(&(rhs.Config)) {
		return false
	}
	return true
}

// struct for container bgp:state.
// State information relating to the transport session(s)
// used for the BGP neighbor or group.
type TransportState struct {
	// original -> bgp:tcp-mss
	// Sets the max segment size for BGP TCP sessions.
	TcpMss uint16 `mapstructure:"tcp-mss" json:"tcp-mss,omitempty"`
	// original -> bgp:mtu-discovery
	// bgp:mtu-discovery's original type is boolean.
	// Turns path mtu discovery for BGP TCP sessions on (true)
	// or off (false).
	MtuDiscovery bool `mapstructure:"mtu-discovery" json:"mtu-discovery,omitempty"`
	// original -> bgp:passive-mode
	// bgp:passive-mode's original type is boolean.
	// Wait for peers to issue requests to open a BGP session,
	// rather than initiating sessions from the local router.
	PassiveMode bool `mapstructure:"passive-mode" json:"passive-mode,omitempty"`
	// original -> bgp:local-address
	// bgp:local-address's original type is union.
	// Set the local IP (either IPv4 or IPv6) address to use
	// for the session when sending BGP update messages.  This
	// may be expressed as either an IP address or reference
	// to the name of an interface.
	LocalAddress string `mapstructure:"local-address" json:"local-address,omitempty"`
	// original -> bgp-op:local-port
	// bgp-op:local-port's original type is inet:port-number.
	// Local TCP port being used for the TCP session supporting
	// the BGP session.
	LocalPort uint16 `mapstructure:"local-port" json:"local-port,omitempty"`
	// original -> bgp-op:remote-address
	// bgp-op:remote-address's original type is inet:ip-address.
	// Remote address to which the BGP session has been
	// established.
	RemoteAddress string `mapstructure:"remote-address" json:"remote-address,omitempty"`
	// original -> bgp-op:remote-port
	// bgp-op:remote-port's original type is inet:port-number.
	// Remote port being used by the peer for the TCP session
	// supporting the BGP session.
	RemotePort uint16 `mapstructure:"remote-port" json:"remote-port,omitempty"`
}

// struct for container bgp:config.
// Configuration parameters relating to the transport
// session(s) used for the BGP neighbor or group.
type TransportConfig struct {
	// original -> bgp:tcp-mss
	// Sets the max segment size for BGP TCP sessions.
	TcpMss uint16 `mapstructure:"tcp-mss" json:"tcp-mss,omitempty"`
	// original -> bgp:mtu-discovery
	// bgp:mtu-discovery's original type is boolean.
	// Turns path mtu discovery for BGP TCP sessions on (true)
	// or off (false).
	MtuDiscovery bool `mapstructure:"mtu-discovery" json:"mtu-discovery,omitempty"`
	// original -> bgp:passive-mode
	// bgp:passive-mode's original type is boolean.
	// Wait for peers to issue requests to open a BGP session,
	// rather than initiating sessions from the local router.
	PassiveMode bool `mapstructure:"passive-mode" json:"passive-mode,omitempty"`
	// original -> bgp:local-address
	// bgp:local-address's original type is union.
	// Set the local IP (either IPv4 or IPv6) address to use
	// for the session when sending BGP update messages.  This
	// may be expressed as either an IP address or reference
	// to the name of an interface.
	LocalAddress string `mapstructure:"local-address" json:"local-address,omitempty"`
	// original -> gobgp:remote-port
	// gobgp:remote-port's original type is inet:port-number.
	RemotePort uint16 `mapstructure:"remote-port" json:"remote-port,omitempty"`
	// original -> gobgp:ttl
	// TTL value for BGP packets.
	Ttl uint8 `mapstructure:"ttl" json:"ttl,omitempty"`
	// original -> gobgp:bind-interface
	// Interface name for binding.
	BindInterface string `mapstructure:"bind-interface" json:"bind-interface,omitempty"`
}

func (lhs *TransportConfig) Equal(rhs *TransportConfig) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if lhs.TcpMss != rhs.TcpMss {
		return false
	}
	if lhs.MtuDiscovery != rhs.MtuDiscovery {
		return false
	}
	if lhs.PassiveMode != rhs.PassiveMode {
		return false
	}
	if lhs.LocalAddress != rhs.LocalAddress {
		return false
	}
	if lhs.RemotePort != rhs.RemotePort {
		return false
	}
	if lhs.Ttl != rhs.Ttl {
		return false
	}
	if lhs.BindInterface != rhs.BindInterface {
		return false
	}
	return true
}

// struct for container bgp:transport.
// Transport session parameters for the BGP neighbor or group.
type Transport struct {
	// original -> bgp:transport-config
	// Configuration parameters relating to the transport
	// session(s) used for the BGP neighbor or group.
	Config TransportConfig `mapstructure:"config" json:"config,omitempty"`
	// original -> bgp:transport-state
	// State information relating to the transport session(s)
	// used for the BGP neighbor or group.
	State TransportState `mapstructure:"state" json:"state,omitempty"`
}

func (lhs *Transport) Equal(rhs *Transport) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if !lhs.Config.Equal(&(rhs.Config)) {
		return false
	}
	return true
}

// struct for container bgp:state.
// State information relating to the timers used for the BGP
// neighbor or group.
type TimersState struct {
	// original -> bgp:connect-retry
	// bgp:connect-retry's original type is decimal64.
	// Time interval in seconds between attempts to establish a
	// session with the peer.
	ConnectRetry float64 `mapstructure:"connect-retry" json:"connect-retry,omitempty"`
	// original -> bgp:hold-time
	// bgp:hold-time's original type is decimal64.
	// Time interval in seconds that a BGP session will be
	// considered active in the absence of keepalive or other
	// messages from the peer.  The hold-time is typically
	// set to 3x the keepalive-interval.
	HoldTime float64 `mapstructure:"hold-time" json:"hold-time,omitempty"`
	// original -> bgp:keepalive-interval
	// bgp:keepalive-interval's original type is decimal64.
	// Time interval in seconds between transmission of keepalive
	// messages to the neighbor.  Typically set to 1/3 the
	// hold-time.
	KeepaliveInterval float64 `mapstructure:"keepalive-interval" json:"keepalive-interval,omitempty"`
	// original -> bgp:minimum-advertisement-interval
	// bgp:minimum-advertisement-interval's original type is decimal64.
	// Minimum time which must elapse between subsequent UPDATE
	// messages relating to a common set of NLRI being transmitted
	// to a peer. This timer is referred to as
	// MinRouteAdvertisementIntervalTimer by RFC 4721 and serves to
	// reduce the number of UPDATE messages transmitted when a
	// particular set of NLRI exhibit instability.
	MinimumAdvertisementInterval float64 `mapstructure:"minimum-advertisement-interval" json:"minimum-advertisement-interval,omitempty"`
	// original -> bgp-op:uptime
	// bgp-op:uptime's original type is yang:timeticks.
	// This timer determines the amount of time since the
	// BGP last transitioned in or out of the Established
	// state.
	Uptime int64 `mapstructure:"uptime" json:"uptime,omitempty"`
	// original -> bgp-op:negotiated-hold-time
	// bgp-op:negotiated-hold-time's original type is decimal64.
	// The negotiated hold-time for the BGP session.
	NegotiatedHoldTime float64 `mapstructure:"negotiated-hold-time" json:"negotiated-hold-time,omitempty"`
	// original -> gobgp:idle-hold-time-after-reset
	// gobgp:idle-hold-time-after-reset's original type is decimal64.
	// Time interval in seconds that a BGP session will be
	// in idle state after neighbor reset operation.
	IdleHoldTimeAfterReset float64 `mapstructure:"idle-hold-time-after-reset" json:"idle-hold-time-after-reset,omitempty"`
	// original -> gobgp:downtime
	// gobgp:downtime's original type is yang:timeticks.
	// This timer determines the amount of time since the
	// BGP last transitioned out of the Established state.
	Downtime int64 `mapstructure:"downtime" json:"downtime,omitempty"`
	// original -> gobgp:update-recv-time
	// The number of seconds elapsed since January 1, 1970 UTC
	// last time the BGP session received an UPDATE message.
	UpdateRecvTime int64 `mapstructure:"update-recv-time" json:"update-recv-time,omitempty"`
}

// struct for container bgp:config.
// Configuration parameters relating to timers used for the
// BGP neighbor or group.
type TimersConfig struct {
	// original -> bgp:connect-retry
	// bgp:connect-retry's original type is decimal64.
	// Time interval in seconds between attempts to establish a
	// session with the peer.
	ConnectRetry float64 `mapstructure:"connect-retry" json:"connect-retry,omitempty"`
	// original -> bgp:hold-time
	// bgp:hold-time's original type is decimal64.
	// Time interval in seconds that a BGP session will be
	// considered active in the absence of keepalive or other
	// messages from the peer.  The hold-time is typically
	// set to 3x the keepalive-interval.
	HoldTime float64 `mapstructure:"hold-time" json:"hold-time,omitempty"`
	// original -> bgp:keepalive-interval
	// bgp:keepalive-interval's original type is decimal64.
	// Time interval in seconds between transmission of keepalive
	// messages to the neighbor.  Typically set to 1/3 the
	// hold-time.
	KeepaliveInterval float64 `mapstructure:"keepalive-interval" json:"keepalive-interval,omitempty"`
	// original -> bgp:minimum-advertisement-interval
	// bgp:minimum-advertisement-interval's original type is decimal64.
	// Minimum time which must elapse between subsequent UPDATE
	// messages relating to a common set of NLRI being transmitted
	// to a peer. This timer is referred to as
	// MinRouteAdvertisementIntervalTimer by RFC 4721 and serves to
	// reduce the number of UPDATE messages transmitted when a
	// particular set of NLRI exhibit instability.
	MinimumAdvertisementInterval float64 `mapstructure:"minimum-advertisement-interval" json:"minimum-advertisement-interval,omitempty"`
	// original -> gobgp:idle-hold-time-after-reset
	// gobgp:idle-hold-time-after-reset's original type is decimal64.
	// Time interval in seconds that a BGP session will be
	// in idle state after neighbor reset operation.
	IdleHoldTimeAfterReset float64 `mapstructure:"idle-hold-time-after-reset" json:"idle-hold-time-after-reset,omitempty"`
}

func (lhs *TimersConfig) Equal(rhs *TimersConfig) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if lhs.ConnectRetry != rhs.ConnectRetry {
		return false
	}
	if lhs.HoldTime != rhs.HoldTime {
		return false
	}
	if lhs.KeepaliveInterval != rhs.KeepaliveInterval {
		return false
	}
	if lhs.MinimumAdvertisementInterval != rhs.MinimumAdvertisementInterval {
		return false
	}
	if lhs.IdleHoldTimeAfterReset != rhs.IdleHoldTimeAfterReset {
		return false
	}
	return true
}

// struct for container bgp:timers.
// Timers related to a BGP neighbor or group.
type Timers struct {
	// original -> bgp:timers-config
	// Configuration parameters relating to timers used for the
	// BGP neighbor or group.
	Config TimersConfig `mapstructure:"config" json:"config,omitempty"`
	// original -> bgp:timers-state
	// State information relating to the timers used for the BGP
	// neighbor or group.
	State TimersState `mapstructure:"state" json:"state,omitempty"`
}

func (lhs *Timers) Equal(rhs *Timers) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if !lhs.Config.Equal(&(rhs.Config)) {
		return false
	}
	return true
}

// struct for container bgp:state.
// State information relating to the AS_PATH manipulation
// mechanisms for the BGP peer or group.
type AsPathOptionsState struct {
	// original -> bgp:allow-own-as
	// Specify the number of occurrences of the local BGP speaker's
	// AS that can occur within the AS_PATH before it is rejected.
	AllowOwnAs uint8 `mapstructure:"allow-own-as" json:"allow-own-as,omitempty"`
	// original -> bgp:replace-peer-as
	// bgp:replace-peer-as's original type is boolean.
	// Replace occurrences of the peer's AS in the AS_PATH
	// with the local autonomous system number.
	ReplacePeerAs bool `mapstructure:"replace-peer-as" json:"replace-peer-as,omitempty"`
}

// struct for container bgp:config.
// Configuration parameters relating to AS_PATH manipulation
// for the BGP peer or group.
type AsPathOptionsConfig struct {
	// original -> bgp:allow-own-as
	// Specify the number of occurrences of the local BGP speaker's
	// AS that can occur within the AS_PATH before it is rejected.
	AllowOwnAs uint8 `mapstructure:"allow-own-as" json:"allow-own-as,omitempty"`
	// original -> bgp:replace-peer-as
	// bgp:replace-peer-as's original type is boolean.
	// Replace occurrences of the peer's AS in the AS_PATH
	// with the local autonomous system number.
	ReplacePeerAs bool `mapstructure:"replace-peer-as" json:"replace-peer-as,omitempty"`
}

func (lhs *AsPathOptionsConfig) Equal(rhs *AsPathOptionsConfig) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if lhs.AllowOwnAs != rhs.AllowOwnAs {
		return false
	}
	if lhs.ReplacePeerAs != rhs.ReplacePeerAs {
		return false
	}
	return true
}

// struct for container bgp:as-path-options.
// AS_PATH manipulation parameters for the BGP neighbor or
// group.
type AsPathOptions struct {
	// original -> bgp:as-path-options-config
	// Configuration parameters relating to AS_PATH manipulation
	// for the BGP peer or group.
	Config AsPathOptionsConfig `mapstructure:"config" json:"config,omitempty"`
	// original -> bgp:as-path-options-state
	// State information relating to the AS_PATH manipulation
	// mechanisms for the BGP peer or group.
	State AsPathOptionsState `mapstructure:"state" json:"state,omitempty"`
}

func (lhs *AsPathOptions) Equal(rhs *AsPathOptions) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if !lhs.Config.Equal(&(rhs.Config)) {
		return false
	}
	return true
}

// struct for container bgp:state.
// State information relating to route reflection for the
// BGP neighbor or group.
type RouteReflectorState struct {
	// original -> bgp:route-reflector-cluster-id
	// route-reflector cluster id to use when local router is
	// configured as a route reflector.  Commonly set at the group
	// level, but allows a different cluster
	// id to be set for each neighbor.
	RouteReflectorClusterId RrClusterIdType `mapstructure:"route-reflector-cluster-id" json:"route-reflector-cluster-id,omitempty"`
	// original -> bgp:route-reflector-client
	// bgp:route-reflector-client's original type is boolean.
	// Configure the neighbor as a route reflector client.
	RouteReflectorClient bool `mapstructure:"route-reflector-client" json:"route-reflector-client,omitempty"`
}

// struct for container bgp:config.
// Configuraton parameters relating to route reflection
// for the BGP neighbor or group.
type RouteReflectorConfig struct {
	// original -> bgp:route-reflector-cluster-id
	// route-reflector cluster id to use when local router is
	// configured as a route reflector.  Commonly set at the group
	// level, but allows a different cluster
	// id to be set for each neighbor.
	RouteReflectorClusterId RrClusterIdType `mapstructure:"route-reflector-cluster-id" json:"route-reflector-cluster-id,omitempty"`
	// original -> bgp:route-reflector-client
	// bgp:route-reflector-client's original type is boolean.
	// Configure the neighbor as a route reflector client.
	RouteReflectorClient bool `mapstructure:"route-reflector-client" json:"route-reflector-client,omitempty"`
}

func (lhs *RouteReflectorConfig) Equal(rhs *RouteReflectorConfig) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if lhs.RouteReflectorClusterId != rhs.RouteReflectorClusterId {
		return false
	}
	if lhs.RouteReflectorClient != rhs.RouteReflectorClient {
		return false
	}
	return true
}

// struct for container bgp:route-reflector.
// Route reflector parameters for the BGP neighbor or group.
type RouteReflector struct {
	// original -> bgp:route-reflector-config
	// Configuraton parameters relating to route reflection
	// for the BGP neighbor or group.
	Config RouteReflectorConfig `mapstructure:"config" json:"config,omitempty"`
	// original -> bgp:route-reflector-state
	// State information relating to route reflection for the
	// BGP neighbor or group.
	State RouteReflectorState `mapstructure:"state" json:"state,omitempty"`
}

func (lhs *RouteReflector) Equal(rhs *RouteReflector) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if !lhs.Config.Equal(&(rhs.Config)) {
		return false
	}
	return true
}

// struct for container bgp:state.
// State information for eBGP multihop, for the BGP neighbor
// or group.
type EbgpMultihopState struct {
	// original -> bgp:enabled
	// bgp:enabled's original type is boolean.
	// When enabled the referenced group or neighbors are permitted
	// to be indirectly connected - including cases where the TTL
	// can be decremented between the BGP peers.
	Enabled bool `mapstructure:"enabled" json:"enabled,omitempty"`
	// original -> bgp:multihop-ttl
	// Time-to-live value to use when packets are sent to the
	// referenced group or neighbors and ebgp-multihop is enabled.
	MultihopTtl uint8 `mapstructure:"multihop-ttl" json:"multihop-ttl,omitempty"`
}

// struct for container bgp:config.
// Configuration parameters relating to eBGP multihop for the
// BGP neighbor or group.
type EbgpMultihopConfig struct {
	// original -> bgp:enabled
	// bgp:enabled's original type is boolean.
	// When enabled the referenced group or neighbors are permitted
	// to be indirectly connected - including cases where the TTL
	// can be decremented between the BGP peers.
	Enabled bool `mapstructure:"enabled" json:"enabled,omitempty"`
	// original -> bgp:multihop-ttl
	// Time-to-live value to use when packets are sent to the
	// referenced group or neighbors and ebgp-multihop is enabled.
	MultihopTtl uint8 `mapstructure:"multihop-ttl" json:"multihop-ttl,omitempty"`
}

func (lhs *EbgpMultihopConfig) Equal(rhs *EbgpMultihopConfig) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if lhs.Enabled != rhs.Enabled {
		return false
	}
	if lhs.MultihopTtl != rhs.MultihopTtl {
		return false
	}
	return true
}

// struct for container bgp:ebgp-multihop.
// eBGP multi-hop parameters for the BGP neighbor or group.
type EbgpMultihop struct {
	// original -> bgp:ebgp-multihop-config
	// Configuration parameters relating to eBGP multihop for the
	// BGP neighbor or group.
	Config EbgpMultihopConfig `mapstructure:"config" json:"config,omitempty"`
	// original -> bgp:ebgp-multihop-state
	// State information for eBGP multihop, for the BGP neighbor
	// or group.
	State EbgpMultihopState `mapstructure:"state" json:"state,omitempty"`
}

func (lhs *EbgpMultihop) Equal(rhs *EbgpMultihop) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if !lhs.Config.Equal(&(rhs.Config)) {
		return false
	}
	return true
}

// struct for container bgp:state.
// State information relating to logging for the BGP neighbor
// or group.
type LoggingOptionsState struct {
	// original -> bgp:log-neighbor-state-changes
	// bgp:log-neighbor-state-changes's original type is boolean.
	// Configure logging of peer state changes.  Default is
	// to enable logging of peer state changes.
	LogNeighborStateChanges bool `mapstructure:"log-neighbor-state-changes" json:"log-neighbor-state-changes,omitempty"`
}

// struct for container bgp:config.
// Configuration parameters enabling or modifying logging
// for events relating to the BGP neighbor or group.
type LoggingOptionsConfig struct {
	// original -> bgp:log-neighbor-state-changes
	// bgp:log-neighbor-state-changes's original type is boolean.
	// Configure logging of peer state changes.  Default is
	// to enable logging of peer state changes.
	LogNeighborStateChanges bool `mapstructure:"log-neighbor-state-changes" json:"log-neighbor-state-changes,omitempty"`
}

func (lhs *LoggingOptionsConfig) Equal(rhs *LoggingOptionsConfig) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if lhs.LogNeighborStateChanges != rhs.LogNeighborStateChanges {
		return false
	}
	return true
}

// struct for container bgp:logging-options.
// Logging options for events related to the BGP neighbor or
// group.
type LoggingOptions struct {
	// original -> bgp:logging-options-config
	// Configuration parameters enabling or modifying logging
	// for events relating to the BGP neighbor or group.
	Config LoggingOptionsConfig `mapstructure:"config" json:"config,omitempty"`
	// original -> bgp:logging-options-state
	// State information relating to logging for the BGP neighbor
	// or group.
	State LoggingOptionsState `mapstructure:"state" json:"state,omitempty"`
}

func (lhs *LoggingOptions) Equal(rhs *LoggingOptions) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if !lhs.Config.Equal(&(rhs.Config)) {
		return false
	}
	return true
}

// struct for container bgp:state.
// State information associated with graceful-restart.
type GracefulRestartState struct {
	// original -> bgp:enabled
	// bgp:enabled's original type is boolean.
	// Enable or disable the graceful-restart capability.
	Enabled bool `mapstructure:"enabled" json:"enabled,omitempty"`
	// original -> bgp:restart-time
	// Estimated time (in seconds) for the local BGP speaker to
	// restart a session. This value is advertise in the graceful
	// restart BGP capability.  This is a 12-bit value, referred to
	// as Restart Time in RFC4724.  Per RFC4724, the suggested
	// default value is <= the hold-time value.
	RestartTime uint16 `mapstructure:"restart-time" json:"restart-time,omitempty"`
	// original -> bgp:stale-routes-time
	// bgp:stale-routes-time's original type is decimal64.
	// An upper-bound on the time thate stale routes will be
	// retained by a router after a session is restarted. If an
	// End-of-RIB (EOR) marker is received prior to this timer
	// expiring stale-routes will be flushed upon its receipt - if
	// no EOR is received, then when this timer expires stale paths
	// will be purged. This timer is referred to as the
	// Selection_Deferral_Timer in RFC4724.
	StaleRoutesTime float64 `mapstructure:"stale-routes-time" json:"stale-routes-time,omitempty"`
	// original -> bgp:helper-only
	// bgp:helper-only's original type is boolean.
	// Enable graceful-restart in helper mode only. When this
	// leaf is set, the local system does not retain forwarding
	// its own state during a restart, but supports procedures
	// for the receiving speaker, as defined in RFC4724.
	HelperOnly bool `mapstructure:"helper-only" json:"helper-only,omitempty"`
	// original -> bgp-op:peer-restart-time
	// The period of time (advertised by the peer) that
	// the peer expects a restart of a BGP session to
	// take.
	PeerRestartTime uint16 `mapstructure:"peer-restart-time" json:"peer-restart-time,omitempty"`
	// original -> bgp-op:peer-restarting
	// bgp-op:peer-restarting's original type is boolean.
	// This flag indicates whether the remote neighbor is currently
	// in the process of restarting, and hence received routes are
	// currently stale.
	PeerRestarting bool `mapstructure:"peer-restarting" json:"peer-restarting,omitempty"`
	// original -> bgp-op:local-restarting
	// bgp-op:local-restarting's original type is boolean.
	// This flag indicates whether the local neighbor is currently
	// restarting. The flag is unset after all NLRI have been
	// advertised to the peer, and the End-of-RIB (EOR) marker has
	// been unset.
	LocalRestarting bool `mapstructure:"local-restarting" json:"local-restarting,omitempty"`
	// original -> bgp-op:mode
	// Ths leaf indicates the mode of operation of BGP graceful
	// restart with the peer.
	Mode Mode `mapstructure:"mode" json:"mode,omitempty"`
	// original -> gobgp:deferral-time
	DeferralTime uint16 `mapstructure:"deferral-time" json:"deferral-time,omitempty"`
	// original -> gobgp:notification-enabled
	// gobgp:notification-enabled's original type is boolean.
	NotificationEnabled bool `mapstructure:"notification-enabled" json:"notification-enabled,omitempty"`
	// original -> gobgp:long-lived-enabled
	// gobgp:long-lived-enabled's original type is boolean.
	LongLivedEnabled bool `mapstructure:"long-lived-enabled" json:"long-lived-enabled,omitempty"`
}

// struct for container bgp:config.
// Configuration parameters relating to graceful-restart.
type GracefulRestartConfig struct {
	// original -> bgp:enabled
	// bgp:enabled's original type is boolean.
	// Enable or disable the graceful-restart capability.
	Enabled bool `mapstructure:"enabled" json:"enabled,omitempty"`
	// original -> bgp:restart-time
	// Estimated time (in seconds) for the local BGP speaker to
	// restart a session. This value is advertise in the graceful
	// restart BGP capability.  This is a 12-bit value, referred to
	// as Restart Time in RFC4724.  Per RFC4724, the suggested
	// default value is <= the hold-time value.
	RestartTime uint16 `mapstructure:"restart-time" json:"restart-time,omitempty"`
	// original -> bgp:stale-routes-time
	// bgp:stale-routes-time's original type is decimal64.
	// An upper-bound on the time thate stale routes will be
	// retained by a router after a session is restarted. If an
	// End-of-RIB (EOR) marker is received prior to this timer
	// expiring stale-routes will be flushed upon its receipt - if
	// no EOR is received, then when this timer expires stale paths
	// will be purged. This timer is referred to as the
	// Selection_Deferral_Timer in RFC4724.
	StaleRoutesTime float64 `mapstructure:"stale-routes-time" json:"stale-routes-time,omitempty"`
	// original -> bgp:helper-only
	// bgp:helper-only's original type is boolean.
	// Enable graceful-restart in helper mode only. When this
	// leaf is set, the local system does not retain forwarding
	// its own state during a restart, but supports procedures
	// for the receiving speaker, as defined in RFC4724.
	HelperOnly bool `mapstructure:"helper-only" json:"helper-only,omitempty"`
	// original -> gobgp:deferral-time
	DeferralTime uint16 `mapstructure:"deferral-time" json:"deferral-time,omitempty"`
	// original -> gobgp:notification-enabled
	// gobgp:notification-enabled's original type is boolean.
	NotificationEnabled bool `mapstructure:"notification-enabled" json:"notification-enabled,omitempty"`
	// original -> gobgp:long-lived-enabled
	// gobgp:long-lived-enabled's original type is boolean.
	LongLivedEnabled bool `mapstructure:"long-lived-enabled" json:"long-lived-enabled,omitempty"`
}

func (lhs *GracefulRestartConfig) Equal(rhs *GracefulRestartConfig) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if lhs.Enabled != rhs.Enabled {
		return false
	}
	if lhs.RestartTime != rhs.RestartTime {
		return false
	}
	if lhs.StaleRoutesTime != rhs.StaleRoutesTime {
		return false
	}
	if lhs.HelperOnly != rhs.HelperOnly {
		return false
	}
	if lhs.DeferralTime != rhs.DeferralTime {
		return false
	}
	if lhs.NotificationEnabled != rhs.NotificationEnabled {
		return false
	}
	if lhs.LongLivedEnabled != rhs.LongLivedEnabled {
		return false
	}
	return true
}

// struct for container bgp:graceful-restart.
// Parameters relating the graceful restart mechanism for BGP.
type GracefulRestart struct {
	// original -> bgp:graceful-restart-config
	// Configuration parameters relating to graceful-restart.
	Config GracefulRestartConfig `mapstructure:"config" json:"config,omitempty"`
	// original -> bgp:graceful-restart-state
	// State information associated with graceful-restart.
	State GracefulRestartState `mapstructure:"state" json:"state,omitempty"`
}

func (lhs *GracefulRestart) Equal(rhs *GracefulRestart) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if !lhs.Config.Equal(&(rhs.Config)) {
		return false
	}
	return true
}

// struct for container gobgp:state.
// State information for TTL Security.
type TtlSecurityState struct {
	// original -> gobgp:enabled
	// gobgp:enabled's original type is boolean.
	// Enable features for TTL Security.
	Enabled bool `mapstructure:"enabled" json:"enabled,omitempty"`
	// original -> gobgp:ttl-min
	// Reference to the port of the BMP server.
	TtlMin uint8 `mapstructure:"ttl-min" json:"ttl-min,omitempty"`
}

// struct for container gobgp:config.
// Configuration parameters for TTL Security.
type TtlSecurityConfig struct {
	// original -> gobgp:enabled
	// gobgp:enabled's original type is boolean.
	// Enable features for TTL Security.
	Enabled bool `mapstructure:"enabled" json:"enabled,omitempty"`
	// original -> gobgp:ttl-min
	// Reference to the port of the BMP server.
	TtlMin uint8 `mapstructure:"ttl-min" json:"ttl-min,omitempty"`
}

func (lhs *TtlSecurityConfig) Equal(rhs *TtlSecurityConfig) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if lhs.Enabled != rhs.Enabled {
		return false
	}
	if lhs.TtlMin != rhs.TtlMin {
		return false
	}
	return true
}

// struct for container gobgp:ttl-security.
// Configure TTL Security feature.
type TtlSecurity struct {
	// original -> gobgp:ttl-security-config
	// Configuration parameters for TTL Security.
	Config TtlSecurityConfig `mapstructure:"config" json:"config,omitempty"`
	// original -> gobgp:ttl-security-state
	// State information for TTL Security.
	State TtlSecurityState `mapstructure:"state" json:"state,omitempty"`
}

func (lhs *TtlSecurity) Equal(rhs *TtlSecurity) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if !lhs.Config.Equal(&(rhs.Config)) {
		return false
	}
	return true
}

// struct for container gobgp:state.
// State information relating to route server
// client(s) used for the BGP neighbor.
type RouteServerState struct {
	// original -> gobgp:route-server-client
	// gobgp:route-server-client's original type is boolean.
	// Configure the neighbor as a route server client.
	RouteServerClient bool `mapstructure:"route-server-client" json:"route-server-client,omitempty"`
	// original -> gobgp:secondary-route
	// gobgp:secondary-route's original type is boolean.
	// if an export policy rejects a selected route, try the next route in
	// order until one that is accepted is found or all routes for the peer
	// are rejected.
	SecondaryRoute bool `mapstructure:"secondary-route" json:"secondary-route,omitempty"`
}

// struct for container gobgp:config.
// Configuration parameters relating to route server
// client(s) used for the BGP neighbor.
type RouteServerConfig struct {
	// original -> gobgp:route-server-client
	// gobgp:route-server-client's original type is boolean.
	// Configure the neighbor as a route server client.
	RouteServerClient bool `mapstructure:"route-server-client" json:"route-server-client,omitempty"`
	// original -> gobgp:secondary-route
	// gobgp:secondary-route's original type is boolean.
	// if an export policy rejects a selected route, try the next route in
	// order until one that is accepted is found or all routes for the peer
	// are rejected.
	SecondaryRoute bool `mapstructure:"secondary-route" json:"secondary-route,omitempty"`
}

func (lhs *RouteServerConfig) Equal(rhs *RouteServerConfig) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if lhs.RouteServerClient != rhs.RouteServerClient {
		return false
	}
	if lhs.SecondaryRoute != rhs.SecondaryRoute {
		return false
	}
	return true
}

// struct for container gobgp:route-server.
// Configure the local router as a route server.
type RouteServer struct {
	// original -> gobgp:route-server-config
	// Configuration parameters relating to route server
	// client(s) used for the BGP neighbor.
	Config RouteServerConfig `mapstructure:"config" json:"config,omitempty"`
	// original -> gobgp:route-server-state
	// State information relating to route server
	// client(s) used for the BGP neighbor.
	State RouteServerState `mapstructure:"state" json:"state,omitempty"`
}

func (lhs *RouteServer) Equal(rhs *RouteServer) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if !lhs.Config.Equal(&(rhs.Config)) {
		return false
	}
	return true
}

// struct for container bgp:messages.
// Counters for BGP messages sent and received from the
// neighbor.
type Messages struct {
	// original -> bgp:sent
	// Counters relating to BGP messages sent to the neighbor.
	Sent Sent `mapstructure:"sent" json:"sent,omitempty"`
	// original -> bgp:received
	// Counters for BGP messages received from the neighbor.
	Received Received `mapstructure:"received" json:"received,omitempty"`
}

func (lhs *Messages) Equal(rhs *Messages) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if !lhs.Sent.Equal(&(rhs.Sent)) {
		return false
	}
	if !lhs.Received.Equal(&(rhs.Received)) {
		return false
	}
	return true
}

// struct for container bgp:queues.
// Counters related to queued messages associated with the
// BGP neighbor.
type Queues struct {
	// original -> bgp-op:input
	// The number of messages received from the peer currently
	// queued.
	Input uint32 `mapstructure:"input" json:"input,omitempty"`
	// original -> bgp-op:output
	// The number of messages queued to be sent to the peer.
	Output uint32 `mapstructure:"output" json:"output,omitempty"`
}

func (lhs *Queues) Equal(rhs *Queues) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if lhs.Input != rhs.Input {
		return false
	}
	if lhs.Output != rhs.Output {
		return false
	}
	return true
}

// struct for container bgp:received.
// Counters for BGP messages received from the neighbor.
type Received struct {
	// original -> bgp-op:UPDATE
	// Number of BGP UPDATE messages announcing, withdrawing
	// or modifying paths exchanged.
	Update uint64 `mapstructure:"update" json:"update,omitempty"`
	// original -> bgp-op:NOTIFICATION
	// Number of BGP NOTIFICATION messages indicating an
	// error condition has occurred exchanged.
	Notification uint64 `mapstructure:"notification" json:"notification,omitempty"`
	// original -> gobgp:OPEN
	// Number of BGP open messages announcing, withdrawing
	// or modifying paths exchanged.
	Open uint64 `mapstructure:"open" json:"open,omitempty"`
	// original -> gobgp:REFRESH
	// Number of BGP Route-Refresh messages indicating an
	// error condition has occurred exchanged.
	Refresh uint64 `mapstructure:"refresh" json:"refresh,omitempty"`
	// original -> gobgp:KEEPALIVE
	// Number of BGP Keepalive messages indicating an
	// error condition has occurred exchanged.
	Keepalive uint64 `mapstructure:"keepalive" json:"keepalive,omitempty"`
	// original -> gobgp:DYNAMIC-CAP
	// Number of BGP dynamic-cap messages indicating an
	// error condition has occurred exchanged.
	DynamicCap uint64 `mapstructure:"dynamic-cap" json:"dynamic-cap,omitempty"`
	// original -> gobgp:WITHDRAW-UPDATE
	// Number of updates subjected to treat-as-withdraw treatment.
	WithdrawUpdate uint32 `mapstructure:"withdraw-update" json:"withdraw-update,omitempty"`
	// original -> gobgp:WITHDRAW-PREFIX
	// Number of prefixes subjected to treat-as-withdraw treatment.
	WithdrawPrefix uint32 `mapstructure:"withdraw-prefix" json:"withdraw-prefix,omitempty"`
	// original -> gobgp:DISCARDED
	// Number of discarded messages indicating an
	// error condition has occurred exchanged.
	Discarded uint64 `mapstructure:"discarded" json:"discarded,omitempty"`
	// original -> gobgp:TOTAL
	// Number of total messages indicating an
	// error condition has occurred exchanged.
	Total uint64 `mapstructure:"total" json:"total,omitempty"`
}

func (lhs *Received) Equal(rhs *Received) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if lhs.Update != rhs.Update {
		return false
	}
	if lhs.Notification != rhs.Notification {
		return false
	}
	if lhs.Open != rhs.Open {
		return false
	}
	if lhs.Refresh != rhs.Refresh {
		return false
	}
	if lhs.Keepalive != rhs.Keepalive {
		return false
	}
	if lhs.DynamicCap != rhs.DynamicCap {
		return false
	}
	if lhs.WithdrawUpdate != rhs.WithdrawUpdate {
		return false
	}
	if lhs.WithdrawPrefix != rhs.WithdrawPrefix {
		return false
	}
	if lhs.Discarded != rhs.Discarded {
		return false
	}
	if lhs.Total != rhs.Total {
		return false
	}
	return true
}

// struct for container bgp:sent.
// Counters relating to BGP messages sent to the neighbor.
type Sent struct {
	// original -> bgp-op:UPDATE
	// Number of BGP UPDATE messages announcing, withdrawing
	// or modifying paths exchanged.
	Update uint64 `mapstructure:"update" json:"update,omitempty"`
	// original -> bgp-op:NOTIFICATION
	// Number of BGP NOTIFICATION messages indicating an
	// error condition has occurred exchanged.
	Notification uint64 `mapstructure:"notification" json:"notification,omitempty"`
	// original -> gobgp:OPEN
	// Number of BGP open messages announcing, withdrawing
	// or modifying paths exchanged.
	Open uint64 `mapstructure:"open" json:"open,omitempty"`
	// original -> gobgp:REFRESH
	// Number of BGP Route-Refresh messages indicating an
	// error condition has occurred exchanged.
	Refresh uint64 `mapstructure:"refresh" json:"refresh,omitempty"`
	// original -> gobgp:KEEPALIVE
	// Number of BGP Keepalive messages indicating an
	// error condition has occurred exchanged.
	Keepalive uint64 `mapstructure:"keepalive" json:"keepalive,omitempty"`
	// original -> gobgp:DYNAMIC-CAP
	// Number of BGP dynamic-cap messages indicating an
	// error condition has occurred exchanged.
	DynamicCap uint64 `mapstructure:"dynamic-cap" json:"dynamic-cap,omitempty"`
	// original -> gobgp:WITHDRAW-UPDATE
	// Number of updates subjected to treat-as-withdraw treatment.
	WithdrawUpdate uint32 `mapstructure:"withdraw-update" json:"withdraw-update,omitempty"`
	// original -> gobgp:WITHDRAW-PREFIX
	// Number of prefixes subjected to treat-as-withdraw treatment.
	WithdrawPrefix uint32 `mapstructure:"withdraw-prefix" json:"withdraw-prefix,omitempty"`
	// original -> gobgp:DISCARDED
	// Number of discarded messages indicating an
	// error condition has occurred exchanged.
	Discarded uint64 `mapstructure:"discarded" json:"discarded,omitempty"`
	// original -> gobgp:TOTAL
	// Number of total messages indicating an
	// error condition has occurred exchanged.
	Total uint64 `mapstructure:"total" json:"total,omitempty"`
}

func (lhs *Sent) Equal(rhs *Sent) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if lhs.Update != rhs.Update {
		return false
	}
	if lhs.Notification != rhs.Notification {
		return false
	}
	if lhs.Open != rhs.Open {
		return false
	}
	if lhs.Refresh != rhs.Refresh {
		return false
	}
	if lhs.Keepalive != rhs.Keepalive {
		return false
	}
	if lhs.DynamicCap != rhs.DynamicCap {
		return false
	}
	if lhs.WithdrawUpdate != rhs.WithdrawUpdate {
		return false
	}
	if lhs.WithdrawPrefix != rhs.WithdrawPrefix {
		return false
	}
	if lhs.Discarded != rhs.Discarded {
		return false
	}
	if lhs.Total != rhs.Total {
		return false
	}
	return true
}

// struct for container gobgp:adj-table.
type AdjTable struct {
	// original -> gobgp:ADVERTISED
	Advertised uint32 `mapstructure:"advertised" json:"advertised,omitempty"`
	// original -> gobgp:FILTERED
	Filtered uint32 `mapstructure:"filtered" json:"filtered,omitempty"`
	// original -> gobgp:RECEIVED
	Received uint32 `mapstructure:"received" json:"received,omitempty"`
	// original -> gobgp:ACCEPTED
	Accepted uint32 `mapstructure:"accepted" json:"accepted,omitempty"`
}

func (lhs *AdjTable) Equal(rhs *AdjTable) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	if lhs.Advertised != rhs.Advertised {
		return false
	}
	if lhs.Filtered != rhs.Filtered {
		return false
	}
	if lhs.Received != rhs.Received {
		return false
	}
	if lhs.Accepted != rhs.Accepted {
		return false
	}
	return true
}
