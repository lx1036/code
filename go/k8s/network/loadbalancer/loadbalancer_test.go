package loadbalancer

import (
	"fmt"
	"golang.org/x/sys/unix"
	"net"
	"testing"
	
	gobgpapi "github.com/osrg/gobgp/v3/api"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netlink/nl"
	"k8s.io/klog/v2"
)

func TestBGP(test *testing.T) {
	fmt.Println(gobgpapi.Family_AFI_IP.String())
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
