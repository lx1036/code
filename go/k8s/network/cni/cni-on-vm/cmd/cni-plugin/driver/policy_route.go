package driver

import (
	"fmt"
	"github.com/containernetworking/plugins/pkg/ip"
	"github.com/containernetworking/plugins/pkg/ns"
	"io/ioutil"
	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/utils/types"
	"math"
	"net"
	"os"
	"strings"

	"github.com/vishvananda/netlink"
)

const (
	toContainerPriority   = 512
	fromContainerPriority = 2048
)

var (
	Gateway = &net.IPNet{
		IP:   net.IPv4(169, 254, 1, 1),
		Mask: net.CIDRMask(32, 32),
	}
)

type PolicyRoute struct{}

func NewPolicyRoute() *PolicyRoute {
	return &PolicyRoute{}
}

// INFO:
//  (1) Pod流量控制: Pod上的ingress和egress的annotation配置，然后通过配置Pod的网卡上的TC的tbf规则来实现对速度的限制

func (d *PolicyRoute) Setup(cfg *types.SetupConfig, netNS ns.NetNS) error {
	// (1) 创建容器侧的 veth interface
	// due to kernel bug we have to create with tmpname or it might
	// collide with the name on the host and error out
	// INFO: cfg.ContainerIfName 有可能是 eth0 等，会触发内核 bug
	tmpName, _ := ip.RandomVethName()
	err := netlink.LinkAdd(&netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name:      tmpName,
			MTU:       cfg.MTU,
			Flags:     net.FlagUp, // `ip link set ipv1 up`
			Namespace: netlink.NsFd(int(netNS.Fd())),
		},
		PeerName: cfg.HostVethIfName, // 该网卡在宿主机侧
	})
	if err != nil {
		return err
	}
	hostVethLink, err := netlink.LinkByName(cfg.HostVethIfName)
	if err != nil {
		return err
	}

	// (2) add ip and add default route, arp neigh
	err = netNS.Do(func(netNS ns.NetNS) error {
		containerLink, err := netlink.LinkByName(tmpName)
		if err != nil {
			return err
		}
		err = netlink.LinkSetName(containerLink, cfg.ContainerIfName)
		if err != nil {
			return err
		}
		containerLink, err = netlink.LinkByName(cfg.ContainerIfName)
		if err != nil {
			return err
		}
		err = netlink.AddrAdd(containerLink, &netlink.Addr{ // `sudo ip netns exec net1 ip addr add 10.0.1.10/24 dev ipv1`
			IPNet: cfg.ContainerIPNet.IPv4,
		})
		if err != nil {
			return err
		}

		routes := []*netlink.Route{
			{
				LinkIndex: containerLink.Attrs().Index,
				Scope:     netlink.SCOPE_UNIVERSE,
				Dst:       defaultRoute,
				Gw:        Gateway.IP,
				Flags:     int(netlink.FLAG_ONLINK),
			},
			{
				LinkIndex: containerLink.Attrs().Index,
				Scope:     netlink.SCOPE_LINK,
				Dst:       Gateway,
			},
		}
		for _, route := range routes {
			_ = netlink.RouteAdd(route)
		}

		err = netlink.NeighAdd(&netlink.Neigh{ // arp host IP to container eth0 mac
			LinkIndex:    containerLink.Attrs().Index,
			IP:           Gateway.IP,
			HardwareAddr: hostVethLink.Attrs().HardwareAddr, // host 侧的网卡
			State:        netlink.NUD_PERMANENT,
		})
		if err != nil {
			return err
		}

		if cfg.Egress > 0 { // pod egress 是在容器侧的网卡加上 tc tbf 限速
			err = createTBF(containerLink, cfg.Egress)
			if err != nil {
				return err
			}
		}
		return nil
	})

	// 在 kube-router 路由表内配置弹性网卡ENI的路由
	eni, err := netlink.LinkByIndex(cfg.ENIIndex)
	if err != nil {
		return err
	}
	/*err = ensureCustomRouteTable(1000+eni.Attrs().Index, CustomRouteTableName)
	if err != nil {
		return err
	}*/
	containerIPNet := &net.IPNet{ // podIP/32
		IP:   cfg.ContainerIPNet.IPv4.IP,
		Mask: net.CIDRMask(32, 32),
	}
	tableID := getRouteTableID(eni.Attrs().Index)
	routes := []*netlink.Route{
		{ // ip route add default table tableID scope global via GatewayIPV4 dev eth-eni
			LinkIndex: eni.Attrs().Index,
			Scope:     netlink.SCOPE_UNIVERSE,
			Table:     tableID,
			Dst:       defaultRoute,
			Gw:        cfg.GatewayIP.IPv4,
			Flags:     int(netlink.FLAG_ONLINK),
		},
		{ // ip route add containerIPNet scope link dev hostVethName
			LinkIndex: hostVethLink.Attrs().Index,
			Scope:     netlink.SCOPE_LINK,
			Dst:       containerIPNet,
		},
	}
	for _, route := range routes {
		if err = netlink.RouteAdd(route); err != nil {
			return err
		}
	}

	rules := []*netlink.Rule{ // `ip rule list table tableID`
		{ // ip rule add to podIP/32 table tableID prio 512
			Dst:      containerIPNet,
			Table:    tableID,
			Priority: toContainerPriority,
		},
		{ // ip rule add from podIP/32 table tableID prio 2048 iif hostVethName (从 hostVethName 网卡且 ip 为 podIP/32 的 packet，查询路由表 tableID)
			Src:      containerIPNet,
			Table:    tableID,
			Priority: fromContainerPriority,
			IifName:  hostVethLink.Attrs().Name,
		},
	}
	for _, rule := range rules {
		if err = netlink.RuleAdd(rule); err != nil {
			return err
		}
	}

	if cfg.Ingress > 0 {
		err = createTBF(hostVethLink, cfg.Ingress)
		if err != nil {
			return err
		}
	}

	return nil
}

