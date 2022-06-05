package main

import (
	"fmt"
	"net"
	"strings"

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

	requestedIPs, remainingIPs := map[string]net.IP{}, map[string]net.IP{} //net.IP cannot be a key
	for _, ip := range ipamConf.IPArgs {
		requestedIPs[ip.String()] = ip
		remainingIPs[ip.String()] = ip
	}

	ipAllocator := allocator.NewIPAllocator(&ipamConf.Ranges, store, 0)
	// Check to see if there are any custom IPs requested in this range.
	var requestedIP net.IP
	if len(requestedIPs) != 0 {
		for k, ip := range requestedIPs {
			if ipamConf.Ranges.Contains(ip) {
				requestedIP = ip
				delete(remainingIPs, k)

				ipConf, err := ipAllocator.AllocateIP(args.ContainerID, args.IfName, requestedIP)
				if err != nil {
					ipAllocator.Release(args.ContainerID, args.IfName)
					continue
				}

				result.IPs = append(result.IPs, ipConf)
			}
		}
	} else {
		ipConf, err := ipAllocator.AllocateNext(args.ContainerID, args.IfName)
		if err != nil {
			ipAllocator.Release(args.ContainerID, args.IfName)
			return fmt.Errorf("failed to allocate for range: %v", err)
		}

		result.IPs = append(result.IPs, ipConf)
	}

	if len(remainingIPs) != 0 {
		var msg []string
		for _, arg := range ipamConf.IPArgs {
			msg = append(msg, arg.String())
		}

		return fmt.Errorf(fmt.Sprintf("requested ip args %s is not in range for subnets %s",
			strings.Join(msg, ","), ipamConf.Ranges.String()))
	}

	result.Routes = ipamConf.Routes

	return types.PrintResult(result, confVersion)
}

func cmdDel(args *skel.CmdArgs) error {
	ipamConf, _, err := allocator.LoadIPAMConfig(args.StdinData, args.Args)
	if err != nil {
		return err
	}

	store, err := disk.New(ipamConf.Name, ipamConf.DataDir)
	if err != nil {
		return err
	}
	defer store.Close()

	ipAllocator := allocator.NewIPAllocator(&ipamConf.Ranges, store, 0)
	err = ipAllocator.Release(args.ContainerID, args.IfName)
	if err != nil {
		return err
	}

	return nil
}

func cmdCheck(args *skel.CmdArgs) error {
	return nil
}
