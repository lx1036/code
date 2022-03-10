package routing

import (
	"context"
	"fmt"
	"reflect"

	gobgpapi "github.com/osrg/gobgp/v3/api"
	"k8s.io/klog/v2"
)

// INFO: BGP route policy: https://github.com/osrg/gobgp/blob/master/docs/sources/policy.md

// AddPolicies adds BGP import and export policies
func (controller *NetworkRoutingController) AddPolicies() error {
	err := controller.addPodCidrDefinedSet()
	if err != nil {
		klog.Errorf("Failed to add `podcidrdefinedset` defined set: %s", err)
	}

	err = controller.addServiceVIPsDefinedSet()
	if err != nil {
		klog.Errorf("Failed to add `servicevipsdefinedset` defined set: %s", err)
	}

	err = controller.addDefaultRouteDefinedSet()
	if err != nil {
		klog.Errorf("Failed to add `defaultroutedefinedset` defined set: %s", err)
	}

	externalBGPPeerCIDRs, err := controller.addExternalBGPPeersDefinedSet()
	if err != nil {
		klog.Errorf("Failed to add `externalpeerset` defined set: %s", err)
	}

	err = controller.addAllBGPPeersDefinedSet(externalBGPPeerCIDRs)
	if err != nil {
		klog.Errorf("Failed to add `allpeerset` defined set: %s", err)
	}

	err = controller.addExportPolicies()
	if err != nil {
		return err
	}

	err = controller.addImportPolicies()
	if err != nil {
		return err
	}

	return nil
}

// create a defined set to represent just the pod CIDR associated with the node
func (controller *NetworkRoutingController) addPodCidrDefinedSet() error {
	var currentDefinedSet *gobgpapi.DefinedSet
	definedsetName := "podcidrdefinedset"
	err := controller.bgpServer.ListDefinedSet(context.TODO(), &gobgpapi.ListDefinedSetRequest{
		DefinedType: gobgpapi.DefinedType_PREFIX,
		Name:        definedsetName,
	}, func(definedSet *gobgpapi.DefinedSet) {
		currentDefinedSet = definedSet
	})
	if err != nil {
		return err
	}

	if currentDefinedSet == nil {
		_, mask, err := controller.splitPodCidr()
		if err != nil {
			return err
		}

		return controller.bgpServer.AddDefinedSet(context.TODO(), &gobgpapi.AddDefinedSetRequest{
			DefinedSet: &gobgpapi.DefinedSet{
				DefinedType: gobgpapi.DefinedType_PREFIX,
				Name:        definedsetName,
				Prefixes: []*gobgpapi.Prefix{
					{
						IpPrefix:      controller.podCidr,
						MaskLengthMin: uint32(mask),
						MaskLengthMax: uint32(mask),
					},
				},
			},
		})
	}

	return nil
}

