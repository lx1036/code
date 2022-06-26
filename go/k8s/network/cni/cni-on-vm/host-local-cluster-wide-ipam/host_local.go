package main

import (
	"context"

	cnitypes "github.com/containernetworking/cni/pkg/types"
	"k8s-lx1036/k8s/network/cni/cni-on-vm/host-local-cluster-wide-ipam/pkg/allocator"
	"k8s-lx1036/k8s/network/cni/cni-on-vm/host-local-cluster-wide-ipam/pkg/store/kubernetes"

	"github.com/containernetworking/cni/pkg/skel"
	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/cni/pkg/version"
	bv "github.com/containernetworking/plugins/pkg/utils/buildversion"
)

func main() {
	skel.PluginMain(cmdAdd, cmdCheck, cmdDel, version.All, bv.BuildString("host-local"))
}

func cmdAdd(args *skel.CmdArgs) error {
	ipamConf, confVersion, err := allocator.LoadIPAMConfig(args.StdinData, args.Args)
	if err != nil {
		return err
	}

	result := &current.Result{CNIVersion: current.ImplementedSpecVersion}
	if ipamConf.ResolvConf != "" {
		dns, err := parseResolvConf(ipamConf.ResolvConf)
		if err != nil {
			return err
		}
		result.DNS = *dns
	}

	ctx, cancel := context.WithTimeout(context.Background(), allocator.AddTimeLimit)
	defer cancel()

	newip, err = kubernetes.IPManagement(ctx, allocator.Allocate, *ipamConf, args.ContainerID, getPodRef(args.Args))

	result.IPs = append(result.IPs, &current.IPConfig{
		Address: newip,
		Gateway: ipamConf.Gateway})

	// Assign all the static IP elements.
	for _, v := range ipamConf.Addresses {
		result.IPs = append(result.IPs, &current.IPConfig{
			Address: v.Address,
			Gateway: v.Gateway})
	}

	return cnitypes.PrintResult(result, confVersion)
}

func cmdDel(args *skel.CmdArgs) error {

}

func cmdCheck(args *skel.CmdArgs) error {
	return nil
}
