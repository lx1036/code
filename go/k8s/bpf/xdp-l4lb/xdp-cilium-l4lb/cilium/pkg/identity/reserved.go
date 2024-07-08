package identity

import (
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/labels"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/lock"
)

var (
	// cacheMU protects the following map.
	cacheMU lock.RWMutex
	// ReservedIdentityCache that maps all reserved identities from their
	// numeric identity to their corresponding identity.
	reservedIdentityCache = map[NumericIdentity]*Identity{}
)

func init() {
	iterateReservedIdentityLabels(AddReservedIdentityWithLabels)
}

// IterateReservedIdentities iterates over all reserved identities and
// executes the given function for each identity.
func IterateReservedIdentities(f func(_ NumericIdentity, _ *Identity)) {
	cacheMU.RLock()
	defer cacheMU.RUnlock()
	for ni, identity := range reservedIdentityCache {
		f(ni, identity)
	}
}

// AddReservedIdentityWithLabels is the same as AddReservedIdentity but accepts
// multiple labels.
func AddReservedIdentityWithLabels(ni NumericIdentity, lbls labels.Labels) {
	identity := NewIdentity(ni, lbls)
	cacheMU.Lock()
	reservedIdentityCache[ni] = identity
	cacheMU.Unlock()
}
