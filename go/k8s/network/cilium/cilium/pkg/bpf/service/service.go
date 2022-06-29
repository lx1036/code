package service

import (
	"fmt"
	"github.com/cilium/cilium/pkg/counter"
	"github.com/cilium/cilium/pkg/logging/logfields"
	log "github.com/sirupsen/logrus"
	"sync"

	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf/maps/lbmap"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/k8s/loadbalancer"
)

// INFO: 4 个 bpf 文件，cilium_lb4_services_v2/cilium_lb4_backends/cilium_lb4_reverse_nat/cilium_lb4_reverse_sk

type svcInfo struct {
	hash          string
	frontend      loadbalancer.L3n4AddrID
	backends      []loadbalancer.Backend
	backendByHash map[string]*loadbalancer.Backend

	svcType                   loadbalancer.SVCType
	svcTrafficPolicy          loadbalancer.SVCTrafficPolicy
	sessionAffinity           bool
	sessionAffinityTimeoutSec uint32
	svcHealthCheckNodePort    uint16
	svcName                   string
	svcNamespace              string

	restoredFromDatapath bool
}

// ServiceBPFManager reflect service-related changes into BPF maps by datapath BPF program.
// The changes can be triggered either by k8s_watcher or directly by API calls to the /services endpoint.
type ServiceBPFManager struct {
	sync.RWMutex

	lbmap *lbmap.LBBPFMap

	backendByHash   map[string]*loadbalancer.Backend
	backendRefCount counter.StringCounter

	svcByHash map[string]*svcInfo
	svcByID   map[loadbalancer.ID]*svcInfo
}

func NewServiceBPFManager(monitorNotify monitorNotify) *ServiceBPFManager {

	return &ServiceBPFManager{

		lbmap: &lbmap.LBBPFMap{},
	}

}

// InitMaps opens or creates BPF maps used by services.
//
// If restore is set to false, entries of the maps are removed.
func (s *ServiceBPFManager) InitMaps(ipv6, ipv4, sockMaps, restore bool) error {
	s.Lock()
	defer s.Unlock()

	// The following two calls can be removed in v1.8+.
	if err := bpf.UnpinMapIfExists("cilium_lb6_rr_seq_v2"); err != nil {
		return nil
	}
	if err := bpf.UnpinMapIfExists("cilium_lb4_rr_seq_v2"); err != nil {
		return nil
	}

	toOpen := []*bpf.Map{}
	toDelete := []*bpf.Map{}
	if ipv4 {
		toOpen = append(toOpen, lbmap.Service4MapV2, lbmap.Backend4Map, lbmap.RevNat4Map)
		if !restore {
			toDelete = append(toDelete, lbmap.Service4MapV2, lbmap.Backend4Map, lbmap.RevNat4Map)
		}
		if sockMaps {
			if err := lbmap.CreateSockRevNat4Map(); err != nil {
				return err
			}
		}
	}

	for _, m := range toOpen {
		if _, err := m.OpenOrCreate(); err != nil {
			return err
		}
	}
	for _, m := range toDelete {
		if err := m.DeleteAll(); err != nil {
			return err
		}
	}

	return nil
}

// RestoreServices restores services from BPF maps.
//
// The method should be called once before establishing a connectivity
// to kube-apiserver.
func (s *ServiceBPFManager) RestoreServices() error {
	s.Lock()
	defer s.Unlock()

	// Restore backend IDs
	if err := s.restoreBackendsLocked(); err != nil {
		return err
	}

	// Restore service cache from BPF maps
	if err := s.restoreServicesLocked(); err != nil {
		return err
	}

	// Remove no longer existing affinity matches
	if err := s.deleteOrphanAffinityMatchesLocked(); err != nil {
		return err
	}

	// Remove obsolete backends and release their IDs
	if err := s.deleteOrphanBackends(); err != nil {
		log.WithError(err).Warn("Failed to remove orphan backends")
	}

	return nil
}

func (s *ServiceBPFManager) restoreBackendsLocked() error {
	backends, err := s.lbmap.DumpBackendMaps()
	if err != nil {
		return fmt.Errorf("Unable to dump backend maps: %s", err)
	}

	for _, b := range backends {
		if err := RestoreBackendID(b.L3n4Addr, b.ID); err != nil {
			return fmt.Errorf("unable to restore backend ID %d for %q: %s", b.ID, b.L3n4Addr, err)
		}

		hash := b.L3n4Addr.Hash()
		s.backendByHash[hash] = b
	}

	return nil
}