// create a defined set to represent all the advertisable IP associated with the services
func (controller *NetworkRoutingController) addServiceVIPsDefinedSet() error {
	var currentDefinedSet *gobgpapi.DefinedSet
	definedsetName := "servicevipsdefinedset"
	err := controller.bgpServer.ListDefinedSet(context.Background(), &gobgpapi.ListDefinedSetRequest{
		DefinedType: gobgpapi.DefinedType_PREFIX,
		Name:        definedsetName,
	}, func(definedSet *gobgpapi.DefinedSet) {
		currentDefinedSet = definedSet
	})
	if err != nil {
		return err
	}

	advIPPrefixList := make([]*gobgpapi.Prefix, 0)
	advIps, _, _ := controller.getAllActiveVIPs()
	for _, ip := range advIps {
		advIPPrefixList = append(advIPPrefixList,
			&gobgpapi.Prefix{
				IpPrefix:      ip + "/32",
				MaskLengthMin: 32,
				MaskLengthMax: 32,
			})
	}
	if currentDefinedSet == nil {
		return controller.bgpServer.AddDefinedSet(context.Background(), &gobgpapi.AddDefinedSetRequest{
			DefinedSet: &gobgpapi.DefinedSet{
				DefinedType: gobgpapi.DefinedType_PREFIX,
				Name:        definedsetName,
				Prefixes:    advIPPrefixList,
			},
		})
	}

	if reflect.DeepEqual(advIPPrefixList, currentDefinedSet.Prefixes) {
		return nil
	}

	// sync DefinedSet
	toAdd := make([]*gobgpapi.Prefix, 0)
	toDelete := make([]*gobgpapi.Prefix, 0)
	for _, prefix := range advIPPrefixList {
		add := true
		for _, currentPrefix := range currentDefinedSet.Prefixes {
			if currentPrefix.IpPrefix == prefix.IpPrefix {
				add = false
			}
		}
		if add {
			toAdd = append(toAdd, prefix)
		}
	}
	for _, currentPrefix := range currentDefinedSet.Prefixes {
		shouldDelete := true
		for _, prefix := range advIPPrefixList {
			if currentPrefix.IpPrefix == prefix.IpPrefix {
				shouldDelete = false
			}
		}
		if shouldDelete {
			toDelete = append(toDelete, currentPrefix)
		}
	}
	err = controller.bgpServer.AddDefinedSet(context.Background(),
		&gobgpapi.AddDefinedSetRequest{
			DefinedSet: &gobgpapi.DefinedSet{
				DefinedType: gobgpapi.DefinedType_PREFIX,
				Name:        definedsetName,
				Prefixes:    toAdd,
			},
		})
	if err != nil {
		return err
	}
	err = controller.bgpServer.DeleteDefinedSet(context.Background(),
		&gobgpapi.DeleteDefinedSetRequest{
			DefinedSet: &gobgpapi.DefinedSet{
				DefinedType: gobgpapi.DefinedType_PREFIX,
				Name:        definedsetName,
				Prefixes:    toDelete,
			},
			All: false,
		})
	if err != nil {
		return err
	}

	return nil
}

// create a defined set to represent just the host default route
func (controller *NetworkRoutingController) addDefaultRouteDefinedSet() error {
	var currentDefinedSet *gobgpapi.DefinedSet
	definedsetName := "defaultroutedefinedset"
	err := controller.bgpServer.ListDefinedSet(context.Background(),
		&gobgpapi.ListDefinedSetRequest{
			DefinedType: gobgpapi.DefinedType_PREFIX,
			Name:        definedsetName,
		}, func(definedSet *gobgpapi.DefinedSet) {
			currentDefinedSet = definedSet
		})
	if err != nil {
		return err
	}

	if currentDefinedSet == nil {
		cidrLen := 0
		return controller.bgpServer.AddDefinedSet(context.Background(), &gobgpapi.AddDefinedSetRequest{
			DefinedSet: &gobgpapi.DefinedSet{
				DefinedType: gobgpapi.DefinedType_PREFIX,
				Name:        definedsetName,
				Prefixes: []*gobgpapi.Prefix{
					{
						IpPrefix:      "0.0.0.0/0",
						MaskLengthMin: uint32(cidrLen),
						MaskLengthMax: uint32(cidrLen),
					},
				},
			},
		})
	}

	return nil
}

func (controller *NetworkRoutingController) addExternalBGPPeersDefinedSet() ([]string, error) {
	var currentDefinedSet *gobgpapi.DefinedSet
	definedsetName := "externalpeerset"
	externalBgpPeers := make([]string, 0)
	externalBGPPeerCIDRs := make([]string, 0)
	err := controller.bgpServer.ListDefinedSet(context.Background(), &gobgpapi.ListDefinedSetRequest{
		DefinedType: gobgpapi.DefinedType_NEIGHBOR,
		Name:        definedsetName,
	}, func(ds *gobgpapi.DefinedSet) {
		currentDefinedSet = ds
	})
	if err != nil {
		return externalBGPPeerCIDRs, err
	}

	if len(controller.globalPeerRouters) > 0 {
		for _, peer := range controller.globalPeerRouters {
			externalBgpPeers = append(externalBgpPeers, peer.Conf.NeighborAddress)
		}
	}
	if len(externalBgpPeers) == 0 {
		return externalBGPPeerCIDRs, nil
	}
	for _, peer := range externalBgpPeers {
		externalBGPPeerCIDRs = append(externalBGPPeerCIDRs, peer+"/32")
	}

	if currentDefinedSet == nil {
		err = controller.bgpServer.AddDefinedSet(context.Background(), &gobgpapi.AddDefinedSetRequest{
			DefinedSet: &gobgpapi.DefinedSet{
				DefinedType: gobgpapi.DefinedType_NEIGHBOR,
				Name:        definedsetName,
				List:        externalBGPPeerCIDRs,
			},
		})
		return externalBGPPeerCIDRs, err
	}

	return externalBGPPeerCIDRs, nil
}

