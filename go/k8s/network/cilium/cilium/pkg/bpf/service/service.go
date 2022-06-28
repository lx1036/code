package service

import (
	"fmt"
	"github.com/cilium/cilium/pkg/logging/logfields"
	"sync"

	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf/maps/lbmap"
)

// INFO: 4 个 bpf 文件，cilium_lb4_services_v2/cilium_lb4_backends/cilium_lb4_reverse_nat/cilium_lb4_reverse_sk

// ServiceBPFManager reflect service-related changes into BPF maps by datapath BPF program.
// The changes can be triggered either by k8s_watcher or directly by API calls to the /services endpoint.
type ServiceBPFManager struct {
	sync.RWMutex

	lbmap *lbmap.LBBPFMap
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

func (s *ServiceBPFManager) UpdateOrInsertService() {
	s.Lock()
	defer s.Unlock()

	// Update lbmaps (BPF service maps)
	if err = s.updateOrInsertServiceIntoLBMaps(svc, onlyLocalBackends, prevBackendCount, newBackends,
		obsoleteBackendIDs,
		prevSessionAffinity, obsoleteSVCBackendIDs,
		scopedLog); err != nil {
		return false, lb.ID(0), err
	}

}

func (s *ServiceBPFManager) updateOrInsertServiceIntoLBMaps() {

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
func (s *ServiceBPFManager) DeleteService(frontend lb.L3n4Addr) (bool, error) {
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
		backendIDs := make([]lb.BackendID, 0, len(svc.backends))
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

func (s *ServiceBPFManager) notifyMonitorServiceDelete(id lb.ID) {

}
