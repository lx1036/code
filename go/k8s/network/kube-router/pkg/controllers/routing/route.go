package routing

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/vishvananda/netlink/nl"
	"golang.org/x/sys/unix"
	"io/ioutil"
	"k8s-lx1036/k8s/network/kube-router/pkg/metrics"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"

	gobgpapi "github.com/osrg/gobgp/api"
	"github.com/vishvananda/netlink"
	"k8s.io/klog/v2"
)

func (controller *NetworkRoutingController) injectRoute(path *gobgpapi.Path) error {
	klog.Infof("injectRoute Path Looks Like: %s", path.String())
	var route *netlink.Route
	var link netlink.Link

	dst, nextHop, err := parseBGPPath(path)
	if err != nil {
		return err
	}

	tunnelName := generateTunnelName(nextHop.String())
	sameSubnet := controller.nodeSubnet.Contains(nextHop)

	// INFO: 如果是删除路由请求
	if path.IsWithdraw {
		klog.Infof("Removing route: '%s via %s' from peer in the routing table", dst, nextHop)

		// The path might be withdrawn because the peer became unestablished or it may be withdrawn because just the
		// path was withdrawn. Check to see if the peer is still established before deciding whether to clean the
		// tunnel and tunnel routes or whether to just delete the destination route.
		peerEstablished, err := controller.isPeerEstablished(nextHop.String())
		if err != nil {
			klog.Errorf("encountered error while checking peer status: %v", err)
		}
		if err == nil && !peerEstablished {
			klog.Infof("Peer '%s' was not found any longer, removing tunnel and routes", nextHop.String())
			controller.cleanupTunnel(dst, tunnelName)
			return nil
		}

		return deleteRoutesByDestination(dst)
	}

	shouldCreateTunnel := func() bool {
		if !controller.enableOverlays {
			return false
		}
		if controller.overlayType == "full" {
			return true
		}
		if controller.overlayType == "subnet" && !sameSubnet {
			return true
		}
		return false
	}

	// create IPIP tunnels only when node is not in same subnet or overlay-type is set to 'full'
	// if the user has disabled overlays, don't create tunnels. If we're not creating a tunnel, check to see if there is
	// any cleanup that needs to happen.
	if shouldCreateTunnel() {
		link, err = controller.setupOverlayTunnel(tunnelName, nextHop)
		if err != nil {
			return err
		}
	} else {
		// knowing that a tunnel shouldn't exist for this route, check to see if there are any lingering tunnels /
		// routes that need to be cleaned up.
		controller.cleanupTunnel(dst, tunnelName)
	}

	switch {
	case link != nil:
		// if we setup an overlay tunnel link, then use it for destination routing
		route = &netlink.Route{
			LinkIndex: link.Attrs().Index,
			Src:       controller.nodeIP,
			Dst:       dst,
			Protocol:  unix.RTPROT_MROUTED,
		}
	case sameSubnet:
		// if the nextHop is within the same subnet, add a route for the destination so that traffic can bet routed
		// at layer 2 and minimize the need to traverse a router
		route = &netlink.Route{
			Dst:      dst,
			Gw:       nextHop,
			Protocol: unix.RTPROT_MROUTED,
		}
	default:
		// otherwise, let BGP do its thing, nothing to do here
		return nil
	}

	// Alright, everything is in place, and we have our route configured, let's add it to the host's routing table
	klog.Infof("Inject route: '%s via %s' from peer to routing table", dst, nextHop)
	return netlink.RouteReplace(route)
}

