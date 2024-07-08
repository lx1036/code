package cache

import (
	"context"
	"net"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/identity"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/labels"
)

// IdentityAllocator is any type which is responsible for allocating security
// identities based of sets of labels, and caching information about identities
// locally.
type IdentityAllocator interface {
	// WaitForInitialGlobalIdentities waits for the initial set of global
	// security identities to have been received.
	WaitForInitialGlobalIdentities(context.Context) error

	// AllocateIdentity allocates an identity described by the specified labels.
	// A possible previously used numeric identity for these labels can be passed
	// in as the last parameter; identity.InvalidIdentity must be passed if no
	// previous numeric identity exists.
	AllocateIdentity(context.Context, labels.Labels, bool, identity.NumericIdentity) (*identity.Identity, bool, error)

	// Release is the reverse operation of AllocateIdentity() and releases the
	// specified identity.
	Release(context.Context, *identity.Identity, bool) (released bool, err error)

	// ReleaseSlice is the slice variant of Release().
	ReleaseSlice(context.Context, IdentityAllocatorOwner, []*identity.Identity) error

	// LookupIdentityByID returns the identity that corresponds to the given
	// labels.
	LookupIdentity(ctx context.Context, lbls labels.Labels) *identity.Identity

	// LookupIdentityByID returns the identity that corresponds to the given
	// numeric identity.
	LookupIdentityByID(ctx context.Context, id identity.NumericIdentity) *identity.Identity

	// GetIdentityCache returns the current cache of identities that the
	// allocator has allocated. The caller should not modify the resulting
	// identities by pointer.
	GetIdentityCache() IdentityCache

	// GetIdentities returns a copy of the current cache of identities.
	GetIdentities() IdentitiesModel

	// AllocateCIDRsForIPs attempts to allocate identities for a list of
	// CIDRs. If any allocation fails, all allocations are rolled back and
	// the error is returned. When an identity is freshly allocated for a
	// CIDR, it is added to the ipcache if 'newlyAllocatedIdentities' is
	// 'nil', otherwise the newly allocated identities are placed in
	// 'newlyAllocatedIdentities' and it is the caller's responsibility to
	// upsert them into ipcache by calling UpsertGeneratedIdentities().
	//
	// Upon success, the caller must also arrange for the resulting identities to
	// be released via a subsequent call to ReleaseCIDRIdentitiesByID().
	//
	// The implementation for this function currently lives in pkg/ipcache.
	AllocateCIDRsForIPs(ips []net.IP, newlyAllocatedIdentities map[string]*identity.Identity) ([]*identity.Identity, error)

	// ReleaseCIDRIdentitiesByID() is a wrapper for ReleaseSlice() that
	// also handles ipcache entries.
	ReleaseCIDRIdentitiesByID(context.Context, []identity.NumericIdentity)
}
