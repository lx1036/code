package vxlan

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/vishvananda/netlink"
	"net"
	"syscall"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/test/tunnel/vxlan/flannel/pkg/backend"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/test/tunnel/vxlan/flannel/pkg/subnet"

	"k8s.io/klog/v2"
)

const (
	encapOverhead = 50
)

type vxlanNetwork struct {
	dev         *vxlanDevice
	subnetMgr   subnet.Manager
	SubnetLease *subnet.Lease
	ExtIface    *backend.ExternalInterface
}

func newNetwork(subnetMgr subnet.Manager, extIface *backend.ExternalInterface, dev *vxlanDevice,
	lease *subnet.Lease) (*vxlanNetwork, error) {
	return &vxlanNetwork{
		dev:         dev,
		subnetMgr:   subnetMgr,
		ExtIface:    extIface,
		SubnetLease: lease,
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

func (network *vxlanNetwork) Lease() *subnet.Lease {
	return network.SubnetLease
}

func (network *vxlanNetwork) MTU() int {
	return network.ExtIface.Iface.MTU - encapOverhead
}

// So we can make it JSON (un)marshalable
type hardwareAddr net.HardwareAddr

type vxlanLeaseAttrs struct {
	VNI     uint16
	VtepMAC hardwareAddr
}

// 配置路由和 arp, 这里 batch 包含除了当前 node subnet 之外的其他所有 node subnet
// INFO: flanneld 通过 watch k8s Node 来动态地维护各节点通信所需的ARP、FDB以及路由条目
func (network *vxlanNetwork) handleSubnetEvents(batch []subnet.Event) {
	for _, event := range batch {
		sn := event.Lease.Subnet
		attrs := event.Lease.Attrs
		if attrs.BackendType != "vxlan" {
			klog.Warningf(fmt.Sprintf("ignoring non-vxlan v4Subnet(%s): type=%v", sn, attrs.BackendType))
			continue
		}

		var (
			vxlanAttrs      vxlanLeaseAttrs
			vxlanRoute      netlink.Route
			directRoute     netlink.Route
			directRoutingOK bool
		)
		if event.Lease.EnableIPv4 && network.dev != nil {
			if err := json.Unmarshal(attrs.BackendData, &vxlanAttrs); err != nil {
				klog.Error("error decoding subnet lease JSON: ", err)
				continue
			}

			// This route is used when traffic should be vxlan encapsulated
			// INFO: 这里是最重点之处，为每一个其他的 node pod cidr 创建一个路由，指向 vxlan 网卡。就是所谓的 FullMesh 模式
			//  这里有个问题是：vxlan 网卡怎么根据目标 map[podIP]nodeIP 来封包的，是可以直接通过该路由就有的功能么，vxlan 的功能？？？
			vxlanRoute = netlink.Route{
				LinkIndex: network.dev.link.Attrs().Index,
				Scope:     netlink.SCOPE_UNIVERSE,
				// 到达新 node 的 vxlan 网卡的路由，`10.230.91.0/24 via 10.230.91.0 dev flannel.1 onlink`,
				// 然后 10.230.91.0 对应的 mac 在 ARP 表里，下面 AddARP() 设置了，然后这个 mac 对应的 IP 在 FDB 表里设置了，就是另一台 nodeIP，最后封包结束
				Dst: sn.ToIPNet(),
				Gw:  sn.IP.ToIP(), // 10.230.91.0
			}
			// flanneld在加入集群时会为每个其他节点生成一条on-link路由，on-link路由表示是直连路由，匹配该条路由的数据包将触发ARP请求获取目的IP的MAC地址
			vxlanRoute.SetFlag(syscall.RTNH_F_ONLINK)

			// INFO: 有了这条 onlink 路由，网桥 cni0 会把来自于 veth pair 的包转给 vxlan flannel.1 网卡
			//  并且接收端的IP地址为10.230.93.0, 需要通过ARP获取MAC地址, @see http://just4coding.com/2021/11/03/flannel/

			// directRouting is where the remote host is on the same subnet so vxlan isn't required.
			directRoute = netlink.Route{
				Dst: sn.ToIPNet(),
				Gw:  attrs.PublicIP.ToIP(),
			}
			if network.dev.directRouting {
				if dr, err := DirectRouting(attrs.PublicIP.ToIP()); err != nil {
					klog.Error(err)
				} else {
					directRoutingOK = dr
				}
			}
		}

		// INFO: 这里是最重点之处，主要配置 route/arp/fdb，就可以实现了 vxlan 封包，回答了上面的问题
		//  其实主要就是实现了这篇文章所说的配置 http://just4coding.com/2020/04/20/vxlan-fdb/ , 配置 arp/fdb 来实现 vxlan 封包
		switch event.Type {
		case subnet.EventAdded:
			if event.Lease.EnableIPv4 {
				if directRoutingOK {
					klog.Infof(fmt.Sprintf("Adding direct route to subnet: %s for nodeIP: %s", sn, attrs.PublicIP))
					if err := netlink.RouteReplace(&directRoute); err != nil {
						klog.Errorf("Error adding route to %v via %v: %v", sn, attrs.PublicIP, err)
						continue
					}
				} else {
					klog.Infof(fmt.Sprintf("adding subnet: %s for nodeIP: %s VtepMAC: %s",
						sn, attrs.PublicIP, net.HardwareAddr(vxlanAttrs.VtepMAC)))
					// 这里 sn.IP 是 vxlan 网卡的 IP
					if err := network.dev.AddARP(neighbor{IP: sn.IP, MAC: net.HardwareAddr(vxlanAttrs.VtepMAC)}); err != nil {
						klog.Error("AddARP failed: ", err)
						continue
					}
					// 这里 attrs.PublicIP 就是 nodeIP, 或者也叫 VTEP IP 地址, 可以参考验证 @see http://just4coding.com/2020/04/20/vxlan-fdb/
					if err := network.dev.AddFDB(neighbor{IP: attrs.PublicIP, MAC: net.HardwareAddr(vxlanAttrs.VtepMAC)}); err != nil {
						// Try to clean up the ARP entry then continue
						if err := network.dev.DelARP(neighbor{IP: sn.IP, MAC: net.HardwareAddr(vxlanAttrs.VtepMAC)}); err != nil {
							klog.Error("DelARP failed: ", err)
						}

						continue
					}

					// Set the route - the kernel would ARP for the Gw IP address if it hadn't already been set above so make sure
					// this is done last.
					if err := netlink.RouteReplace(&vxlanRoute); err != nil {
						klog.Errorf(fmt.Sprintf("failed to add vxlanRoute (%s -> %s): %v", vxlanRoute.Dst, vxlanRoute.Gw, err))
						// Try to clean up both the ARP and FDB entries then continue
						if err := network.dev.DelARP(neighbor{IP: sn.IP, MAC: net.HardwareAddr(vxlanAttrs.VtepMAC)}); err != nil {
							klog.Error("DelARP failed: ", err)
						}
						if err := network.dev.DelFDB(neighbor{IP: attrs.PublicIP, MAC: net.HardwareAddr(vxlanAttrs.VtepMAC)}); err != nil {
							klog.Error("DelFDB failed: ", err)
						}

						continue
					}
				}
			}
		case subnet.EventRemoved:
			if event.Lease.EnableIPv4 {
				if directRoutingOK {
					klog.Infof("Removing direct route to subnet: %s PublicIP: %s", sn, attrs.PublicIP)
					if err := netlink.RouteDel(&directRoute); err != nil {
						klog.Errorf("Error deleting route to %v via %v: %v", sn, attrs.PublicIP, err)
					}
				} else {
					klog.Infof("removing subnet: %s PublicIP: %s VtepMAC: %s", sn, attrs.PublicIP, net.HardwareAddr(vxlanAttrs.VtepMAC))
					// Try to remove all entries - don't bail out if one of them fails.
					if err := network.dev.DelARP(neighbor{IP: sn.IP, MAC: net.HardwareAddr(vxlanAttrs.VtepMAC)}); err != nil {
						klog.Error("DelARP failed: ", err)
					}
					if err := network.dev.DelFDB(neighbor{IP: attrs.PublicIP, MAC: net.HardwareAddr(vxlanAttrs.VtepMAC)}); err != nil {
						klog.Error("DelFDB failed: ", err)
					}
					if err := netlink.RouteDel(&vxlanRoute); err != nil {
						klog.Errorf("failed to delete vxlanRoute (%s -> %s): %v", vxlanRoute.Dst, vxlanRoute.Gw, err)
					}
				}
			}
		}
	}
}

// DirectRouting 没有 gatewayIP 则是 DirectRouting
func DirectRouting(ip net.IP) (bool, error) {
	routes, err := netlink.RouteGet(ip)
	if err != nil {
		return false, fmt.Errorf("couldn't lookup route to %v: %v", ip, err)
	}
	if len(routes) == 1 && routes[0].Gw == nil {
		// There is only a single route and there's no gateway (i.e. it's directly connected)
		return true, nil
	}
	return false, nil
}
