package vxlan

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/vishvananda/netlink"
	"syscall"

	"k8s-lx1036/k8s/network/cni/flannel/pkg/backend"
	"k8s-lx1036/k8s/network/cni/flannel/pkg/ip"
	"k8s-lx1036/k8s/network/cni/flannel/pkg/subnet"

	"k8s.io/klog/v2"
)

type vxlanNetwork struct {
	dev *vxlanDevice
}

func newNetwork(subnetMgr subnet.Manager, extIface *backend.ExternalInterface, dev *vxlanDevice,
	_ ip.IP4Net, lease *subnet.Lease) (*vxlanNetwork, error) {
	return &vxlanNetwork{
		dev: dev,
	}, nil
}

func (network *vxlanNetwork) Run(ctx context.Context) {
	klog.Info("watching for new subnet leases")
	events := make(chan []subnet.Event)
	go subnet.WatchLeases(ctx, network.subnetMgr, network.SubnetLease, events)

	for {
		evtBatch, ok := <-events
		if !ok {
			klog.Infof("evts chan closed")
			return
		}

		network.handleSubnetEvents(evtBatch)
	}
}

type vxlanLeaseAttrs struct {
	VNI     uint16
	VtepMAC hardwareAddr
}

func (network *vxlanNetwork) handleSubnetEvents(batch []subnet.Event) {

	for _, event := range batch {
		sn := event.Lease.Subnet
		attrs := event.Lease.Attrs
		if attrs.BackendType != "vxlan" {
			klog.Warningf(fmt.Sprintf("ignoring non-vxlan v4Subnet(%s): type=%v", sn, attrs.BackendType))
			continue
		}

		var (
			vxlanAttrs  vxlanLeaseAttrs
			vxlanRoute  netlink.Route
			directRoute netlink.Route
		)

		if event.Lease.EnableIPv4 && network.dev != nil {
			if err := json.Unmarshal(attrs.BackendData, &vxlanAttrs); err != nil {
				klog.Error("error decoding subnet lease JSON: ", err)
				continue
			}

			// This route is used when traffic should be vxlan encapsulated
			vxlanRoute = netlink.Route{
				LinkIndex: network.dev.link.Attrs().Index,
				Scope:     netlink.SCOPE_UNIVERSE,
				Dst:       sn.ToIPNet(),
				Gw:        sn.IP.ToIP(),
			}
			vxlanRoute.SetFlag(syscall.RTNH_F_ONLINK)

			// directRouting is where the remote host is on the same subnet so vxlan isn't required.
			directRoute = netlink.Route{
				Dst: sn.ToIPNet(),
				Gw:  attrs.PublicIP.ToIP(),
			}
			if nw.dev.directRouting {
				if dr, err := ip.DirectRouting(attrs.PublicIP.ToIP()); err != nil {
					log.Error(err)
				} else {
					directRoutingOK = dr
				}
			}
		}

	}
}
