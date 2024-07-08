package linux

import (
	"net"
	"time"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/counter"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/datapath"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/datapath/types"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/lock"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/logging"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/logging/logfields"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/maps/tunnel"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/option"

	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

var (
	log = logging.DefaultLogger.WithField(logfields.LogSubsys, "linux-datapath")
)

type linuxNodeHandler struct {
	mutex                lock.Mutex
	isInitialized        bool
	nodeConfig           datapath.LocalNodeConfiguration
	nodeAddressing       types.NodeAddressing
	datapathConfig       DatapathConfiguration
	nodes                map[nodeTypes.Identity]*nodeTypes.Node
	enableNeighDiscovery bool
	neighLock            lock.Mutex // protects neigh* fields below
	neighDiscoveryLinks  []netlink.Link
	neighNextHopByNode4  map[nodeTypes.Identity]map[string]string // val = (key=link, value=string(net.IP))
	neighNextHopByNode6  map[nodeTypes.Identity]map[string]string // val = (key=link, value=string(net.IP))
	// All three mappings below hold both IPv4 and IPv6 entries.
	neighNextHopRefCount   counter.StringCounter
	neighByNextHop         map[string]*netlink.Neigh // key = string(net.IP)
	neighLastPingByNextHop map[string]time.Time      // key = string(net.IP)
	wgAgent                datapath.WireguardAgent

	ipsecMetricCollector prometheus.Collector
}

func (n *linuxNodeHandler) NodeAdd(newNode nodeTypes.Node) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	n.nodes[newNode.Identity()] = &newNode

	if n.isInitialized {
		return n.nodeUpdate(nil, &newNode, true)
	}

	return nil
}

func (n *linuxNodeHandler) NodeUpdate(oldNode, newNode nodeTypes.Node) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	n.nodes[newNode.Identity()] = &newNode

	if n.isInitialized {
		return n.nodeUpdate(&oldNode, &newNode, false)
	}

	return nil
}

// Must be called with linuxNodeHandler.mutex held.
func (n *linuxNodeHandler) nodeUpdate(oldNode, newNode *nodeTypes.Node, firstAddition bool) error {
	var (
		oldIP4Cidr, oldIP6Cidr                   *cidr.CIDR
		oldAllIP4AllocCidrs, oldAllIP6AllocCidrs []*cidr.CIDR
		newAllIP4AllocCidrs                      = newNode.GetIPv4AllocCIDRs()
		oldIP4, oldIP6                           net.IP
		newIP4                                   = newNode.GetNodeIP(false)
		newIP6                                   = newNode.GetNodeIP(true)
		oldKey, newKey                           uint8
		isLocalNode                              = false
	)

	if oldNode != nil {
		oldIP4Cidr = oldNode.IPv4AllocCIDR
		oldAllIP4AllocCidrs = oldNode.GetIPv4AllocCIDRs()
		oldIP4 = oldNode.GetNodeIP(false)
		oldKey = oldNode.EncryptionKey
	}

	if n.nodeConfig.EnableIPSec && !n.subnetEncryption() && !n.nodeConfig.EncryptNode {
		n.enableIPsec(newNode)
		newKey = newNode.EncryptionKey
	}

	if n.enableNeighDiscovery && !newNode.IsLocal() {
		// Running insertNeighbor in a separate goroutine relies on the following
		// assumptions:
		// 1. newNode is accessed only by reads.
		// 2. It is safe to invoke insertNeighbor for the same node.
		go n.insertNeighbor(context.Background(), newNode, false)
	}

	if n.nodeConfig.EnableIPSec && n.nodeConfig.EncryptNode && !n.subnetEncryption() {
		n.encryptNode(newNode)
	}

	if newNode.IsLocal() {
		isLocalNode = true
		if n.nodeConfig.EnableLocalNodeRoute {
			n.updateOrRemoveNodeRoutes(oldAllIP4AllocCidrs, newAllIP4AllocCidrs, isLocalNode)
			n.updateOrRemoveNodeRoutes(oldAllIP6AllocCidrs, newAllIP6AllocCidrs, isLocalNode)
		}
		if n.subnetEncryption() {
			n.enableSubnetIPsec(n.nodeConfig.IPv4PodSubnets, n.nodeConfig.IPv6PodSubnets)
		}
		if firstAddition && n.nodeConfig.EnableIPSec {
			metrics.Register(n.ipsecMetricCollector)
		}
		return nil
	}

	if option.Config.EnableWireguard && newNode.WireguardPubKey != "" {
		if err := n.wgAgent.UpdatePeer(newNode.Name, newNode.WireguardPubKey, newIP4, newIP6); err != nil {
			log.WithError(err).
				WithField(logfields.NodeName, newNode.Name).
				Warning("Failed to update wireguard configuration for peer")
		}
	}

	if n.nodeConfig.EnableAutoDirectRouting {
		n.updateDirectRoutes(oldAllIP4AllocCidrs, newAllIP4AllocCidrs, oldIP4, newIP4, firstAddition, n.nodeConfig.EnableIPv4)
		return nil
	}

	if n.nodeConfig.EnableEncapsulation {
		// Update the tunnel mapping of the node. In case the
		// node has changed its CIDR range, a new entry in the
		// map is created and the old entry is removed.
		/**
		 * update cilium_tunnel_map
		 * root@minikube:/home/cilium# cilium bpf tunnel list
		 * TUNNEL         VALUE
		 * 10.244.1.0:0   192.168.49.3:0
		 */
		updateTunnelMapping(oldIP4Cidr, newNode.IPv4AllocCIDR, oldIP4, newIP4, firstAddition, n.nodeConfig.EnableIPv4, oldKey, newKey)
		if !n.nodeConfig.UseSingleClusterRoute {
			n.updateOrRemoveNodeRoutes(oldAllIP4AllocCidrs, newAllIP4AllocCidrs, isLocalNode)
		}

		return nil
	} else if firstAddition {
		for _, ipv4AllocCIDR := range newAllIP4AllocCidrs {
			if rt, _ := n.lookupNodeRoute(ipv4AllocCIDR, isLocalNode); rt != nil {
				n.deleteNodeRoute(ipv4AllocCIDR, isLocalNode)
			}
		}
	}

	return nil
}

