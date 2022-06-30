package app

import (
	"fmt"
	"github.com/cilium/cilium/pkg/identity"
	"github.com/cilium/cilium/pkg/ipcache"
	"github.com/cilium/cilium/pkg/logging/logfields"
	"github.com/cilium/cilium/pkg/maps/ctmap"
	"github.com/cilium/cilium/pkg/node"
	"github.com/cilium/cilium/pkg/source"
	log "github.com/sirupsen/logrus"
	"k8s.io/klog/v2"
	"net"
	"os"

	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf/maps/lxcmap"
)

// initMaps opens all BPF maps (and creates them if they do not exist). This
// must be done *before* any operations which read BPF maps, especially
// restoring endpoints and services.
func (d *Daemon) initMaps() error {

	// Delete old maps if left over from an upgrade.
	for _, name := range []string{"cilium_proxy4", "cilium_proxy6", "cilium_policy"} {
		path := bpf.MapPath(name)
		if _, err := os.Stat(path); err == nil {
			if err = os.RemoveAll(path); err == nil {
				klog.Infof("removed legacy map file %s", path)
			}
		}
	}

	// INFO: init service bpf maps
	createSockRevNatMaps := true
	if err := d.serviceBPFManager.InitMaps(false, true, createSockRevNatMaps, true); err != nil {
		log.WithError(err).Fatal("Unable to initialize service maps")
	}

	// INFO: init endpoint bpf maps
	for _, ep := range d.endpointManager.GetEndpoints() {
		ep.InitMap()
	}
	for _, ep := range d.endpointManager.GetEndpoints() {
		if !ep.ConntrackLocal() {
			continue
		}
		for _, m := range ctmap.LocalMaps(ep, option.Config.EnableIPv4,
			option.Config.EnableIPv6) {
			if _, err := m.Create(); err != nil {
				return err
			}
		}
	}
	for _, m := range ctmap.GlobalMaps(option.Config.EnableIPv4,
		option.Config.EnableIPv6) {
		if _, err := m.Create(); err != nil {
			return err
		}
	}

}

// syncLXCMap adds local host enties to bpf lxcmap, as well as
// ipcache, if needed, and also notifies the daemon and network policy
// hosts cache if changes were made.
func (d *Daemon) syncEndpointsAndHostIPs() error {
	specialIdentities := []identity.IPIdentityPair{}

	addrs, err := d.datapath.LocalNodeAddressing().IPv4().LocalAddresses()
	if err != nil {
		log.WithError(err).Warning("Unable to list local IPv4 addresses")
	}

	for _, ip := range addrs {
		if len(ip) > 0 {
			specialIdentities = append(specialIdentities,
				identity.IPIdentityPair{
					IP: ip,
					ID: identity.ReservedIdentityHost,
				})
		}
	}

	specialIdentities = append(specialIdentities,
		identity.IPIdentityPair{
			IP:   net.IPv4zero,
			Mask: net.CIDRMask(0, net.IPv4len*8),
			ID:   identity.ReservedIdentityWorld,
		})

	existingEndpoints, err := lxcmap.DumpToMap()
	if err != nil {
		return err
	}

	for _, ipIDPair := range specialIdentities {
		hostKey := node.GetIPsecKeyIdentity()
		isHost := ipIDPair.ID == identity.ReservedIdentityHost
		if isHost {
			added, err := lxcmap.SyncHostEntry(ipIDPair.IP)
			if err != nil {
				return fmt.Errorf("Unable to add host entry to endpoint map: %s", err)
			}
			if added {
				log.WithField(logfields.IPAddr, ipIDPair.IP).Debugf("Added local ip to endpoint map")
			}
		}

		delete(existingEndpoints, ipIDPair.IP.String())

		// Upsert will not propagate (reserved:foo->ID) mappings across the cluster,
		// and we specifically don't want to do so.
		ipcache.IPIdentityCache.Upsert(ipIDPair.PrefixString(), nil, hostKey, nil, ipcache.Identity{
			ID:     ipIDPair.ID,
			Source: source.Local,
		})
	}

	for hostIP, info := range existingEndpoints {
		if ip := net.ParseIP(hostIP); info.IsHost() && ip != nil {
			if err := lxcmap.DeleteEntry(ip); err != nil {
				log.WithError(err).WithFields(log.Fields{
					logfields.IPAddr: hostIP,
				}).Warn("Unable to delete obsolete host IP from BPF map")
			} else {
				log.Debugf("Removed outdated host ip %s from endpoint map", hostIP)
			}

			ipcache.IPIdentityCache.Delete(hostIP, source.Local)
		}
	}

	return nil
}
