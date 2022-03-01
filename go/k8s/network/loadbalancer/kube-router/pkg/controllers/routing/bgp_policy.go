package routing

import (
	"context"
	"fmt"
	"k8s-lx1036/k8s/network/loadbalancer/kube-router/pkg/utils"
	"k8s.io/apimachinery/pkg/labels"
	"reflect"
	
	gobgpapi "github.com/osrg/gobgp/api"
	"k8s.io/klog/v2"
)

// INFO: BGP route policy: https://github.com/osrg/gobgp/blob/master/docs/sources/policy.md

// AddPolicies adds BGP import and export policies
func (controller *NetworkRoutingController) AddPolicies() error {
	err := controller.syncPodCidrDefinedSet()
	if err != nil {
		klog.Errorf("Failed to add `podcidrdefinedset` defined set: %s", err)
	}
	
	err = controller.syncServiceVIPsDefinedSet()
	if err != nil {
		klog.Errorf("Failed to add `servicevipsdefinedset` defined set: %s", err)
	}
	
	err = controller.syncDefaultRouteDefinedSet()
	if err != nil {
		klog.Errorf("Failed to add `defaultroutedefinedset` defined set: %s", err)
	}
	
	iBGPPeerCIDRs, err := controller.addiBGPPeersDefinedSet()
	if err != nil {
		klog.Errorf("Failed to add `iBGPpeerset` defined set: %s", err)
	}
	
	externalBGPPeerCIDRs, err := controller.addExternalBGPPeersDefinedSet()
	if err != nil {
		klog.Errorf("Failed to add `externalpeerset` defined set: %s", err)
	}
	
	err = controller.addAllBGPPeersDefinedSet(iBGPPeerCIDRs, externalBGPPeerCIDRs)
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
func (controller *NetworkRoutingController) syncPodCidrDefinedSet() error {
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
func (controller *NetworkRoutingController) syncServiceVIPsDefinedSet() error {
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
	advIps, _, _ := controller.getAllVIPs()
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
func (controller *NetworkRoutingController) syncDefaultRouteDefinedSet() error {
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

func (controller *NetworkRoutingController) addiBGPPeersDefinedSet() ([]string, error) {
	iBGPPeerCIDRs := make([]string, 0)
	if !controller.enableIBGP {
		return iBGPPeerCIDRs, nil
	}
	
	definedsetName := "iBGPpeerset"
	// Get the current list of the nodes from the local cache
	nodes, err := controller.nodeLister.List(labels.Everything())
	if err != nil {
		return nil, err
	}
	for _, node := range nodes {
		nodeIP, err := utils.GetNodeIP(node)
		if err != nil {
			klog.Errorf("Failed to find a node IP and therefore cannot add internal BGP Peer: %v", err)
			continue
		}
		iBGPPeerCIDRs = append(iBGPPeerCIDRs, fmt.Sprintf("%s/32", nodeIP.String()))
	}
	
	var currentDefinedSet *gobgpapi.DefinedSet
	err = controller.bgpServer.ListDefinedSet(context.Background(),
		&gobgpapi.ListDefinedSetRequest{
			DefinedType: gobgpapi.DefinedType_NEIGHBOR,
			Name:        definedsetName,
		},
		func(ds *gobgpapi.DefinedSet) {
			currentDefinedSet = ds
		})
	if err != nil {
		return iBGPPeerCIDRs, err
	}
	if currentDefinedSet == nil {
		iBGPPeerNS := &gobgpapi.DefinedSet{
			DefinedType: gobgpapi.DefinedType_NEIGHBOR,
			Name:        definedsetName,
			List:        iBGPPeerCIDRs,
		}
		err = controller.bgpServer.AddDefinedSet(context.Background(), &gobgpapi.AddDefinedSetRequest{
			DefinedSet: iBGPPeerNS,
		})
		return iBGPPeerCIDRs, err
	}
	
	if reflect.DeepEqual(iBGPPeerCIDRs, currentDefinedSet.List) {
		return iBGPPeerCIDRs, nil
	}
	toAdd := make([]string, 0)
	toDelete := make([]string, 0)
	for _, prefix := range iBGPPeerCIDRs {
		add := true
		for _, currentPrefix := range currentDefinedSet.List {
			if prefix == currentPrefix {
				add = false
			}
		}
		if add {
			toAdd = append(toAdd, prefix)
		}
	}
	for _, currentPrefix := range currentDefinedSet.List {
		shouldDelete := true
		for _, prefix := range iBGPPeerCIDRs {
			if currentPrefix == prefix {
				shouldDelete = false
			}
		}
		if shouldDelete {
			toDelete = append(toDelete, currentPrefix)
		}
	}
	err = controller.bgpServer.AddDefinedSet(context.Background(), &gobgpapi.AddDefinedSetRequest{
		DefinedSet: &gobgpapi.DefinedSet{
			DefinedType: gobgpapi.DefinedType_NEIGHBOR,
			Name:        definedsetName,
			List:        toAdd,
		},
	})
	if err != nil {
		return iBGPPeerCIDRs, err
	}
	err = controller.bgpServer.DeleteDefinedSet(context.Background(),
		&gobgpapi.DeleteDefinedSetRequest{
			DefinedSet: &gobgpapi.DefinedSet{
				DefinedType: gobgpapi.DefinedType_NEIGHBOR,
				Name:        definedsetName,
				List:        toDelete,
			},
			All: false,
		})
	if err != nil {
		return iBGPPeerCIDRs, err
	}
	
	return iBGPPeerCIDRs, nil
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
func (controller *NetworkRoutingController) addAllBGPPeersDefinedSet(iBGPPeerCIDRs, externalBGPPeerCIDRs []string) error {
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
	allBgpPeers := append(externalBGPPeerCIDRs, iBGPPeerCIDRs...)
	if currentDefinedSet == nil {
		allPeerNS := &gobgpapi.DefinedSet{
			DefinedType: gobgpapi.DefinedType_NEIGHBOR,
			Name:        definedsetName,
			List:        allBgpPeers,
		}
		return controller.bgpServer.AddDefinedSet(context.Background(), &gobgpapi.AddDefinedSetRequest{DefinedSet: allPeerNS})
	}
	
	toAdd := make([]string, 0)
	toDelete := make([]string, 0)
	for _, peer := range allBgpPeers {
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
		for _, peer := range allBgpPeers {
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
	
	if controller.enableIBGP {
		actions := gobgpapi.Actions{
			RouteAction: gobgpapi.RouteAction_ACCEPT,
		}
		if controller.overrideNextHop {
			actions.Nexthop = &gobgpapi.NexthopAction{Self: true}
		}
		// statement to represent the export policy to permit advertising node's pod CIDR
		statements = append(statements,
			&gobgpapi.Statement{
				Conditions: &gobgpapi.Conditions{
					PrefixSet: &gobgpapi.MatchSet{
						MatchType: gobgpapi.MatchType_ANY,
						Name:      "podcidrdefinedset",
					},
					NeighborSet: &gobgpapi.MatchSet{
						MatchType: gobgpapi.MatchType_ANY,
						Name:      "iBGPpeerset",
					},
				},
				Actions: &actions,
			})
	}
	
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
	
	definition := gobgpapi.Policy{
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
		err = controller.bgpServer.AddPolicy(context.Background(), &gobgpapi.AddPolicyRequest{Policy: &definition})
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
				Policies:      []*gobgpapi.Policy{&definition},
				DefaultAction: gobgpapi.RouteAction_REJECT,
			}})
		if err != nil {
			return fmt.Errorf(fmt.Sprintf("Failed to add policy assignment: %v", err))
		}
	}
	
	return nil
}
