package types

import (
	"net"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/cidr"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/source"
)

// Node contains the nodes name, the list of addresses to this address
//
// +k8s:deepcopy-gen=true
type Node struct {
	// Name is the name of the node. This is typically the hostname of the node.
	Name string

	// Cluster is the name of the cluster the node is associated with
	Cluster string

	IPAddresses []Address

	// IPv4AllocCIDR if set, is the IPv4 address pool out of which the node
	// allocates IPs for local endpoints from
	IPv4AllocCIDR *cidr.CIDR

	// IPv4SecondaryAllocCIDRs contains additional IPv4 CIDRs from which this
	//node allocates IPs for its local endpoints from
	IPv4SecondaryAllocCIDRs []*cidr.CIDR

	// IPv6AllocCIDR if set, is the IPv6 address pool out of which the node
	// allocates IPs for local endpoints from
	IPv6AllocCIDR *cidr.CIDR

	// IPv6SecondaryAllocCIDRs contains additional IPv6 CIDRs from which this
	// node allocates IPs for its local endpoints from
	IPv6SecondaryAllocCIDRs []*cidr.CIDR

	// IPv4HealthIP if not nil, this is the IPv4 address of the
	// cilium-health endpoint located on the node.
	IPv4HealthIP net.IP

	// IPv6HealthIP if not nil, this is the IPv6 address of the
	// cilium-health endpoint located on the node.
	IPv6HealthIP net.IP

	// IPv4IngressIP if not nil, this is the IPv4 address of the
	// Ingress listener on the node.
	IPv4IngressIP net.IP

	// IPv6IngressIP if not nil, this is the IPv6 address of the
	// Ingress listener located on the node.
	IPv6IngressIP net.IP

	// ClusterID is the unique identifier of the cluster
	ClusterID int

	// Source is the source where the node configuration was generated / created.
	Source source.Source

	// Key index used for transparent encryption or 0 for no encryption
	EncryptionKey uint8

	// Node labels
	Labels map[string]string

	// NodeIdentity is the numeric identity allocated for the node
	NodeIdentity uint32

	// WireguardPubKey is the WireGuard public key of this node
	WireguardPubKey string
}