// updateTunnelMapping is called when a node update is received while running
// with encapsulation mode enabled. The CIDR and IP of both the old and new
// node are provided as context. The caller expects the tunnel mapping in the
// datapath to be updated.
func updateTunnelMapping(oldCIDR, newCIDR *cidr.CIDR, oldIP, newIP net.IP, firstAddition, encapEnabled bool, oldEncryptKey, newEncryptKey uint8) {
	if !encapEnabled {
		// When the protocol family is disabled, the initial node addition will
		// trigger a deletion to clean up leftover entries. The deletion happens
		// in quiet mode as we don't know whether it exists or not
		if newCIDR != nil && firstAddition {
			deleteTunnelMapping(newCIDR, true)
		}

		return
	}

	if cidrNodeMappingUpdateRequired(oldCIDR, newCIDR, oldIP, newIP, oldEncryptKey, newEncryptKey) {
		log.WithFields(logrus.Fields{
			logfields.IPAddr: newIP,
			"allocCIDR":      newCIDR,
		}).Debug("Updating tunnel map entry")

		if err := tunnel.TunnelMap.SetTunnelEndpoint(newEncryptKey, newCIDR.IP, newIP); err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"allocCIDR": newCIDR,
			}).Error("bpf: Unable to update in tunnel endpoint map")
		}
	}

	// Determine whether an old tunnel mapping must be cleaned up. The
	// below switch lists all conditions in which case the oldCIDR must be
	// removed from the tunnel mapping
	switch {
	// CIDR no longer announced
	case newCIDR == nil && oldCIDR != nil:
		fallthrough
	// Node allocation CIDR has changed
	case oldCIDR != nil && newCIDR != nil && !oldCIDR.Equal(newCIDR):
		deleteTunnelMapping(oldCIDR, false)
	}
}

func deleteTunnelMapping(oldCIDR *cidr.CIDR, quietMode bool) {
	if oldCIDR == nil {
		return
	}

	log.WithFields(logrus.Fields{
		"allocCIDR": oldCIDR,
		"quietMode": quietMode,
	}).Debug("Deleting tunnel map entry")

	if !quietMode {
		if err := tunnel.TunnelMap.DeleteTunnelEndpoint(oldCIDR.IP); err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"allocCIDR": oldCIDR,
			}).Error("Unable to delete in tunnel endpoint map")
		}
	} else {
		_ = tunnel.TunnelMap.SilentDeleteTunnelEndpoint(oldCIDR.IP)
	}
}