func getRouteTableID(index int) int {
	return 1000 + index
}

const (
	latencyInMillis   = 25
	hardwareHeaderLen = 1500
)

func time2Tick(time uint32) uint32 {
	return uint32(float64(time) * float64(netlink.TickInUsec()))
}

func burst(rate uint64, mtu int) uint32 {
	return uint32(math.Ceil(math.Max(float64(rate)/netlink.Hz(), float64(mtu))))
}
func buffer(rate uint64, burst uint32) uint32 {
	return time2Tick(uint32(float64(burst) * float64(netlink.TIME_UNITS_PER_SEC) / float64(rate)))
}

func latencyInUsec(latencyInMillis float64) float64 {
	return float64(netlink.TIME_UNITS_PER_SEC) * (latencyInMillis / 1000.0)
}

func limit(rate uint64, latency float64, buffer uint32) uint32 {
	return uint32(float64(rate)*latency/float64(netlink.TIME_UNITS_PER_SEC)) + buffer
}

// throttle traffic on ifb device，对 linkIndex 网卡限流
func createTBF(link netlink.Link, rateInBytes uint64) error {
	burstInBytes := burst(rateInBytes, link.Attrs().MTU+hardwareHeaderLen)
	bufferInBytes := buffer(uint64(rateInBytes), uint32(burstInBytes))
	latency := latencyInUsec(latencyInMillis)
	limitInBytes := limit(uint64(rateInBytes), latency, uint32(burstInBytes))
	qdisc := &netlink.Tbf{
		QdiscAttrs: netlink.QdiscAttrs{
			LinkIndex: link.Attrs().Index,
			// https://www.kernel.org/doc/html/latest/admin-guide/cgroup-v1/net_cls.html
			// 0x10000 classid is 1:0，如果 classid is 10:1 0x100001
			Handle: netlink.MakeHandle(1, 0),
			Parent: netlink.HANDLE_ROOT,
		},
		Limit:  uint32(limitInBytes),
		Rate:   uint64(rateInBytes),
		Buffer: uint32(bufferInBytes),
	}
	// tc qdisc add dev lxcXXX root tbf rate rateInBytes burst netConf.BandwidthLimits.Burst
	return netlink.QdiscAdd(qdisc)
}

const (
	RouteTable           = "/etc/iproute2/rt_tables"
	CustomRouteTableID   = "77"
	CustomRouteTableName = "kube-router"
)

// https://linuxgeeks.github.io/2017/03/17/170119-Linux%E7%9A%84%E7%AD%96%E7%95%A5%E8%B7%AF%E7%94%B1/
/*
	# `cat /etc/iproute2/rt_tables`
	#
	# reserved values
	#
	255 local
	254 main
	253 default
	0 unspec
	#
	# local
	#
	#1 inr.ruhep

	77 kube-router
*/
func ensureCustomRouteTable(tableNumber int, tableName string) error {
	content, err := ioutil.ReadFile(RouteTable)
	if err != nil {
		return fmt.Errorf("failed to read: %s", err.Error())
	}

	if !strings.Contains(string(content), tableName) {
		f, err := os.OpenFile(RouteTable, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return fmt.Errorf("failed to open: %s", err.Error())
		}
		defer f.Close()
		if _, err = f.WriteString(fmt.Sprintf("%d %s\n", tableNumber, tableName)); err != nil {
			return fmt.Errorf("failed to write: %s", err.Error())
		}
	}

	return nil
}
