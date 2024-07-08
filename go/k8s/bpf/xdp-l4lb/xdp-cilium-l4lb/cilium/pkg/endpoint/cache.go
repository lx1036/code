package endpoint

import (
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/addressing"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/identity"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/mac"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/option"
)

// epInfoCache describes the set of lxcmap entries necessary to describe an Endpoint
// in the BPF maps. It is generated while holding the Endpoint lock, then used
// after releasing that lock to push the entries into the datapath.
// Functions below implement the EndpointFrontend interface with this cached information.
type epInfoCache struct {
	// revision is used by the endpoint regeneration code to determine
	// whether this cache is out-of-date wrt the underlying endpoint.
	revision uint64

	// For datapath.loader.endpoint
	epdir  string
	id     uint64
	ifName string

	// For datapath.EndpointConfiguration
	identity                               identity.NumericIdentity
	mac                                    mac.MAC
	ipv4                                   addressing.CiliumIPv4
	ipv6                                   addressing.CiliumIPv6
	conntrackLocal                         bool
	requireARPPassthrough                  bool
	requireEgressProg                      bool
	requireRouting                         bool
	requireEndpointRoute                   bool
	disableSIPVerification                 bool
	policyVerdictLogFilter                 uint32
	cidr4PrefixLengths, cidr6PrefixLengths []int
	options                                *option.IntOptions
	lxcMAC                                 mac.MAC
	ifIndex                                int

	// endpoint is used to get the endpoint's logger.
	//
	// Do NOT use this for fetching endpoint data directly; this structure
	// is intended as a safe cache of endpoint data that is assembled while
	// holding the endpoint lock, for use beyond the holding of that lock.
	// Dereferencing fields in this endpoint is not guaranteed to be safe.
	endpoint *Endpoint
}

func (ep *epInfoCache) LXCMac() mac.MAC {
	return ep.lxcMAC
}

func (ep *epInfoCache) GetNodeMAC() mac.MAC {
	return ep.mac
}

func (ep *epInfoCache) GetIfIndex() int {
	return ep.ifIndex
}

// GetID returns the endpoint's ID.
func (ep *epInfoCache) GetID() uint64 {
	return ep.id
}

// IPv4Address returns the cached IPv4 address for the endpoint.
func (ep *epInfoCache) IPv4Address() addressing.CiliumIPv4 {
	return ep.ipv4
}
