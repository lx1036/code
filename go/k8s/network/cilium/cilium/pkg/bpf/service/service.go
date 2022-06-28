package service

import (
	"sync"

	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf/service/lbmap"
)

// Service reflect service-related changes into BPF maps by datapath BPF program.
// The changes can be triggered either by k8s_watcher or directly by API calls to the /services endpoint.
type Service struct {
	sync.RWMutex

	lbmap *lbmap.LBBPFMap
}

func NewService(monitorNotify monitorNotify) *Service {

	return &Service{

		lbmap: &lbmap.LBBPFMap{},
	}

}

func (s *Service) UpdateOrInsertService() {
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

func (s *Service) updateOrInsertServiceIntoLBMaps() {

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
