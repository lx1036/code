package vxlan

import (
	"fmt"
	"github.com/cilium/cilium/pkg/mac"
	"github.com/containernetworking/plugins/pkg/utils/sysctl"
	"net"
	"syscall"

	"github.com/vishvananda/netlink"
)

type vxlanDeviceAttrs struct {
	vni       uint32
	name      string
	vtepIndex int
	vtepAddr  net.IP
	vtepPort  int
	gbp       bool
	learning  bool
}

type vxlanDevice struct {
	link          *netlink.Vxlan
	directRouting bool
}

func newVXLANDevice(devAttrs *vxlanDeviceAttrs) (*vxlanDevice, error) {
	macAddr, err := mac.GenerateRandMAC()
	if err != nil {
		return nil, err
	}
	link := &netlink.Vxlan{
		LinkAttrs: netlink.LinkAttrs{
			Name:         devAttrs.name,
			HardwareAddr: net.HardwareAddr(macAddr),
		},
		VxlanId:      int(devAttrs.vni),
		VtepDevIndex: devAttrs.vtepIndex,
		SrcAddr:      devAttrs.vtepAddr,
		Port:         devAttrs.vtepPort,
		Learning:     devAttrs.learning,
		GBP:          devAttrs.gbp,
	}
	link, err = ensureLink(link)
	if err != nil {
		return nil, err
	}

	// disable IPv6 Router Advertisement
	_, _ = sysctl.Sysctl(fmt.Sprintf("net/ipv6/conf/%s/accept_ra", devAttrs.name), "0")

	return &vxlanDevice{
		link: link,
	}, nil
}

func (dev *vxlanDevice) MACAddr() net.HardwareAddr {
	return dev.link.HardwareAddr
}

// 创建一个 vxlan 网卡
func ensureLink(vxlan *netlink.Vxlan) (*netlink.Vxlan, error) {
	err := netlink.LinkAdd(vxlan)
	if err == syscall.EEXIST {
		existing, err := netlink.LinkByName(vxlan.Name)
		if err != nil {
			return nil, err
		}
		if checkVxlanLink(vxlan, existing) {
			return existing.(*netlink.Vxlan), nil
		}

		if err = netlink.LinkDel(existing); err != nil {
			return nil, fmt.Errorf("failed to delete interface: %v", err)
		}
		if err = netlink.LinkAdd(vxlan); err != nil {
			return nil, fmt.Errorf("failed to create vxlan interface: %v", err)
		}
	}

	link, err := netlink.LinkByIndex(vxlan.Index)
	if err != nil {
		return nil, fmt.Errorf("can't locate created vxlan device with index %v", vxlan.Index)
	}

	return link.(*netlink.Vxlan), nil
}

func checkVxlanLink(l1, l2 netlink.Link) bool {
	if l1.Type() != l2.Type() {
		return false
	}

	v1 := l1.(*netlink.Vxlan)
	v2 := l2.(*netlink.Vxlan)
	if v1.VxlanId != v2.VxlanId ||
		(v1.VtepDevIndex > 0 && v2.VtepDevIndex > 0 && v1.VtepDevIndex != v2.VtepDevIndex) ||
		(len(v1.SrcAddr) > 0 && len(v2.SrcAddr) > 0 && !v1.SrcAddr.Equal(v2.SrcAddr)) ||
		(len(v1.Group) > 0 && len(v2.Group) > 0 && !v1.Group.Equal(v2.Group)) ||
		v1.L2miss != v2.L2miss ||
		(v1.Port > 0 && v2.Port > 0 && v1.Port != v2.Port) ||
		v1.GBP != v2.GBP {
		return false
	}

	return true
}
