package app

import (
	"context"
	"github.com/cilium/cilium/pkg/controller"
	"github.com/cilium/cilium/pkg/k8s"
	"io/ioutil"
	"net"

	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf/endpoint"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf/maps/lxcmap"

	log "github.com/sirupsen/logrus"
	"k8s.io/klog/v2"
)

type endpointRestoreState struct {
	restored []*endpoint.Endpoint
	toClean  []*endpoint.Endpoint
}

// restoreOldEndpoints reads the list of existing endpoints previously managed
// Cilium when it was last run and associated it with container workloads. This
// function performs the first step in restoring the endpoint structure,
// allocating their existing IP out of the CIDR block and then inserting the
// endpoints into the endpoints list. It needs to be followed by a call to
// regenerateRestoredEndpoints() once the endpoint builder is ready.
//
// If clean is true, endpoints which cannot be associated with a container
// workloads are deleted.
// dir="/var/run/cilium/state"
func (d *Daemon) restoreOldEndpoints(dir string, clean bool) (*endpointRestoreState, error) {
	state := &endpointRestoreState{
		restored: []*endpoint.Endpoint{},
		toClean:  []*endpoint.Endpoint{},
	}

	if !option.Config.RestoreState {
		klog.Info("Endpoint restore is disabled, skipping restore step")
		return state, nil
	}

	klog.Info("Restoring endpoints...")

	existingEndpoints, err := lxcmap.DumpToMap()
	if err != nil {
		log.WithError(err).Warning("Unable to open endpoint map while restoring. Skipping cleanup of endpoint map on startup")
	}

	dirFiles, err := ioutil.ReadDir(dir)
	if err != nil {
		return state, err
	}
	endpointsID := endpoint.FilterEPDir(dirFiles)
	possibleEPs := endpoint.ReadEPsFromDirNames(d.ctx, d, dir, endpointsID)
	if len(possibleEPs) == 0 {
		log.Info("No old endpoints found.")
		return state, nil
	}

	for _, ep := range possibleEPs {

		// We have to set the allocator for identities here during the Endpoint
		// lifecycle, because the identity allocator has be initialized *after*
		// endpoints are restored from disk. This is because we have to reserve
		// IPs for the endpoints that are restored via IPAM. Reserving of IPs
		// affects the allocation of IPs w.r.t. node addressing, which we need
		// to know before the identity allocator is initialized. We need to
		// know the node addressing because when adding a reference to the
		// kvstore because the local node's IP is used as a suffix for the key
		// in the key-value store.
		ep.SetAllocator(d.identityAllocator)

		restore, err := d.validateEndpoint(ep)
		if err != nil {
			// Disconnected EPs are not failures, clean them silently below
			if !ep.IsDisconnecting() {
				scopedLog.WithError(err).Warningf("Unable to restore endpoint, ignoring")
				failed++
			}
		}
		if !restore {
			if clean {
				state.toClean = append(state.toClean, ep)
			}
			continue
		}

		ep.SetDefaultConfiguration(true)
		ep.SetProxy(d.l7Proxy)
		ep.SkipStateClean()

		state.restored = append(state.restored, ep)

		if existingEndpoints != nil {
			delete(existingEndpoints, ep.IPv4.String())
			delete(existingEndpoints, ep.IPv6.String())
		}
	}

	// delete non-exist endpoint from lxc BPF maps
	if existingEndpoints != nil {
		for hostIP, info := range existingEndpoints {
			if ip := net.ParseIP(hostIP); !info.IsHost() && ip != nil {
				if err := lxcmap.DeleteEntry(ip); err != nil {
					log.WithError(err).Warn("Unable to delete obsolete endpoint from BPF map")
				} else {
					log.Debugf("Removed outdated endpoint %d from endpoint map", info.LxcID)
				}
			}
		}
	}

	return state, nil
}

func (d *Daemon) initRestore(restoredEndpoints *endpointRestoreState) chan struct{} {
	var restoreComplete chan struct{}
	if option.Config.RestoreState {
		// When we regenerate restored endpoints, it is guaranteed tha we have
		// received the full list of policies present at the time the daemon
		// is bootstrapped.
		restoreComplete = d.regenerateRestoredEndpoints(restoredEndpoints)
		go func() {
			<-restoreComplete
			endParallelMapMode()
		}()

		go func() {
			if k8s.IsEnabled() {
				// Start controller which removes any leftover Kubernetes
				// services that may have been deleted while Cilium was not
				// running. Once this controller succeeds, because it has no
				// RunInterval specified, it will not run again unless updated
				// elsewhere. This means that if, for instance, a user manually
				// adds a service via the CLI into the BPF maps, that it will
				// not be cleaned up by the daemon until it restarts.
				controller.NewManager().UpdateController("sync-lb-maps-with-k8s-services",
					controller.ControllerParams{
						DoFunc: func(ctx context.Context) error {
							return d.serviceBPFManager.SyncWithK8sFinished()
						},
						Context: d.ctx,
					},
				)
			}
		}()
	} else {
		log.Info("State restore is disabled. Existing endpoints on node are ignored")

		// No restore happened, end parallel map mode immediately
		endParallelMapMode()
	}

	return restoreComplete
}
