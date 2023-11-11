package datapath

import (
	"context"
	"github.com/cilium/cilium/pkg/addressing"
	"github.com/cilium/cilium/pkg/datapath/loader/metrics"
	"github.com/cilium/cilium/pkg/identity"
	"github.com/cilium/cilium/pkg/mac"
	"github.com/cilium/cilium/pkg/option"
	"github.com/sirupsen/logrus"
)

type IptablesManager interface {
	// InstallProxyRules creates the necessary datapath config (e.g., iptables
	// rules for redirecting host proxy traffic on a specific ProxyPort)
	InstallProxyRules(ctx context.Context, proxyPort uint16, ingress bool, name string) error

	// SupportsOriginalSourceAddr tells if the datapath supports
	// use of original source addresses in proxy upstream
	// connections.
	SupportsOriginalSourceAddr() bool
	InstallRules(ctx context.Context, ifName string, quiet, install bool) error

	// GetProxyPort fetches the existing proxy port configured for the
	// specified listener. Used early in bootstrap to reopen proxy ports.
	GetProxyPort(listener string) uint16

	// InstallNoTrackRules is explicitly called when a pod has valid
	// "io.cilium.no-track-port" annotation.  When
	// InstallNoConntrackIptRules flag is set, a super set of v4 NOTRACK
	// rules will be automatically installed upon agent bootstrap (via
	// function addNoTrackPodTrafficRules) and this function will be
	// skipped.  When InstallNoConntrackIptRules is not set, this function
	// will be executed to install NOTRACK rules.  The rules installed by
	// this function is very specific, for now, the only user is
	// node-local-dns pods.
	InstallNoTrackRules(IP string, port uint16, ipv6 bool) error

	// See comments for InstallNoTrackRules.
	RemoveNoTrackRules(IP string, port uint16, ipv6 bool) error
}

type Loader interface {
	CallsMapPath(id uint16) string
	CustomCallsMapPath(id uint16) string
	CompileAndLoad(ctx context.Context, ep Endpoint, stats *metrics.SpanStat) error
	CompileOrLoad(ctx context.Context, ep Endpoint, stats *metrics.SpanStat) error
	ReloadDatapath(ctx context.Context, ep Endpoint, stats *metrics.SpanStat) error
	EndpointHash(cfg EndpointConfiguration) (string, error)
	Unload(ep Endpoint)
	Reinitialize(ctx context.Context, o BaseProgramOwner, deviceMTU int, iptMgr IptablesManager, p Proxy) error
}

// Endpoint provides access endpoint configuration information that is necessary
// to compile and load the datapath.
type Endpoint interface {
	EndpointConfiguration
	InterfaceName() string
	Logger(subsystem string) *logrus.Entry
	StateDir() string
	MapPath() string
}

// EndpointConfiguration provides datapath implementations a clean interface
// to access endpoint-specific configuration when configuring the datapath.
type EndpointConfiguration interface {
	CompileTimeConfiguration
	LoadTimeConfiguration
}

// LoadTimeConfiguration provides datapath implementations a clean interface
// to access endpoint-specific configuration that can be changed at load time.
type LoadTimeConfiguration interface {
	// GetID returns a locally-significant endpoint identification number.
	GetID() uint64
	// StringID returns the string-formatted version of the ID from GetID().
	StringID() string
	// GetIdentity returns a globally-significant numeric security identity.
	GetIdentity() identity.NumericIdentity

	// GetIdentityLocked returns a globally-significant numeric security
	// identity while assuming that the backing data structure is locked.
	// This function should be removed in favour of GetIdentity()
	GetIdentityLocked() identity.NumericIdentity

	IPv4Address() addressing.CiliumIPv4
	IPv6Address() addressing.CiliumIPv6
	GetNodeMAC() mac.MAC
}

// CompileTimeConfiguration provides datapath implementations a clean interface
// to access endpoint-specific configuration that can only be changed at
// compile time.
type CompileTimeConfiguration interface {
	DeviceConfiguration

	// TODO: Move this detail into the datapath
	HasIpvlanDataPath() bool
	ConntrackLocalLocked() bool

	// RequireARPPassthrough returns true if the datapath must implement
	// ARP passthrough for this endpoint
	RequireARPPassthrough() bool

	// RequireEgressProg returns true if the endpoint requires an egress
	// program attached to the InterfaceName() invoking the section
	// "to-container"
	RequireEgressProg() bool

	// RequireRouting returns true if the endpoint requires BPF routing to
	// be enabled, when disabled, routing is delegated to Linux routing
	RequireRouting() bool

	// RequireEndpointRoute returns true if the endpoint wishes to have a
	// per endpoint route installed in the host's routing table to point to
	// the endpoint's interface
	RequireEndpointRoute() bool

	// GetPolicyVerdictLogFilter returns the PolicyVerdictLogFilter for the endpoint
	GetPolicyVerdictLogFilter() uint32

	// IsHost returns true if the endpoint is the host endpoint.
	IsHost() bool
}

type DeviceConfiguration interface {
	// GetCIDRPrefixLengths fetches the lists of unique IPv6 and IPv4
	// prefix lengths used for datapath lookups, each of which is sorted
	// from longest prefix to shortest prefix. It must return more than
	// one element in each returned array.
	GetCIDRPrefixLengths() (s6, s4 []int)

	// GetOptions fetches the configurable datapath options from the owner.
	GetOptions() *option.IntOptions
}