func (s *ServiceBPFManager) restoreServicesLocked() error {
	failed, restored := 0, 0

	svcs, errors := s.lbmap.DumpServiceMaps()
	for _, err := range errors {
		log.WithError(err).Warning("Error occurred while dumping service maps")
	}

	for _, svc := range svcs {
		scopedLog := log.WithFields(logrus.Fields{
			"ServiceID": svc.Frontend.ID,
			"ServiceIP": svc.Frontend.L3n4Addr.String(),
		})
		scopedLog.Debug("Restoring service")

		if _, err := RestoreID(svc.Frontend.L3n4Addr, uint32(svc.Frontend.ID)); err != nil {
			failed++
			scopedLog.WithError(err).Warning("Unable to restore service ID")
		}

		newSVC := &svcInfo{
			hash:          svc.Frontend.Hash(),
			frontend:      svc.Frontend,
			backends:      svc.Backends,
			backendByHash: map[string]*loadbalancer.Backend{},
			// Correct traffic policy will be restored by k8s_watcher after k8s
			// service cache has been initialized
			svcType:          svc.Type,
			svcTrafficPolicy: svc.TrafficPolicy,

			sessionAffinity:           svc.SessionAffinity,
			sessionAffinityTimeoutSec: svc.SessionAffinityTimeoutSec,

			// Indicate that the svc was restored from the BPF maps, so that
			// SyncWithK8sFinished() could remove services which were restored
			// from the maps but not present in the k8sServiceCache (e.g. a svc
			// was deleted while cilium-agent was down).
			restoredFromDatapath: true,
		}

		for j, backend := range svc.Backends {
			hash := backend.L3n4Addr.Hash()
			s.backendRefCount.Add(hash)
			newSVC.backendByHash[hash] = &svc.Backends[j]
		}

		s.svcByHash[newSVC.hash] = newSVC
		s.svcByID[newSVC.frontend.ID] = newSVC
		restored++
	}

	log.WithFields(log.Fields{
		"restored": restored,
		"failed":   failed,
	}).Info("Restored services from maps")

	return nil
}

// deleteOrphanAffinityMatchesLocked removes affinity matches which point to
// non-existent svc ID and backend ID tuples.
func (s *ServiceBPFManager) deleteOrphanAffinityMatchesLocked() error {
	matches, err := s.lbmap.DumpAffinityMatches()
	if err != nil {
		return err
	}

	toRemove := map[loadbalancer.ID][]loadbalancer.BackendID{}

	local := make(map[loadbalancer.ID]map[loadbalancer.BackendID]struct{}, len(s.svcByID))
	for id, svc := range s.svcByID {
		if !svc.sessionAffinity {
			continue
		}
		local[id] = make(map[loadbalancer.BackendID]struct{}, len(svc.backends))
		for _, backend := range svc.backends {
			local[id][backend.ID] = struct{}{}
		}
	}

	for svcID, backendIDs := range matches {
		for bID := range backendIDs {
			found := false
			if _, ok := local[loadbalancer.ID(svcID)]; ok {
				if _, ok := local[loadbalancer.ID(svcID)][loadbalancer.BackendID(bID)]; ok {
					found = true
				}
			}
			if !found {
				toRemove[loadbalancer.ID(svcID)] = append(toRemove[loadbalancer.ID(svcID)], loadbalancer.BackendID(bID))
			}
		}
	}

	for svcID, backendIDs := range toRemove {
		s.deleteBackendsFromAffinityMatchMap(svcID, backendIDs)
	}

	return nil
}

func (s *ServiceBPFManager) deleteBackendsFromAffinityMatchMap(svcID loadbalancer.ID, backendIDs []loadbalancer.BackendID) {
	log.WithFields(log.Fields{
		logfields.Backends:  backendIDs,
		logfields.ServiceID: svcID,
	}).Debug("Deleting backends from session affinity match")

	for _, bID := range backendIDs {
		if err := s.lbmap.DeleteAffinityMatch(uint16(svcID), uint16(bID)); err != nil {
			log.WithFields(log.Fields{
				logfields.BackendID: bID,
				logfields.ServiceID: svcID,
			}).WithError(err).Warn("Unable to remove entry from affinity match map")
		}
	}
}

func (s *ServiceBPFManager) deleteOrphanBackends() error {
	for hash, b := range s.backendByHash {
		if s.backendRefCount[hash] == 0 {
			log.WithField(logfields.BackendID, b.ID).Debug("Removing orphan backend")

			DeleteBackendID(b.ID)
			if err := s.lbmap.DeleteBackendByID(uint16(b.ID), false); err != nil {
				return fmt.Errorf("unable to remove backend %d from map: %s", b.ID, err)
			}
			delete(s.backendByHash, hash)
		}
	}

	return nil
}

