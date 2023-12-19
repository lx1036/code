package datapath

import (
	"net"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/cidr"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/mtu"
)

// LocalNodeConfiguration represents the configuration of the local node
type LocalNodeConfiguration struct {
	// MtuConfig is the MTU configuration of the node.
	//
	// This field is immutable at runtime. The value will not change in
	// subsequent calls to NodeConfigurationChanged().
	MtuConfig mtu.Configuration

	// AuxiliaryPrefixes is the list of auxiliary prefixes that should be
	// configured in addition to the node PodCIDR
	//
	// This field is mutable. The implementation of
	// NodeConfigurationChanged() must adjust the routes accordingly.
	AuxiliaryPrefixes []*cidr.CIDR

	// EnableIPv4 enables use of IPv4. Routing to the IPv4 allocation CIDR
	// of other nodes must be enabled.
	//
	// This field is immutable at runtime. The value will not change in
	// subsequent calls to NodeConfigurationChanged().
	EnableIPv4 bool

	// EnableIPv6 enables use of IPv6. Routing to the IPv6 allocation CIDR
	// of other nodes must be enabled.
	//
	// This field is immutable at runtime. The value will not change in
	// subsequent calls to NodeConfigurationChanged().
	EnableIPv6 bool

	// UseSingleClusterRoute enables the use of a single cluster-wide route
	// to direct traffic from the host into the Cilium datapath.  This
	// avoids the requirement to install a separate route for each node
	// CIDR and can thus improve the overhead when operating large clusters
	// with significant node event churn due to auto-scaling.
	//
	// Use of UseSingleClusterRoute must be compatible with
	// EnableAutoDirectRouting. When both are enabled, any direct node
	// route must take precedence over the cluster-wide route as per LPM
	// routing definition.
	//
	// This field is mutable. The implementation of
	// NodeConfigurationChanged() must adjust the routes accordingly.
	UseSingleClusterRoute bool

	// EnableEncapsulation enables use of encapsulation in communication
	// between nodes.
	//
	// This field is immutable at runtime. The value will not change in
	// subsequent calls to NodeConfigurationChanged().
	EnableEncapsulation bool

	// EnableAutoDirectRouting enables the use of direct routes for
	// communication between nodes if two nodes have direct L2
	// connectivity.
	//
	// EnableAutoDirectRouting must be compatible with EnableEncapsulation
	// and must provide a fallback to use encapsulation if direct routing
	// is not feasible and encapsulation is enabled.
	//
	// This field is immutable at runtime. The value will not change in
	// subsequent calls to NodeConfigurationChanged().
	EnableAutoDirectRouting bool

	// EnableLocalNodeRoute enables installation of the route which points
	// the allocation prefix of the local node. Disabling this option is
	// useful when another component is responsible for the routing of the
	// allocation CIDR IPs into Cilium endpoints.
	EnableLocalNodeRoute bool

	// EnableIPSec enables IPSec routes
	EnableIPSec bool

	// EncryptNode enables encrypting NodeIP traffic requires EnableIPSec
	EncryptNode bool

	// IPv4PodSubnets is a list of IPv4 subnets that pod IPs are assigned from
	// these are then used when encryption is enabled to configure the node
	// for encryption over these subnets at node initialization.
	IPv4PodSubnets []*net.IPNet

	// IPv6PodSubnets is a list of IPv6 subnets that pod IPs are assigned from
	// these are then used when encryption is enabled to configure the node
	// for encryption over these subnets at node initialization.
	IPv6PodSubnets []*net.IPNet
}