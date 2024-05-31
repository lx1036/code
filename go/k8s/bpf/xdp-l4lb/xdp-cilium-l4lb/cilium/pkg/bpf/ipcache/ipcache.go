package ipcache

import (
	"github.com/cilium/cilium/pkg/identity"
	"net"
)

// INFO: ipcache BPF map 主要是用来？？？

var (
	// IPIdentityCache caches the mapping of endpoint IPs to their corresponding
	// security identities across the entire cluster in which this instance of
	// Cilium is running.
	IPIdentityCache = NewIPCache()
)

// Identity is the identity representation of an IP<->Identity cache.
type Identity struct {
	// ID is the numeric identity
	ID identity.NumericIdentity

	// Source is the source of the identity in the cache
	Source source.Source

	// shadowed determines if another entry overlaps with this one.
	// Shadowed identities are not propagated to listeners by default.
	// Most commonly set for Identity with Source = source.Generated when
	// a pod IP (other source) has the same IP.
	shadowed bool
}

// IPCache is a collection of mappings:
//   - mapping of endpoint IP or CIDR to security identities of all endpoints
//     which are part of the same cluster, and vice-versa
//   - mapping of endpoint IP or CIDR to host IP (maybe nil)
type IPCache struct {
}

// NewIPCache returns a new IPCache with the mappings of endpoint IP to security
// identity (and vice-versa) initialized.
func NewIPCache() *IPCache {
	return &IPCache{}
}

// UpdateOrInsert adds / updates the provided IP (endpoint or CIDR prefix) and identity
// into the IPCache.
//
// Returns false if the entry is not owned by the self declared source, i.e.
// returns false if the kubernetes layer is trying to upsert an entry now
// managed by the kvstore layer. See source.AllowOverwrite() for rules on
// ownership. hostIP is the location of the given IP. It is optional (may be
// nil) and is propagated to the listeners. k8sMeta contains Kubernetes-specific
// metadata such as pod namespace and pod name belonging to the IP (may be nil).
func (ipc *IPCache) UpdateOrInsert(ip string, hostIP net.IP, hostKey uint8, k8sMeta *K8sMetadata,
	newIdentity Identity) (updated bool, namedPortsChanged bool) {

}
