package vxlan

import (
	"fmt"
	"github.com/cilium/cilium/pkg/mac"
	"github.com/containernetworking/plugins/pkg/utils/sysctl"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/test/tunnel/vxlan/flannel/pkg/ip"
	"k8s.io/klog/v2"
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

// Configure ensures that there is only one v4 Addr on `link` within the `ipn` address space and it equals `ipa`.
func (dev *vxlanDevice) Configure(flannelIP ip.IP4Net, flannelNet ip.IP4Net) error {
	existingAddrs, err := netlink.AddrList(dev.link, netlink.FAMILY_V4)
	if err != nil {
		return err
	}

	var hasAddr bool
	addr := netlink.Addr{IPNet: flannelIP.ToIPNet()}
	for _, existingAddr := range existingAddrs {
		if existingAddr.Equal(addr) {
			hasAddr = true
			continue
		}

		if flannelNet.Contains(ip.FromIP(existingAddr.IP)) {
			if err := netlink.AddrDel(dev.link, &existingAddr); err != nil {
				return fmt.Errorf("failed to remove IP address %s from %s: %s",
					existingAddr.String(), dev.link.Attrs().Name, err)
			}
			klog.Infof("removed IP address %s from %s", existingAddr.String(), dev.link.Attrs().Name)
		}
	}

	// Actually add the desired address to the interface if needed.
	if !hasAddr {
		if err := netlink.AddrAdd(dev.link, &addr); err != nil {
			return fmt.Errorf("failed to add IP address %s to %s: %s", addr.String(), dev.link.Attrs().Name, err)
		}
	}

	return nil
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

// ARP
type neighbor struct {
	MAC net.HardwareAddr
	IP  ip.IP4
}

// AddARP 给新建的 vxlan 网卡添加 ARP
func (dev *vxlanDevice) AddARP(n neighbor) error {
	klog.Infof(fmt.Sprintf("calling AddARP: %v, %v", n.IP, n.MAC))
	return netlink.NeighSet(&netlink.Neigh{
		LinkIndex:    dev.link.Index,
		State:        netlink.NUD_PERMANENT,
		Type:         syscall.RTN_UNICAST,
		IP:           n.IP.ToIP(),
		HardwareAddr: n.MAC,
	})
}

func (dev *vxlanDevice) DelARP(n neighbor) error {
	klog.Infof("calling DelARP: %v, %v", n.IP, n.MAC)
	return netlink.NeighDel(&netlink.Neigh{
		LinkIndex:    dev.link.Index,
		State:        netlink.NUD_PERMANENT,
		Type:         syscall.RTN_UNICAST,
		IP:           n.IP.ToIP(),
		HardwareAddr: n.MAC,
	})
}

// FDB: Forwarding Database
// VXLAN FDB 格式：<MAC> <VNI> <REMOTE IP>, VXLAN设备根据MAC地址来查找相应的VTEP IP地址，继而将二层数据帧封装发送至相应VTEP

func (dev *vxlanDevice) AddFDB(n neighbor) error {
	klog.Infof(fmt.Sprintf("calling AddFDB: %v, %v", n.IP, n.MAC))
	return netlink.NeighSet(&netlink.Neigh{
		LinkIndex:    dev.link.Index,
		State:        netlink.NUD_PERMANENT,
		Family:       syscall.AF_BRIDGE,
		Flags:        netlink.NTF_SELF,
		IP:           n.IP.ToIP(),
		HardwareAddr: n.MAC,
	})
}

func (dev *vxlanDevice) DelFDB(n neighbor) error {
	klog.Infof("calling DelFDB: %v, %v", n.IP, n.MAC)
	return netlink.NeighDel(&netlink.Neigh{
		LinkIndex:    dev.link.Index,
		Family:       syscall.AF_BRIDGE,
		Flags:        netlink.NTF_SELF,
		IP:           n.IP.ToIP(),
		HardwareAddr: n.MAC,
	})
}
