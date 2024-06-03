package policy

import (
    "github.com/cilium/cilium/pkg/identity/cache"
    "sync"

    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/lock"
)

// SelectorCache caches identities, identity selectors, and the
// subsets of identities each selector selects.
type SelectorCache struct {
    mutex lock.RWMutex

    // idAllocator is used to allocate and release identities. It is used
    // by the NameManager to manage identities corresponding to FQDNs.
    idAllocator cache.IdentityAllocator

    // idCache contains all known identities as informed by the
    // kv-store and the local identity facility via our
    // UpdateIdentities() function.
    idCache scIdentityCache

    // map key is the string representation of the selector being cached.
    selectors map[string]identitySelector

    localIdentityNotifier identityNotifier

    // userCond is a condition variable for receiving signals
    // about addition of new elements in userNotes
    userCond *sync.Cond
    // userMutex protects userNotes and is linked to userCond
    userMutex lock.Mutex
    // userNotes holds a FIFO list of user notifications to be made
    userNotes []userNotification
}
