package main

import (
	"fmt"
	"net"

	"k8s-lx1036/k8s/network/cni/cni-on-vm/host-local-ipam/pkg/allocator"
	"k8s-lx1036/k8s/network/cni/cni-on-vm/host-local-ipam/pkg/store/disk"

	"github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/100"

	"github.com/containernetworking/cni/pkg/skel"
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

	store, err := disk.New(ipamConf.Name, ipamConf.DataDir)
	if err != nil {
		return err
	}
	defer store.Close()

	var requestedIPs []net.IP
	for _, ip := range ipamConf.IPArgs {
		requestedIPs = append(requestedIPs, ip) // 指定的分配的这些 ip
	}

	if len(requestedIPs) != 0 {
		for idx, rangeset := range ipamConf.Ranges {
			ipAllocator := allocator.NewIPAllocator(&rangeset, store, idx)
			for _, ip := range requestedIPs {
				if rangeset.Contains(ip) {
					ipConf, err := ipAllocator.AllocateIP(args.ContainerID, args.IfName, ip)
					if err != nil {
						ipAllocator.Release(args.ContainerID, args.IfName)
						return fmt.Errorf("failed to allocate for range %d: %v", idx, err)
					}

					result.IPs = append(result.IPs, ipConf)
				}
			}
		}
	} else {
		for idx, rangeset := range ipamConf.Ranges {
			ipAllocator := allocator.NewIPAllocator(&rangeset, store, idx)
			ipConf, err := ipAllocator.AllocateNext(args.ContainerID, args.IfName)
			if err != nil {
				ipAllocator.Release(args.ContainerID, args.IfName)
				return fmt.Errorf("failed to allocate for range %d: %v", idx, err)
			}

			result.IPs = append(result.IPs, ipConf)
		}
	}

	result.Routes = ipamConf.Routes

	return types.PrintResult(result, confVersion)
}

func cmdDel(args *skel.CmdArgs) error {
	return nil
}

func cmdCheck(args *skel.CmdArgs) error {

}