func deleteRoutesByDestination(destinationSubnet *net.IPNet) error {
	routes, err := netlink.RouteListFiltered(nl.FAMILY_ALL, &netlink.Route{
		Dst:      destinationSubnet,
		Protocol: unix.RTPROT_MROUTED,
	}, netlink.RT_FILTER_DST|netlink.RT_FILTER_PROTOCOL)
	if err != nil {
		return fmt.Errorf("failed to get routes from netlink: %v", err)
	}

	for i, route := range routes {
		klog.Infof("Found route to remove: %s", route.String())
		if err := netlink.RouteDel(&routes[i]); err != nil {
			return fmt.Errorf("failed to remove route due to %v", err)
		}
	}

	return nil
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
func ensureCustomRouteTable(tableNumber, tableName string) error {
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
		if _, err = f.WriteString(fmt.Sprintf("%s %s\n", tableNumber, tableName)); err != nil {
			return fmt.Errorf("failed to write: %s", err.Error())
		}
	}

	return nil
}

// INFO: 自定义一个 kube-router 路由表，然后添加 route policy: `ip rule add from ${pod_cidr} lookup kube-router`
func (controller *NetworkRoutingController) enablePolicyBasedRouting() error {
	err := ensureCustomRouteTable(CustomRouteTableID, CustomRouteTableName)
	if err != nil {
		return fmt.Errorf("failed to update rt_tables file: %s", err)
	}

	out, err := exec.Command("ip", "rule", "list").Output()
	if err != nil {
		return fmt.Errorf("failed to verify if `ip rule` exists: %s", err.Error())
	}

	if !strings.Contains(string(out), controller.podCidr) {
		err = exec.Command("ip", "rule", "add", "from", controller.podCidr, "lookup", CustomRouteTableName).Run()
		if err != nil {
			return fmt.Errorf("failed to add ip rule due to: %s", err.Error())
		}
	}

	return nil
}

func (controller *NetworkRoutingController) disablePolicyBasedRouting() error {
	err := ensureCustomRouteTable(CustomRouteTableID, CustomRouteTableName)
	if err != nil {
		return fmt.Errorf("failed to update rt_tables file: %s", err)
	}

	out, err := exec.Command("ip", "rule", "list").Output()
	if err != nil {
		return fmt.Errorf("failed to verify if `ip rule` exists: %s", err.Error())
	}

	if strings.Contains(string(out), controller.podCidr) {
		err = exec.Command("ip", "rule", "del", "from", controller.podCidr, "table", CustomRouteTableName).Run()
		if err != nil {
			return fmt.Errorf("failed to delete ip rule due to: %s", err.Error())
		}
	}

	return nil
}

func (controller *NetworkRoutingController) advertisePodRoute() error {
	metrics.ControllerBGPAdvertisementsSent.Inc()

	subnet, mask, err := controller.splitPodCidr()
	if err != nil {
		return err
	}

	// only ipv4
	klog.Infof(fmt.Sprintf("Advertising route: '%s/%d via %s' to peers", subnet, mask, controller.nodeIP.String()))
	nlri, _ := ptypes.MarshalAny(&gobgpapi.IPAddressPrefix{
		PrefixLen: uint32(mask),
		Prefix:    subnet,
	})
	a1, _ := ptypes.MarshalAny(&gobgpapi.OriginAttribute{
		Origin: 0,
	})
	a2, _ := ptypes.MarshalAny(&gobgpapi.NextHopAttribute{
		NextHop: controller.nodeIP.String(),
	})
	attrs := []*any.Any{a1, a2}
	_, err = controller.bgpServer.AddPath(context.Background(), &gobgpapi.AddPathRequest{
		Path: &gobgpapi.Path{
			Family: &gobgpapi.Family{Afi: gobgpapi.Family_AFI_IP, Safi: gobgpapi.Family_SAFI_UNICAST},
			Nlri:   nlri,
			Pattrs: attrs,
		},
	})
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("advertise pod cidr %s err:%v", controller.podCidr, err))
	}

	return nil
}

func (controller *NetworkRoutingController) splitPodCidr() (subnet string, mask int, err error) {
	cidrStr := strings.Split(controller.podCidr, "/")
	subnet = cidrStr[0]
	mask, err = strconv.Atoi(cidrStr[1])
	if err != nil || mask < 0 || mask > 32 {
		return "", 0, fmt.Errorf("the pod CIDR IP given is not a proper mask: %d", mask)
	}

	return subnet, mask, nil
}
