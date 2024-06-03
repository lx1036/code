package policy

import (
    "unsafe"

    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/labels"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/lock"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/policy/api"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/u8proto"
)

// L4Filter represents the policy (allowed remote sources / destinations of
// traffic) that applies at a specific L4 port/protocol combination (including
// all ports and protocols), at either ingress or egress. The policy here is
// specified in terms of selectors that are mapped to security identities via
// the selector cache.
type L4Filter struct {
    // Port is the destination port to allow. Port 0 indicates that all traffic
    // is allowed at L4.
    Port     int    `json:"port"`
    PortName string `json:"port-name,omitempty"`
    // Protocol is the L4 protocol to allow or NONE
    Protocol api.L4Proto `json:"protocol"`
    // U8Proto is the Protocol in numeric format, or 0 for NONE
    U8Proto u8proto.U8proto `json:"-"`
    // wildcard is the cached selector representing a wildcard in this filter, if any.
    // This is nil the wildcard selector in not in 'L7RulesPerSelector'.
    // When the wildcard selector is in 'L7RulesPerSelector' this is set to that
    // same selector, which can then be used as a map key to find the corresponding
    // L4-only L7 policy (which can be nil).
    wildcard CachedSelector
    // L7RulesPerSelector is a list of L7 rules per endpoint passed to the L7 proxy.
    // nil values represent cached selectors that have no L7 restriction.
    // Holds references to the cached selectors, which must be released!
    L7RulesPerSelector L7DataMap `json:"l7-rules,omitempty"`
    // L7Parser specifies the L7 protocol parser (optional). If specified as
    // an empty string, then means that no L7 proxy redirect is performed.
    L7Parser L7ParserType `json:"-"`
    // Ingress is true if filter applies at ingress; false if it applies at egress.
    Ingress bool `json:"-"`
    // The rule labels of this Filter
    DerivedFromRules labels.LabelArrayList `json:"-"`

    // This reference is circular, but it is cleaned up at Detach()
    policy unsafe.Pointer // *L4Policy
}

// L4PolicyMap is a list of L4 filters indexable by protocol/port
// key format: "port/proto"
type L4PolicyMap map[string]*L4Filter

type L4Policy struct {
    Ingress L4PolicyMap
    Egress  L4PolicyMap

    // Revision is the repository revision used to generate this policy.
    Revision uint64

    // Endpoint policies using this L4Policy
    // These are circular references, cleaned up in Detach()
    // This mutex is taken while Endpoint mutex is held, so Endpoint lock
    // MUST always be taken before this mutex.
    mutex lock.RWMutex
    users map[*EndpointPolicy]struct{}
}
