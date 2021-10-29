package config

import api "github.com/osrg/gobgp/api"

// INFO: route policy:
//  https://www.juniper.net/documentation/cn/zh/software/junos/bgp/topics/topic-map/basic-routing-policies.html
//  https://github.com/osrg/gobgp/blob/master/docs/sources/policy.md

/*
Route Policy: a way to control how BGP routes inserted to RIB(route information based) or advertised to peers.
Policy has two parts, Condition and Action.
When a policy is configured, Action is applied to routes which meet Condition before routes proceed to next step.

*/

func newApplyPolicyFromConfigStruct(c *ApplyPolicy) *api.ApplyPolicy {
	f := func(t DefaultPolicyType) api.RouteAction {
		if t == DEFAULT_POLICY_TYPE_ACCEPT_ROUTE {
			return api.RouteAction_ACCEPT
		} else if t == DEFAULT_POLICY_TYPE_REJECT_ROUTE {
			return api.RouteAction_REJECT
		}
		return api.RouteAction_NONE
	}
	applyPolicy := &api.ApplyPolicy{
		ImportPolicy: &api.PolicyAssignment{
			Direction:     api.PolicyDirection_IMPORT,
			DefaultAction: f(c.Config.DefaultImportPolicy),
		},
		ExportPolicy: &api.PolicyAssignment{
			Direction:     api.PolicyDirection_EXPORT,
			DefaultAction: f(c.Config.DefaultExportPolicy),
		},
	}

	for _, pname := range c.Config.ImportPolicyList {
		applyPolicy.ImportPolicy.Policies = append(applyPolicy.ImportPolicy.Policies, &api.Policy{Name: pname})
	}
	for _, pname := range c.Config.ExportPolicyList {
		applyPolicy.ExportPolicy.Policies = append(applyPolicy.ExportPolicy.Policies, &api.Policy{Name: pname})
	}

	return applyPolicy
}
