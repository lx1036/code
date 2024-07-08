package ipcache

import (
	"net"
	"time"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/controller"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/identity"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/identity/cache"
	ipcacheTypes "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/ipcache/types"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/lock"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/policy"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/source"
)

// INFO: ipcache BPF map 主要是用来？？？

var (
	// IPIdentityCache caches the mapping of endpoint IPs to their corresponding
	// security identities across the entire cluster in which this instance of
	// Cilium is running.
	IPIdentityCache = NewIPCache()
)

// Configuration is init-time configuration for the IPCache.
type Configuration struct {
	// Accessors to other subsystems, provided by the daemon
	cache.IdentityAllocator
	ipcacheTypes.PolicyHandler
	ipcacheTypes.DatapathHandler
}

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
	mutex             lock.SemaphoredMutex
	ipToIdentityCache map[string]Identity
	identityToIPCache map[identity.NumericIdentity]map[string]struct{}
	ipToHostIPCache   map[string]IPKeyPair
	ipToK8sMetadata   map[string]K8sMetadata

	listeners []IPIdentityMappingListener

	// controllers manages the async controllers for this IPCache
	controllers *controller.Manager

	// needNamedPorts is initially 'false', but will be changd to 'true' when the
	// clusterwide named port mappings are needed for network policy computation
	// for the first time. This avoids the overhead of maintaining 'namedPorts' map
	// when it is known not to be needed.
	// Protected by 'mutex'.
	needNamedPorts bool

	// namedPorts is a collection of all named ports in the cluster. This is needed
	// only if an egress policy refers to a port by name.
	// This map is returned to users so all updates must be made into a fresh map that
	// is then swapped in place while 'mutex' is being held.
	namedPorts policy.NamedPortMultiMap

	// k8sSyncedChecker knows how to check for whether the K8s watcher cache
	// has been fully synced.
	k8sSyncedChecker k8sSyncedChecker

	// Configuration provides pointers towards other agent components that
	// the IPCache relies upon at runtime.
	*Configuration

	// metadata is the ipcache identity metadata map, which maps IPs to labels.
	metadata *metadata

	// deferredPrefixRelease is a queue for garbage collecting old
	// references to identities and removing the corresponding IPCache
	// entries if unused.
	deferredPrefixRelease *asyncPrefixReleaser
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

// NewIPCache returns a new IPCache with the mappings of endpoint IP to security
// identity (and vice-versa) initialized.
func NewIPCache(c *Configuration) *IPCache {
	ipc := &IPCache{
		mutex:             lock.NewSemaphoredMutex(),
		ipToIdentityCache: map[string]Identity{},
		identityToIPCache: map[identity.NumericIdentity]map[string]struct{}{},
		ipToHostIPCache:   map[string]IPKeyPair{},
		ipToK8sMetadata:   map[string]K8sMetadata{},
		controllers:       controller.NewManager(),
		namedPorts:        nil,
		metadata:          newMetadata(),
		Configuration:     c,
	}
	ipc.deferredPrefixRelease = newAsyncPrefixReleaser(ipc, 1*time.Millisecond)
	return ipc
}
