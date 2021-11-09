package routing

import (
	"context"
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
	}, func(ds *gobgpapi.DefinedSet) {
		currentDefinedSet = ds
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
func (controller *NetworkRoutingController) addDefaultRouteDefinedSet() error {
	var currentDefinedSet *gobgpapi.DefinedSet
	definedsetName := "defaultroutedefinedset"
	err := controller.bgpServer.ListDefinedSet(context.Background(),
		&gobgpapi.ListDefinedSetRequest{
			DefinedType: gobgpapi.DefinedType_PREFIX,
			Name:        definedsetName,
		}, func(ds *gobgpapi.DefinedSet) {
			currentDefinedSet = ds
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