func (s *ServiceBPFManager) UpdateOrInsertService(
	frontend loadbalancer.L3n4AddrID, backends []loadbalancer.Backend, svcType loadbalancer.SVCType,
	svcTrafficPolicy loadbalancer.SVCTrafficPolicy,
	sessionAffinity bool, sessionAffinityTimeoutSec uint32,
	svcHealthCheckNodePort uint16,
	svcName, svcNamespace string) (bool, loadbalancer.ID, error) {
	s.Lock()
	defer s.Unlock()

	// If needed, create svcInfo and allocate service ID
	svc, ok, prevSessionAffinity, err := s.createSVCInfoIfNotExist(frontend, svcType, svcTrafficPolicy,
		sessionAffinity, sessionAffinityTimeoutSec,
		svcHealthCheckNodePort, svcName, svcNamespace)
	if err != nil {
		return false, loadbalancer.ID(0), err
	}

	// Update lbmaps (BPF service maps)
	if err = s.updateOrInsertServiceIntoLBMaps(svc, onlyLocalBackends, prevBackendCount, newBackends,
		obsoleteBackendIDs, prevSessionAffinity, obsoleteSVCBackendIDs); err != nil {
		return false, loadbalancer.ID(0), err
	}

	if ok {
		addMetric.Inc()
	} else {
		updateMetric.Inc()
	}

	s.notifyMonitorServiceUpsert(svc.frontend, svc.backends,
		svc.svcType, svc.svcTrafficPolicy, svc.svcName, svc.svcNamespace)

	return ok, loadbalancer.ID(svc.frontend.ID), nil
}

func (s *ServiceBPFManager) createSVCInfoIfNotExist(
	frontend loadbalancer.L3n4AddrID,
	svcType loadbalancer.SVCType,
	svcTrafficPolicy loadbalancer.SVCTrafficPolicy,
	sessionAffinity bool, sessionAffinityTimeoutSec uint32,
	svcHealthCheckNodePort uint16,
	svcName, svcNamespace string,
) (*svcInfo, bool, bool, error) {

}

func (s *ServiceBPFManager) updateOrInsertServiceIntoLBMaps(svc *svcInfo, onlyLocalBackends bool,
	prevBackendCount int, newBackends []loadbalancer.Backend, obsoleteBackendIDs []loadbalancer.BackendID,
	prevSessionAffinity bool, obsoleteSVCBackendIDs []loadbalancer.BackendID) error {

	err := s.lbmap.UpdateOrInsertService(
		uint16(svc.frontend.ID), svc.frontend.L3n4Addr.IP,
		svc.frontend.L3n4Addr.L4Addr.Port,
		backendIDs, prevBackendCount,
		ipv6, svc.svcType, onlyLocalBackends,
		svc.frontend.L3n4Addr.Scope,
		svc.sessionAffinity, svc.sessionAffinityTimeoutSec)
	if err != nil {
		return err
	}

}

// DeleteService removes the given service.
func (s *ServiceBPFManager) DeleteService(frontend loadbalancer.L3n4Addr) (bool, error) {
	s.Lock()
	defer s.Unlock()

	if svc, found := s.svcByHash[frontend.Hash()]; found {
		return true, s.deleteServiceLocked(svc)
	}

	return false, nil
}

func (s *ServiceBPFManager) deleteServiceLocked(svc *svcInfo) error {
	obsoleteBackendIDs := s.deleteBackendsFromCacheLocked(svc)

	if err := s.lbmap.DeleteService(svc.frontend, len(svc.backends)); err != nil {
		return err
	}

	// Delete affinity matches
	if svc.sessionAffinity {
		backendIDs := make([]loadbalancer.BackendID, 0, len(svc.backends))
		for _, b := range svc.backends {
			backendIDs = append(backendIDs, b.ID)
		}
		s.deleteBackendsFromAffinityMatchMap(svc.frontend.ID, backendIDs)
	}

	delete(s.svcByHash, svc.hash)
	delete(s.svcByID, svc.frontend.ID)

	ipv6 := svc.frontend.L3n4Addr.IsIPv6()
	for _, id := range obsoleteBackendIDs {
		scopedLog.WithField(logfields.BackendID, id).
			Debug("Deleting obsolete backend")

		if err := s.lbmap.DeleteBackendByID(uint16(id), ipv6); err != nil {
			return err
		}
	}
	if err := DeleteID(uint32(svc.frontend.ID)); err != nil {
		return fmt.Errorf("Unable to release service ID %d: %s", svc.frontend.ID, err)
	}

	deleteMetric.Inc()
	s.notifyMonitorServiceDelete(svc.frontend.ID)

	return nil
}

func (s *ServiceBPFManager) notifyMonitorServiceDelete(id loadbalancer.ID) {

}

func (s *ServiceBPFManager) notifyMonitorServiceUpsert(frontend loadbalancer.L3n4AddrID, backends []loadbalancer.Backend,
	svcType loadbalancer.SVCType, svcTrafficPolicy loadbalancer.SVCTrafficPolicy, svcName, svcNamespace string) {
	if s.monitorNotify == nil {
		return
	}

}
