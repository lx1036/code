package policy

// selectorPolicy is a structure which contains the resolved policy for a
// particular Identity across all layers (L3, L4, and L7), with the policy
// still determined in terms of EndpointSelectors.
type selectorPolicy struct {
    // Revision is the revision of the policy repository used to generate
    // this selectorPolicy.
    Revision uint64

    // SelectorCache managing selectors in L4Policy
    SelectorCache *SelectorCache

    // L4Policy contains the computed L4 and L7 policy.
    L4Policy *L4Policy

    // IngressPolicyEnabled specifies whether this policy contains any policy
    // at ingress.
    IngressPolicyEnabled bool

    // EgressPolicyEnabled specifies whether this policy contains any policy
    // at egress.
    EgressPolicyEnabled bool
}

func newSelectorPolicy(revision uint64, selectorCache *SelectorCache) *selectorPolicy {
    return &selectorPolicy{
        Revision:      revision,
        SelectorCache: selectorCache,
    }
}

// EndpointPolicy is a structure which contains the resolved policy across all
// layers (L3, L4, and L7), distilled against a set of identities.
type EndpointPolicy struct {
    // Note that all Endpoints sharing the same identity will be
    // referring to a shared selectorPolicy!
    *selectorPolicy

    // PolicyMapState contains the state of this policy as it relates to the
    // datapath. In the future, this will be factored out of this object to
    // decouple the policy as it relates to the datapath vs. its userspace
    // representation.
    // It maps each Key to the proxy port if proxy redirection is needed.
    // Proxy port 0 indicates no proxy redirection.
    // All fields within the Key and the proxy port must be in host byte-order.
    // Must only be accessed with PolicyOwner (aka Endpoint) lock taken.
    PolicyMapState MapState

    // policyMapChanges collects pending changes to the PolicyMapState
    policyMapChanges MapChanges

    // PolicyOwner describes any type which consumes this EndpointPolicy object.
    PolicyOwner PolicyOwner
}

func NewEndpointPolicy(repo *Repository) *EndpointPolicy {
    return &EndpointPolicy{
        selectorPolicy: newSelectorPolicy(0, repo.GetSelectorCache()),
    }
}
