package driver

import (
	"fmt"
	"net"

	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/utils"
	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/utils/sysctl"
	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/utils/types"

	"github.com/vishvananda/netlink"
	k8snet "k8s.io/apimachinery/pkg/util/net"
)

func GetHostIP(ipv4, ipv6 bool) (*types.IPNetSet, error) {
	var nodeIPv4, nodeIPv6 *net.IPNet

	if ipv4 {
		v4, err := k8snet.ResolveBindAddress(net.ParseIP("127.0.0.1"))
		if err != nil {
			return nil, err
		}
		if utils.IPv6(v4) {
			return nil, fmt.Errorf("error get node ipv4 address.This may dure to 1. no ipv4 address 2. no ipv4 default route")
		}
		nodeIPv4 = &net.IPNet{
			IP:   v4,
			Mask: net.CIDRMask(32, 32),
		}
	}

	if ipv6 {
		v6, err := k8snet.ResolveBindAddress(net.ParseIP("::1"))
		if err != nil {
			return nil, err
		}
		if !utils.IPv6(v6) {
			return nil, fmt.Errorf("error get node ipv6 address.This may dure to 1. no ipv6 address 2. no ipv6 default route")
		}
		nodeIPv6 = &net.IPNet{
			IP:   v6,
			Mask: net.CIDRMask(128, 128),
		}
	}
	return &types.IPNetSet{
		IPv4: nodeIPv4,
		IPv6: nodeIPv6,
	}, nil
}

func EnableIPv6() error {
	err := sysctl.WriteProcSys("/proc/sys/net/ipv6/conf/all/disable_ipv6", "0")
	if err != nil {
		return err
	}
	err = sysctl.WriteProcSys("/proc/sys/net/ipv6/conf/default/disable_ipv6", "0")
	if err != nil {
		return err
	}
	return nil
}

// EnsureLinkUp set link up,return changed and err
func EnsureLinkUp(link netlink.Link) (bool, error) {
	if link.Attrs().Flags&net.FlagUp != 0 {
		return false, nil
	}

	err := netlink.LinkSetUp(link)
	if err != nil {
		return false, fmt.Errorf(fmt.Sprintf("`ip link set %s up` err:%v", link.Attrs().Name, err))
	}

	return true, nil
}
