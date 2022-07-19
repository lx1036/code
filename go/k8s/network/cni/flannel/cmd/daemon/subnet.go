package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"

	"k8s-lx1036/k8s/network/cni/flannel/pkg/ip"
	"k8s-lx1036/k8s/network/cni/flannel/pkg/subnet"
	"k8s.io/klog/v2"
)

func recycleIPTables(network ip.IP4Net, lease *subnet.Lease) error {
	prevNetwork := ReadCIDRFromSubnetFile(*subnetFile, "FLANNEL_NETWORK")
	prevSubnet := ReadCIDRFromSubnetFile(*subnetFile, "FLANNEL_SUBNET")
	// recycle iptables rules only when network configured or subnet leased is not equal to current one.
	if prevNetwork != network && prevSubnet != lease.Subnet {
		klog.Infof(fmt.Sprintf("Current network or subnet (%v, %v) is not equal to previous one (%v, %v), trying to recycle old iptables rules",
			network, lease.Subnet, prevNetwork, prevSubnet))
		lease := &subnet.Lease{
			Subnet: prevSubnet,
		}
		if err := iptables.DeleteIP4Tables(iptables.MasqRules(prevNetwork, lease)); err != nil {
			return err
		}
	}
	return nil
}

func ReadCIDRFromSubnetFile(path string, CIDRKey string) ip.IP4Net {
	var prevCIDR ip.IP4Net
	_, err := os.Stat(path)
	if err != nil {
		klog.Errorf(fmt.Sprintf("%v", err))
		return prevCIDR
	}
	prevSubnetVals, err := godotenv.Read(path)
	if err != nil {
		klog.Errorf(fmt.Sprintf("%v", err))
		return prevCIDR
	}
	if value, ok := prevSubnetVals[CIDRKey]; ok {
		err = prevCIDR.UnmarshalJSON([]byte(value))
		if err != nil {
			klog.Errorf(fmt.Sprintf("Couldn't parse previous %s from subnet file at %s: %s", CIDRKey, path, err))
		}
	}

	return prevCIDR
}
