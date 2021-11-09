package routing

import (
	"net"

	"github.com/vishvananda/netlink"
	"k8s.io/klog/v2"
)

// setupOverlayTunnel attempts to create an tunnel link and corresponding routes for IPIP based overlay networks
func (controller *NetworkRoutingController) setupOverlayTunnel(tunnelName string, nextHop net.IP) (netlink.Link, error) {

}

// cleanupTunnel removes any traces of tunnels / routes that were setup by nrc.setupOverlayTunnel() and are no longer
// needed. All errors are logged only, as we want to attempt to perform all cleanup actions regardless of their success
func (controller *NetworkRoutingController) cleanupTunnel(destinationSubnet *net.IPNet, tunnelName string) {
	klog.Infof("Cleaning up old routes for %s if there are any", destinationSubnet.String())
	if err := deleteRoutesByDestination(destinationSubnet); err != nil {
		klog.Errorf("Failed to cleanup routes: %v", err)
	}

	klog.Infof("Cleaning up any lingering tunnel interfaces named: %s", tunnelName)
	if link, err := netlink.LinkByName(tunnelName); err == nil {
		if err = netlink.LinkDel(link); err != nil {
			klog.Errorf("Failed to delete tunnel link for the node due to " + err.Error())
		}
	}
}
