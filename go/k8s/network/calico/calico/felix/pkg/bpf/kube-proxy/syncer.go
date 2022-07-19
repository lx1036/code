package proxy

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"

	"k8s-lx1036/k8s/network/calico/calico/felix/pkg/bpf"
	"k8s-lx1036/k8s/network/calico/calico/felix/pkg/bpf/maps/cachingmap"
	"k8s-lx1036/k8s/network/calico/calico/felix/pkg/bpf/maps/nat"
)

type Syncer struct {

	// synced is true after reconciling the first Apply
	synced bool

	bpfSvcs     *cachingmap.CachingMap
	bpfEps      *cachingmap.CachingMap
	bpfAffinity bpf.Map
}

func NewSyncer(nodePortIPs []net.IP, svcsmap, epsmap *cachingmap.CachingMap, affmap bpf.Map,
	rt Routes) (*Syncer, error) {

	s := &Syncer{
		bpfSvcs:     svcsmap,
		bpfEps:      epsmap,
		bpfAffinity: affmap,
	}

	err := s.bpfSvcs.LoadCacheFromDataplane()
	if err != nil {
		return nil, err
	}
	err = s.bpfEps.LoadCacheFromDataplane()
	if err != nil {
		return nil, err
	}

	return s, nil
}

// Apply applies the new state
func (s *Syncer) Apply(state DPSyncerState) error {
	if !s.synced {
		log.Infof("Loading BPF map state from dataplane")
		if err := s.startupSync(state); err != nil {
			return fmt.Errorf(fmt.Sprintf("startup sync err: %v", err))
		}
		log.Infof("Loaded BPF map state from dataplane")
		s.mapsLock.Lock()
	} else {
		// if we were not synced yet, the fixer cannot run yet
		//s.StopExpandNPFixup()

		s.mapsLock.Lock()
		s.prevSvcMap = s.newSvcMap
		s.prevEpsMap = s.newEpsMap
	}

}

func (s *Syncer) startupSync(state DPSyncerState) error {

	s.bpfSvcs.IterDataplaneCache(func(k, v []byte) {
		var svck nat.FrontendKey
		var svcv nat.FrontendValue
		copy(svck[:], k)
		copy(svcv[:], v)

		xref, ok := svcRef[svck]
		if !ok {
			return
		}

	})

}
