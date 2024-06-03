package policy

import (
    "github.com/cilium/cilium/pkg/policy/api"
    cilium "github.com/cilium/proxy/go/cilium/api"

    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/eventqueue"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/lock"
)

// Repository is a list of policy rules which in combination form the security
// policy. A policy repository can be
type Repository struct {
    // Mutex protects the whole policy tree
    Mutex lock.RWMutex
    rules ruleSlice

    // revision is the revision of the policy repository. It will be
    // incremented whenever the policy repository is changed.
    // Always positive (>0).
    revision uint64

    // RepositoryChangeQueue is a queue which serializes changes to the policy
    // repository.
    RepositoryChangeQueue *eventqueue.EventQueue

    // RuleReactionQueue is a queue which serializes the resultant events that
    // need to occur after updating the state of the policy repository. This
    // can include queueing endpoint regenerations, policy revision increments
    // for endpoints, etc.
    RuleReactionQueue *eventqueue.EventQueue

    // SelectorCache tracks the selectors used in the policies
    // resolved from the repository.
    selectorCache *SelectorCache

    // PolicyCache tracks the selector policies created from this repo
    policyCache *PolicyCache

    certManager CertificateManager

    getEnvoyHTTPRules func(CertificateManager, *api.L7Rules, string) (*cilium.HttpNetworkPolicyRules, bool)
}
