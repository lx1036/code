package types

import (
	"github.com/cilium/cilium/pkg/logging/logfields"
	log "github.com/sirupsen/logrus"
	"net"
	"os"

	"github.com/cilium/cilium/pkg/cidr"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/config/defaults"
)

var (
	nodeName = "localhost"
)

func SetName(name string) {
	nodeName = name
}

// GetName returns the name of the local node. The value returned was either
// previously set with SetName(), retrieved via `os.Hostname()`, or as a last
// resort is hardcoded to "localhost".
func GetName() string {
	return nodeName
}

func init() {
	// Give priority to the environment variable available in the Cilium agent
	if name := os.Getenv(defaults.EnvNodeNameSpec); name != "" {
		nodeName = name
		return
	}
	if h, err := os.Hostname(); err != nil {
		log.WithError(err).Warn("Unable to retrieve local hostname")
	} else {
		log.WithField(logfields.NodeName, h).Debug("os.Hostname() returned")
		nodeName = h
	}
}

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

	// IPv6AllocCIDR if set, is the IPv6 address pool out of which the node
	// allocates IPs for local endpoints from
	IPv6AllocCIDR *cidr.CIDR

	// IPv4HealthIP if not nil, this is the IPv4 address of the
	// cilium-health endpoint located on the node.
	IPv4HealthIP net.IP

	// IPv6HealthIP if not nil, this is the IPv6 address of the
	// cilium-health endpoint located on the node.
	IPv6HealthIP net.IP

	// ClusterID is the unique identifier of the cluster
	ClusterID int

	// Source is the source where the node configuration was generated / created.
	Source source.Source

	// Key index used for transparent encryption or 0 for no encryption
	EncryptionKey uint8

	// Node labels
	Labels map[string]string
}