// a slice of all peers is used as a match condition for reject statement of servicevipsdefinedset import policy
func (controller *NetworkRoutingController) addAllBGPPeersDefinedSet(externalBGPPeerCIDRs []string) error {
	var currentDefinedSet *gobgpapi.DefinedSet
	definedsetName := "allpeerset"
	err := controller.bgpServer.ListDefinedSet(context.Background(),
		&gobgpapi.ListDefinedSetRequest{DefinedType: gobgpapi.DefinedType_NEIGHBOR, Name: definedsetName},
		func(ds *gobgpapi.DefinedSet) {
			currentDefinedSet = ds
		})
	if err != nil {
		return err
	}
	if currentDefinedSet == nil {
		allPeerNS := &gobgpapi.DefinedSet{
			DefinedType: gobgpapi.DefinedType_NEIGHBOR,
			Name:        definedsetName,
			List:        externalBGPPeerCIDRs,
		}
		return controller.bgpServer.AddDefinedSet(context.Background(), &gobgpapi.AddDefinedSetRequest{DefinedSet: allPeerNS})
	}

	toAdd := make([]string, 0)
	toDelete := make([]string, 0)
	for _, peer := range externalBGPPeerCIDRs {
		add := true
		for _, currentPeer := range currentDefinedSet.List {
			if peer == currentPeer {
				add = false
			}
		}
		if add {
			toAdd = append(toAdd, peer)
		}
	}
	for _, currentPeer := range currentDefinedSet.List {
		shouldDelete := true
		for _, peer := range externalBGPPeerCIDRs {
			if peer == currentPeer {
				shouldDelete = false
			}
		}
		if shouldDelete {
			toDelete = append(toDelete, currentPeer)
		}
	}
	allPeerNS := &gobgpapi.DefinedSet{
		DefinedType: gobgpapi.DefinedType_NEIGHBOR,
		Name:        definedsetName,
		List:        toAdd,
	}
	err = controller.bgpServer.AddDefinedSet(context.Background(), &gobgpapi.AddDefinedSetRequest{DefinedSet: allPeerNS})
	if err != nil {
		return err
	}
	allPeerNS = &gobgpapi.DefinedSet{
		DefinedType: gobgpapi.DefinedType_NEIGHBOR,
		Name:        definedsetName,
		List:        toDelete,
	}
	err = controller.bgpServer.DeleteDefinedSet(context.Background(),
		&gobgpapi.DeleteDefinedSetRequest{DefinedSet: allPeerNS, All: false})
	if err != nil {
		return err
	}
	return nil
}

