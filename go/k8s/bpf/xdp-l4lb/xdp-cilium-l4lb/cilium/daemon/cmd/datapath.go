package cmd

import (
	"github.com/cilium/cilium/pkg/controller"
	"github.com/cilium/cilium/pkg/datapath/linux/probes"
	"github.com/cilium/cilium/pkg/ipcache"
	"github.com/cilium/cilium/pkg/logging/logfields"
	"github.com/cilium/cilium/pkg/maps/egressmap"
	"github.com/cilium/cilium/pkg/maps/eventsmap"
	"github.com/cilium/cilium/pkg/maps/ipmasq"
	"github.com/cilium/cilium/pkg/maps/metricsmap"
	"github.com/cilium/cilium/pkg/maps/signalmap"
	"github.com/cilium/cilium/pkg/maps/tunnel"
	ipcachemap "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/maps/ipcache"
	"os"
	"time"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/bpf"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/datapath"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/maps/fragmap"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/maps/lbmap"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/maps/lxcmap"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/option"
)

// /var/run/cilium/state/globals/node_config.h
func (d *Daemon) createNodeConfigHeaderfile() error {
	nodeConfigPath := option.Config.GetNodeConfigPath()
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

func (d *Daemon) Datapath() datapath.Datapath {
	return d.datapath
}

// initMaps opens all BPF maps (and creates them if they do not exist). This
// must be done *before* any operations which read BPF maps, especially
// restoring endpoints and services.
func (d *Daemon) initMaps() error {
	if option.Config.DryMode {
		return nil
	}

	// Rename old policy call map to avoid packet drops during upgrade.
	// TODO: Remove this renaming step once Cilium 1.8 is the oldest supported
	// release.
	policyMapPath := bpf.MapPath("cilium_policy")
	if _, err := os.Stat(policyMapPath); err == nil {
		newPolicyMapPath := bpf.MapPath(policymap.PolicyCallMapName)
		if err = os.Rename(policyMapPath, newPolicyMapPath); err != nil {
			log.WithError(err).Fatalf("Failed to rename policy call map from %s to %s",
				policyMapPath, newPolicyMapPath)
		}
	}

	if _, err := lxcmap.LXCMap.OpenOrCreate(); err != nil {
		return err
	}

	// The ipcache is shared between endpoints. Parallel mode needs to be
	// used to allow existing endpoints that have not been regenerated yet
	// to continue using the existing ipcache until the endpoint is
	// regenerated for the first time. Existing endpoints are using a
	// policy map which is potentially out of sync as local identities are
	// re-allocated on startup. Parallel mode allows to continue using the
	// old version until regeneration. Note that the old version is not
	// updated with new identities. This is fine as any new identity
	// appearing would require a regeneration of the endpoint anyway in
	// order for the endpoint to gain the privilege of communication.
	if _, err := ipcachemap.IPCache.OpenParallel(); err != nil {
		return err
	}

	if err := metricsmap.Metrics.OpenOrCreate(); err != nil {
		return err
	}

	if option.Config.TunnelingEnabled() || option.Config.EnableEgressGateway {
		// The IPv4 egress gateway feature also uses tunnel map
		if _, err := tunnel.TunnelMap.OpenOrCreate(); err != nil {
			return err
		}
	}

	if option.Config.EnableEgressGateway {
		if err := egressmap.InitEgressMaps(); err != nil {
			return err
		}
	}

	pm := probes.NewProbeManager()
	supportedMapTypes := pm.GetMapTypes()
	createSockRevNatMaps := option.Config.EnableHostReachableServices &&
		option.Config.EnableHostServicesUDP && supportedMapTypes.HaveLruHashMapType
	if err := d.svc.InitMaps(option.Config.EnableIPv6, option.Config.EnableIPv4,
		createSockRevNatMaps, option.Config.RestoreState); err != nil {
		log.WithError(err).Fatal("Unable to initialize service maps")
	}

	possibleCPUs := common.GetNumPossibleCPUs(log)

	if err := eventsmap.InitMap(possibleCPUs); err != nil {
		return err
	}

	if err := signalmap.InitMap(possibleCPUs); err != nil {
		return err
	}

	if err := policymap.InitCallMap(); err != nil {
		return err
	}

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

	ipv4Nat, ipv6Nat := nat.GlobalMaps(option.Config.EnableIPv4,
		option.Config.EnableIPv6, option.Config.EnableNodePort)
	if ipv4Nat != nil {
		if _, err := ipv4Nat.Create(); err != nil {
			return err
		}
	}
	if ipv6Nat != nil {
		if _, err := ipv6Nat.Create(); err != nil {
			return err
		}
	}

	if option.Config.EnableNodePort {
		if err := neighborsmap.InitMaps(option.Config.EnableIPv4,
			option.Config.EnableIPv6); err != nil {
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

	if option.Config.EnableIPv4 && option.Config.EnableIPMasqAgent {
		if _, err := ipmasq.IPMasq4Map.OpenOrCreate(); err != nil {
			return err
		}
	}

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
		if option.Config.EnableIPv6 {
			if _, err := lbmap.Affinity6Map.OpenOrCreate(); err != nil {
				return err
			}
		}
	}

	if option.Config.EnableSVCSourceRangeCheck {
		if option.Config.EnableIPv4 {
			if _, err := lbmap.SourceRange4Map.OpenOrCreate(); err != nil {
				return err
			}
		}
		if option.Config.EnableIPv6 {
			if _, err := lbmap.SourceRange6Map.OpenOrCreate(); err != nil {
				return err
			}
		}
	}

	if option.Config.NodePortAlg == option.NodePortAlgMaglev {
		if err := lbmap.InitMaglevMaps(option.Config.EnableIPv4, option.Config.EnableIPv6); err != nil {
			return err
		}
	}

	return nil
}
