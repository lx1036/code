package vxlan

import (
	"context"
	"encoding/json"
	"fmt"
	"k8s-lx1036/k8s/network/cni/flannel/pkg/backend"
	"k8s-lx1036/k8s/network/cni/flannel/pkg/ip"
	"k8s-lx1036/k8s/network/cni/flannel/pkg/subnet"
	"k8s.io/klog/v2"
	"net"
)

func init() {
	backend.Register("vxlan", New)
}

const (
	defaultVNI = 1
)

type VxlanBackend struct {
	subnetMgr subnet.Manager
	extIface  *backend.ExternalInterface
}

func New(subnetMgr subnet.Manager, extIface *backend.ExternalInterface) (backend.Backend, error) {
	return &VxlanBackend{
		subnetMgr: subnetMgr,
		extIface:  extIface,
	}, nil
}

func (backend *VxlanBackend) RegisterNetwork(ctx context.Context, config *subnet.Config) (backend.Network, error) {
	/*
			{
		      "Network": "10.244.0.0/16",
		      "Backend": {
		        "Type": "vxlan"
		      }
		    }
	*/
	// @see vxlan config https://github.com/flannel-io/flannel/blob/master/Documentation/backends.md#vxlan
	cfg := struct {
		VNI           int
		Port          int  //  UDP port to use for sending encapsulated packets. On Linux, defaults to kernel default, currently 8472
		GBP           bool // Enable VXLAN Group Based Policy. Defaults to false.
		Learning      bool
		DirectRouting bool //  Enable direct routes (like host-gw) when the hosts are on the same subnet. VXLAN will only be used to encapsulate packets to hosts on different subnets. Defaults to false
	}{
		VNI: defaultVNI, // VXLAN Identifier (VNI) to be used. On Linux, defaults to 1
	}
	if len(config.Backend) > 0 {
		if err := json.Unmarshal(config.Backend, &cfg); err != nil {
			return nil, fmt.Errorf("error decoding VXLAN backend config: %v", err)
		}
	}
	klog.Infof(fmt.Sprintf("VXLAN config: VNI=%d Port=%d GBP=%v Learning=%v DirectRouting=%v",
		cfg.VNI, cfg.Port, cfg.GBP, cfg.Learning, cfg.DirectRouting))

	var dev *vxlanDevice
	var err error
	if config.EnableIPv4 {
		devAttrs := vxlanDeviceAttrs{
			vni:       uint32(cfg.VNI),
			name:      fmt.Sprintf("flannel.%v", cfg.VNI),
			vtepIndex: backend.extIface.Iface.Index,
			vtepAddr:  backend.extIface.IfaceAddr,
			vtepPort:  cfg.Port,
			gbp:       cfg.GBP,
			learning:  cfg.Learning,
		}

		// (1)创建了一个 vxlan 网卡
		dev, err = newVXLANDevice(&devAttrs)
		if err != nil {
			return nil, err
		}
		dev.directRouting = cfg.DirectRouting
	}

	subnetAttrs, err := newSubnetAttrs(backend.extIface.ExtAddr, uint16(cfg.VNI), dev)
	if err != nil {
		return nil, err
	}
	lease, err := backend.subnetMgr.AcquireLease(ctx, subnetAttrs)

	// Ensure that the device has a /32 address so that no broadcast routes are created.
	// This IP is just used as a source address for host to workload traffic (so
	// the return path for the traffic has an address on the flannel network to use as the destination)
	if config.EnableIPv4 {
		// (2)并配置了一个 subnetIP/32 地址，这个 subnet 其实来自于 newNode.Spec.PodCIDR
		if err := dev.Configure(ip.IP4Net{IP: lease.Subnet.IP, PrefixLen: 32}, config.Network); err != nil {
			return nil, fmt.Errorf("failed to configure interface %s: %w", dev.link.Attrs().Name, err)
		}
	}

	return newNetwork(backend.subnetMgr, backend.extIface, dev, lease)
}

// 这里的 publicIP 就是 nodeIP
func newSubnetAttrs(publicIP net.IP, vxlanID uint16, dev *vxlanDevice) (*subnet.LeaseAttrs, error) {
	leaseAttrs := &subnet.LeaseAttrs{
		BackendType: "vxlan",
	}
	if publicIP != nil && dev != nil {
		data, err := json.Marshal(&vxlanLeaseAttrs{
			VNI:     vxlanID,
			VtepMAC: hardwareAddr(dev.MACAddr()),
		})
		if err != nil {
			return nil, err
		}
		leaseAttrs.PublicIP = ip.FromIP(publicIP)
		leaseAttrs.BackendData = json.RawMessage(data)
	}

	return leaseAttrs, nil
}
