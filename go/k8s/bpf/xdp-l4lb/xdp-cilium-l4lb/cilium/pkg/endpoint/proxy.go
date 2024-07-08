package endpoint

import (
	"context"

	"github.com/cilium/cilium/pkg/completion"
	"github.com/cilium/cilium/pkg/policy"
	"github.com/cilium/cilium/pkg/proxy/logger"
	"github.com/cilium/cilium/pkg/revert"
)

// EndpointProxy defines any L7 proxy with which an Endpoint must interact.
type EndpointProxy interface {
	CreateOrUpdateRedirect(ctx context.Context, l4 policy.ProxyPolicy, id string, localEndpoint logger.EndpointUpdater, wg *completion.WaitGroup) (proxyPort uint16, err error, finalizeFunc revert.FinalizeFunc, revertFunc revert.RevertFunc)
	RemoveRedirect(id string, wg *completion.WaitGroup) (error, revert.FinalizeFunc, revert.RevertFunc)
	UpdateNetworkPolicy(ep logger.EndpointUpdater, vis *policy.VisibilityPolicy, policy *policy.L4Policy, ingressPolicyEnforced, egressPolicyEnforced bool, wg *completion.WaitGroup) (error, func() error)
	UseCurrentNetworkPolicy(ep logger.EndpointUpdater, policy *policy.L4Policy, wg *completion.WaitGroup)
	RemoveNetworkPolicy(ep logger.EndpointInfoSource)
}