// BGP export policies are added so that following conditions are met:
//
// - by default export of all routes from the RIB to the neighbour's is denied, and explicitly statements are added
//   to permit the desired routes to be exported
// - each node is allowed to advertise its assigned pod CIDR's to all of its iBGP peer neighbours with same
//   ASN if --enable-ibgp=true
// - each node is allowed to advertise its assigned pod CIDR's to all of its external BGP peer neighbours
//   only if --advertise-pod-cidr flag is set to true
// - each node is NOT allowed to advertise its assigned pod CIDR's to all of its external BGP peer neighbours
//   only if --advertise-pod-cidr flag is set to false
// - each node is allowed to advertise service VIP's (cluster ip, load balancer ip, external IP) ONLY to external
//   BGP peers
// - each node is NOT allowed to advertise service VIP's (cluster ip, load balancer ip, external IP) to
//   iBGP peers
// - an option to allow overriding the next-hop-address with the outgoing ip for external bgp peers
func (controller *NetworkRoutingController) addExportPolicies() error {
	const policyName = "kube_router_export"
	statements := make([]*gobgpapi.Statement, 0)
	var bgpActions gobgpapi.Actions

	if len(controller.globalPeerRouters) > 0 {
		bgpActions.RouteAction = gobgpapi.RouteAction_ACCEPT
		if controller.overrideNextHop {
			bgpActions.Nexthop = &gobgpapi.NexthopAction{Self: true}
		}

		// set BGP communities for the routes advertised to peers for VIPs
		if len(controller.nodeCommunities) > 0 {
			bgpActions.Community = &gobgpapi.CommunityAction{
				ActionType:  gobgpapi.CommunityActionType_COMMUNITY_ADD,
				Communities: controller.nodeCommunities,
			}
		}

		// statement to represent the export policy to permit advertising cluster IP's
		// only to the global BGP peer or node specific BGP peer
		statements = append(statements, &gobgpapi.Statement{
			Conditions: &gobgpapi.Conditions{
				PrefixSet: &gobgpapi.MatchSet{
					MatchType: gobgpapi.MatchType_ANY,
					Name:      "servicevipsdefinedset",
				},
				NeighborSet: &gobgpapi.MatchSet{
					MatchType: gobgpapi.MatchType_ANY,
					Name:      "externalpeerset",
				},
			},
			Actions: &bgpActions,
		})

		if controller.advertisePodCidr {
			actions := gobgpapi.Actions{
				RouteAction: gobgpapi.RouteAction_ACCEPT,
			}
			// set BGP communities for the routes advertised to peers for the pod network
			if len(controller.nodeCommunities) > 0 {
				actions.Community = &gobgpapi.CommunityAction{
					ActionType:  gobgpapi.CommunityActionType_COMMUNITY_ADD,
					Communities: controller.nodeCommunities,
				}
			}
			if controller.overrideNextHop {
				actions.Nexthop = &gobgpapi.NexthopAction{Self: true}
			}
			statements = append(statements, &gobgpapi.Statement{
				Conditions: &gobgpapi.Conditions{
					PrefixSet: &gobgpapi.MatchSet{
						MatchType: gobgpapi.MatchType_ANY,
						Name:      "podcidrdefinedset",
					},
					NeighborSet: &gobgpapi.MatchSet{
						MatchType: gobgpapi.MatchType_ANY,
						Name:      "externalpeerset",
					},
				},
				Actions: &actions,
			})
		}
	}

	policy := gobgpapi.Policy{
		Name:       policyName,
		Statements: statements,
	}
	policyAlreadyExists := false
	err := controller.bgpServer.ListPolicy(context.Background(), &gobgpapi.ListPolicyRequest{}, func(existingPolicy *gobgpapi.Policy) {
		if existingPolicy.Name == policyName {
			policyAlreadyExists = true
		}
	})
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Failed to verify if kube-router BGP export policy exists: %v", err))
	}
	if !policyAlreadyExists {
		err = controller.bgpServer.AddPolicy(context.Background(), &gobgpapi.AddPolicyRequest{Policy: &policy})
		if err != nil {
			return fmt.Errorf(fmt.Sprintf("Failed to add policy: %v", err))
		}
	}

	policyAssignmentExists := false
	err = controller.bgpServer.ListPolicyAssignment(context.Background(), &gobgpapi.ListPolicyAssignmentRequest{
		Name:      "global",
		Direction: gobgpapi.PolicyDirection_EXPORT,
	}, func(existingPolicyAssignment *gobgpapi.PolicyAssignment) {
		for _, policy := range existingPolicyAssignment.Policies {
			if policy.Name == policyName {
				policyAssignmentExists = true
			}
		}
	})
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Failed to verify if kube-router BGP export policy assignment exists: %v", err))
	}
	if !policyAssignmentExists {
		err = controller.bgpServer.AddPolicyAssignment(context.Background(),
			&gobgpapi.AddPolicyAssignmentRequest{Assignment: &gobgpapi.PolicyAssignment{
				Name:          "global",
				Direction:     gobgpapi.PolicyDirection_EXPORT,
				Policies:      []*gobgpapi.Policy{&policy},
				DefaultAction: gobgpapi.RouteAction_REJECT,
			}})
		if err != nil {
			return fmt.Errorf(fmt.Sprintf("Failed to add policy assignment: %v", err))
		}
	}

	return nil
}

