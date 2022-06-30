package app

import (
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
