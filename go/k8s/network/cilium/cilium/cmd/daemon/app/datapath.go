package app

import (
	"fmt"
	"github.com/cilium/cilium/pkg/controller"
	"github.com/cilium/cilium/pkg/identity"
	"github.com/cilium/cilium/pkg/logging/logfields"
	"github.com/cilium/cilium/pkg/node"
	"github.com/cilium/cilium/pkg/source"
	log "github.com/sirupsen/logrus"
	"k8s.io/klog/v2"
	"net"
	"os"
	"time"

	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf/ipcache"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf/maps/ctmap"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf/maps/fragmap"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf/maps/lbmap"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf/maps/lxcmap"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf/maps/metricsmap"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf/maps/nat"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf/maps/neighborsmap"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf/maps/policymap"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/config/option"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/datapath"
)

func (d *Daemon) Datapath() datapath.Datapath {
	return d.datapath
}

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

	// INFO: init cilium_call_policy BPF maps
	if err := policymap.InitCallMap(); err != nil {
		return err
	}

	// INFO: init endpoint bpf maps
	for _, ep := range d.endpointManager.GetEndpoints() {
		ep.InitPolicyMap()
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
	for _, m := range ctmap.GlobalMaps(option.Config.EnableIPv4, option.Config.EnableIPv6) {
		if _, err := m.Create(); err != nil {
			return err
		}
	}

	ipv4Nat, _ := nat.GlobalMaps(option.Config.EnableIPv4, option.Config.EnableIPv6)
	if option.Config.EnableIPv4 {
		if _, err := ipv4Nat.Create(); err != nil {
			return err
		}
	}

	if option.Config.EnableNodePort {
		if err := neighborsmap.InitMaps(option.Config.EnableIPv4, option.Config.EnableIPv6); err != nil {
			return err
		}
	}

	if option.Config.EnableIPv4FragmentsTracking {
		if err := fragmap.InitMap(option.Config.FragmentsMapEntries); err != nil {
			return err
		}
	}

	// Set up the list of IPCache listeners in the daemon, to be
	// used by syncEndpointsAndHostIPs()
	// xDS cache will be added later by calling AddListener(), but only if necessary.
	ipcache.IPIdentityCache.SetListeners([]ipcache.IPIdentityMappingListener{
		datapathIpcache.NewListener(d, d),
	})

	// Start the controller for periodic sync of the metrics map with
	// the prometheus server.
	controller.NewManager().UpdateController("metricsmap-bpf-prom-sync",
		controller.ControllerParams{
			DoFunc:      metricsmap.SyncMetricsMap,
			RunInterval: 5 * time.Second,
			Context:     d.ctx,
		})

	if !option.Config.RestoreState {
		// If we are not restoring state, all endpoints can be
		// deleted. Entries will be re-populated.
		lxcmap.LXCMap.DeleteAll()
	}

	if option.Config.EnableSessionAffinity {
		if _, err := lbmap.AffinityMatchMap.OpenOrCreate(); err != nil {
			return err
		}
		if option.Config.EnableIPv4 {
			if _, err := lbmap.Affinity4Map.OpenOrCreate(); err != nil {
				return err
			}
		}
	}

	return nil
}

// syncLXCMap adds local host enties to bpf lxcmap, as well as
// ipcache, if needed, and also notifies the daemon and network policy
// hosts cache if changes were made.
func (d *Daemon) syncEndpointsAndHostIPs() error {
	// INFO: netlink 获取本机器的所有真实网卡地址，如 eth0/eth1 等地址，和 cilium_host 虚拟网卡地址
	addrs, err := d.datapath.LocalNodeAddressing().IPv4().LocalAddresses()
	if err != nil {
		log.WithError(err).Warning("Unable to list local IPv4 addresses")
	}

	var specialIdentities []identity.IPIdentityPair
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
		identity.IPIdentityPair{ // "0.0.0.0/0"
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
				return fmt.Errorf("unable to add host entry to endpoint map: %s", err)
			}
			if added {
				log.WithField(logfields.IPAddr, ipIDPair.IP).Debugf("Added local ip to endpoint map")
			}
		}

		delete(existingEndpoints, ipIDPair.IP.String())

		// Upsert will not propagate (reserved:foo->ID) mappings across the cluster,
		// and we specifically don't want to do so.
		ipcache.IPIdentityCache.UpdateOrInsert(ipIDPair.PrefixString(), nil, hostKey, nil, ipcache.Identity{
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

// INFO: 把当前 node 的配置写入 /var/run/cilium/state/globals/node_config.h
func (d *Daemon) createNodeConfigHeaderfile() error {
	nodeConfigPath := option.Config.GetNodeConfigPath() // /var/run/cilium/state/globals/node_config.h
	f, err := os.Create(nodeConfigPath)
	if err != nil {
		log.WithError(err).WithField(logfields.Path, nodeConfigPath).Fatal("Failed to create node configuration file")
		return err
	}
	defer f.Close()

	if err = d.datapath.WriteNodeConfig(f, &d.nodeDiscovery.LocalConfig); err != nil {
		log.WithError(err).WithField(logfields.Path, nodeConfigPath).Fatal("Failed to write node configuration file")
		return err
	}
	return nil
}
