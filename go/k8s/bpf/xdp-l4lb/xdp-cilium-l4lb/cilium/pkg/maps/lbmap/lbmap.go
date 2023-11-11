package lbmap

import (
	"errors"
	"fmt"
	"net"

	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/loadbalancer"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/logging"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/logging/logfields"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/option"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/u8proto"
)

var log = logging.DefaultLogger.WithField(logfields.LogSubsys, "map-lb")

var (
	// MaxEntries contains the maximum number of entries that are allowed
	// in Cilium LB service, backend and affinity maps.
	MaxEntries = 65536
)

type InitParams struct {
	IPv4, IPv6 bool

	MaxSockRevNatMapEntries, MaxEntries int
}

func Init(params InitParams) {
	if params.MaxSockRevNatMapEntries != 0 {
		MaxSockRevNat4MapEntries = params.MaxSockRevNatMapEntries
		//MaxSockRevNat6MapEntries = params.MaxSockRevNatMapEntries
	}

	if params.MaxEntries != 0 {
		MaxEntries = params.MaxEntries
	}

	initSVC(params)
	initAffinity(params)
	initSourceRange(params)
}

type UpsertServiceParams struct {
	ID                        uint16
	IP                        net.IP
	Port                      uint16
	Backends                  map[string]uint16
	PrevBackendCount          int
	IPv6                      bool
	Type                      loadbalancer.SVCType
	Local                     bool
	Scope                     uint8
	SessionAffinity           bool
	SessionAffinityTimeoutSec uint32
	CheckSourceRange          bool
	UseMaglev                 bool
}

// LBBPFMap is an implementation of the LBMap interface.
type LBBPFMap struct {
	// Buffer used to avoid excessive allocations to temporarily store backend
	// IDs. Concurrent access is protected by the
	// pkg/service.go:(Service).UpsertService() lock.
	maglevBackendIDsBuffer []uint16
	maglevTableSize        uint64
}

func New(maglev bool, maglevTableSize int) *LBBPFMap {
	m := &LBBPFMap{}

	if maglev {
		m.maglevBackendIDsBuffer = make([]uint16, maglevTableSize)
		m.maglevTableSize = uint64(maglevTableSize)
	}

	return m
}

// UpsertService inserts or updates the given service in a BPF map.
func (lbmap *LBBPFMap) UpsertService(p *UpsertServiceParams) error {
	var svcKey ServiceKey

	if p.ID == 0 {
		return fmt.Errorf("Invalid svc ID 0")
	}

	if p.IPv6 {
		//svcKey = NewService6Key(p.IP, p.Port, u8proto.ANY, p.Scope, 0)
	} else {
		svcKey = NewService4Key(p.IP, p.Port, u8proto.ANY, p.Scope, 0)
	}

	slot := 1
	svcVal := svcKey.NewValue().(ServiceValue)
	if p.UseMaglev && len(p.Backends) != 0 {
		if err := lbmap.UpsertMaglevLookupTable(p.ID, p.Backends, p.IPv6); err != nil {
			return err
		}
	}

	backendIDs := make([]uint16, 0, len(p.Backends))
	for _, id := range p.Backends {
		backendIDs = append(backendIDs, id)
	}
	for _, backendID := range backendIDs {
		if backendID == 0 {
			return fmt.Errorf("Invalid backend ID 0")
		}
		svcVal.SetBackendID(loadbalancer.BackendID(backendID))
		svcVal.SetRevNat(int(p.ID))
		svcKey.SetBackendSlot(slot)
		if err := updateServiceEndpoint(svcKey, svcVal); err != nil {
			if errors.Is(err, unix.E2BIG) {
				return fmt.Errorf("Unable to update service entry %+v => %+v: "+
					"Unable to update element for LB bpf map: "+
					"You can resize it with the flag \"--%s\". "+
					"The resizing might break existing connections to services",
					svcKey, svcVal, option.LBMapEntriesName)
			}

			return fmt.Errorf("Unable to update service entry %+v => %+v: %w", svcKey, svcVal, err)
		}
		slot++
	}

	zeroValue := svcKey.NewValue().(ServiceValue)
	zeroValue.SetRevNat(int(p.ID)) // TODO change to uint16
	revNATKey := zeroValue.RevNatKey()
	revNATValue := svcKey.RevNatValue()
	if err := updateRevNatLocked(revNATKey, revNATValue); err != nil {
		return fmt.Errorf("Unable to update reverse NAT %+v => %+v: %s", revNATKey, revNATValue, err)
	}

	if err := updateMasterService(svcKey, len(backendIDs), int(p.ID), p.Type, p.Local,
		p.SessionAffinity, p.SessionAffinityTimeoutSec, p.CheckSourceRange); err != nil {
		deleteRevNatLocked(revNATKey)
		return fmt.Errorf("Unable to update service %+v: %s", svcKey, err)
	}

	for i := slot; i <= p.PrevBackendCount; i++ {
		svcKey.SetBackendSlot(i)
		if err := deleteServiceLocked(svcKey); err != nil {
			log.WithFields(logrus.Fields{
				logfields.ServiceKey:  svcKey,
				logfields.BackendSlot: svcKey.GetBackendSlot(),
			}).WithError(err).Warn("Unable to delete service entry from BPF map")
		}
	}

	return nil
}

func updateMasterService(fe ServiceKey, nbackends int, revNATID int, svcType loadbalancer.SVCType,
	svcLocal bool, sessionAffinity bool, sessionAffinityTimeoutSec uint32,
	checkSourceRange bool) error {

	// isRoutable denotes whether this service can be accessed from outside the cluster.
	isRoutable := !fe.IsSurrogate() &&
		(svcType != loadbalancer.SVCTypeClusterIP || option.Config.ExternalClusterIP)

	fe.SetBackendSlot(0)
	zeroValue := fe.NewValue().(ServiceValue)
	zeroValue.SetCount(nbackends)
	zeroValue.SetRevNat(revNATID)
	flag := loadbalancer.NewSvcFlag(&loadbalancer.SvcFlagParam{
		SvcType:          svcType,
		SvcLocal:         svcLocal,
		SessionAffinity:  sessionAffinity,
		IsRoutable:       isRoutable,
		CheckSourceRange: checkSourceRange,
	})
	zeroValue.SetFlags(flag.UInt16())
	if sessionAffinity {
		zeroValue.SetSessionAffinityTimeoutSec(sessionAffinityTimeoutSec)
	}

	return updateServiceEndpoint(fe, zeroValue)
}

func updateServiceEndpoint(key ServiceKey, value ServiceValue) error {
	log.WithFields(logrus.Fields{
		logfields.ServiceKey:   key,
		logfields.ServiceValue: value,
		logfields.BackendSlot:  key.GetBackendSlot(),
	}).Debug("Upserting service entry")

	if key.GetBackendSlot() != 0 && value.RevNatKey().GetKey() == 0 {
		return fmt.Errorf("invalid RevNat ID (0) in the Service Value")
	}
	if _, err := key.Map().OpenOrCreate(); err != nil {
		return err
	}

	return key.Map().Update(key.ToNetwork(), value.ToNetwork())
}