// BGP import policies are added so that the following conditions are met:
// - do not import Service VIPs advertised from any peers, instead each kube-router originates and injects
//   Service VIPs into local rib.
func (controller *NetworkRoutingController) addImportPolicies() error {
	statements := make([]*gobgpapi.Statement, 0)
	actions := gobgpapi.Actions{
		RouteAction: gobgpapi.RouteAction_REJECT,
	}
	statements = append(statements, &gobgpapi.Statement{
		Conditions: &gobgpapi.Conditions{
			PrefixSet: &gobgpapi.MatchSet{
				MatchType: gobgpapi.MatchType_ANY,
				Name:      "servicevipsdefinedset",
			},
			NeighborSet: &gobgpapi.MatchSet{
				MatchType: gobgpapi.MatchType_ANY,
				Name:      "allpeerset",
			},
		},
		Actions: &actions,
	}, &gobgpapi.Statement{
		Conditions: &gobgpapi.Conditions{
			PrefixSet: &gobgpapi.MatchSet{
				MatchType: gobgpapi.MatchType_ANY,
				Name:      "defaultroutedefinedset",
			},
			NeighborSet: &gobgpapi.MatchSet{
				MatchType: gobgpapi.MatchType_ANY,
				Name:      "allpeerset",
			},
		},
		Actions: &actions,
	})
	policyAlreadyExists := false
	err := controller.bgpServer.ListPolicy(context.Background(), &gobgpapi.ListPolicyRequest{}, func(existingPolicy *gobgpapi.Policy) {
		if existingPolicy.Name == "kube_router_import" {
			policyAlreadyExists = true
		}
	})
	if err != nil {
		return fmt.Errorf("Failed to verify if kube-router BGP import policy exists: " + err.Error())
	}
	policy := gobgpapi.Policy{
		Name:       "kube_router_import",
		Statements: statements,
	}
	if !policyAlreadyExists {
		err = controller.bgpServer.AddPolicy(context.Background(), &gobgpapi.AddPolicyRequest{Policy: &policy})
		if err != nil {
			return fmt.Errorf("Failed to add policy: " + err.Error())
		}
	}

	policyAssignmentExists := false
	err = controller.bgpServer.ListPolicyAssignment(context.Background(),
		&gobgpapi.ListPolicyAssignmentRequest{Name: "global", Direction: gobgpapi.PolicyDirection_IMPORT},
		func(existingPolicyAssignment *gobgpapi.PolicyAssignment) {
			for _, p := range existingPolicyAssignment.Policies {
				if p.Name == "kube_router_import" {
					policyAssignmentExists = true
				}
			}
		})
	if err != nil {
		return fmt.Errorf("Failed to verify if kube-router BGP import policy assignment exists: " + err.Error())
	}
	if !policyAssignmentExists {
		policyAssignment := gobgpapi.PolicyAssignment{
			Name:          "global",
			Direction:     gobgpapi.PolicyDirection_IMPORT,
			Policies:      []*gobgpapi.Policy{&policy},
			DefaultAction: gobgpapi.RouteAction_ACCEPT,
		}
		err = controller.bgpServer.AddPolicyAssignment(context.Background(),
			&gobgpapi.AddPolicyAssignmentRequest{Assignment: &policyAssignment})
		if err != nil {
			return fmt.Errorf("Failed to add policy assignment: " + err.Error())
		}
	}

	return nil
}
