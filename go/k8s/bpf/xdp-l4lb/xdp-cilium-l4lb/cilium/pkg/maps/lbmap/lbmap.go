package lbmap

import (
	"errors"
	"fmt"
	"net"
	"sort"
	"strconv"

	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/bpf"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/loadbalancer"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/logging"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/logging/logfields"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/maglev"
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

type svcMap map[string]loadbalancer.SVC

func (svcs svcMap) addFE(fe *loadbalancer.L3n4AddrID) *loadbalancer.SVC {
	hash := fe.Hash()
	lbsvc, ok := svcs[hash]
	if !ok {
		lbsvc = loadbalancer.SVC{Frontend: *fe}
		svcs[hash] = lbsvc
	}
	return &lbsvc
}

func (svcs svcMap) addFEnBE(fe *loadbalancer.L3n4AddrID, be *loadbalancer.Backend, beIndex int) *loadbalancer.SVC {
	hash := fe.Hash()
	lbsvc, ok := svcs[hash]
	if !ok {
		var bes []loadbalancer.Backend
		if beIndex == 0 {
			bes = make([]loadbalancer.Backend, 1)
			bes[0] = *be
		} else {
			bes = make([]loadbalancer.Backend, beIndex)
			bes[beIndex-1] = *be
		}
		lbsvc = loadbalancer.SVC{
			Frontend: *fe,
			Backends: bes,
		}
	} else {
		var bes []loadbalancer.Backend
		if len(lbsvc.Backends) < beIndex {
			bes = make([]loadbalancer.Backend, beIndex)
			copy(bes, lbsvc.Backends)
			lbsvc.Backends = bes
		}
		if beIndex == 0 {
			lbsvc.Backends = append(lbsvc.Backends, *be)
		} else {
			lbsvc.Backends[beIndex-1] = *be
		}
	}

	svcs[hash] = lbsvc
	return &lbsvc
}

// DumpServiceMaps dumps the services from the BPF maps.
func (*LBBPFMap) DumpServiceMaps() ([]*loadbalancer.SVC, []error) {
	newSVCMap := svcMap{}
	var errs []error
	flagsCache := map[string]loadbalancer.ServiceFlags{}
	backendValueMap := map[loadbalancer.BackendID]BackendValue{}

	parseBackendEntries := func(key bpf.MapKey, value bpf.MapValue) {
		backendKey := key.(BackendKey)
		backendValue := value.DeepCopyMapValue().(BackendValue).ToHost()
		backendValueMap[backendKey.GetID()] = backendValue
	}

	parseSVCEntries := func(key bpf.MapKey, value bpf.MapValue) {
		svcKey := key.DeepCopyMapKey().(ServiceKey).ToHost()
		svcValue := value.DeepCopyMapValue().(ServiceValue).ToHost()

		fe := svcFrontend(svcKey, svcValue)

		// Create master entry in case there are no backends.
		if svcKey.GetBackendSlot() == 0 {
			// Build a cache of flags stored in the value of the master key to
			// map it later.
			// FIXME proto is being ignored everywhere in the datapath.
			addrStr := svcKey.GetAddress().String()
			portStr := strconv.Itoa(int(svcKey.GetPort()))
			flagsCache[net.JoinHostPort(addrStr, portStr)] = loadbalancer.ServiceFlags(svcValue.GetFlags())

			newSVCMap.addFE(fe)
			return
		}

		backendID := svcValue.GetBackendID()
		backendValue, found := backendValueMap[backendID]
		if !found {
			errs = append(errs, fmt.Errorf("backend %d not found", backendID))
			return
		}

		be := svcBackend(backendID, backendValue)
		newSVCMap.addFEnBE(fe, be, svcKey.GetBackendSlot())
	}

	if option.Config.EnableIPv4 {
		// TODO(brb) optimization: instead of dumping the backend map, we can
		// pass its content to the function.
		err := Backend4Map.DumpWithCallback(parseBackendEntries)
		if err != nil {
			errs = append(errs, err)
		}
		err = Service4MapV2.DumpWithCallback(parseSVCEntries)
		if err != nil {
			errs = append(errs, err)
		}
	}

	newSVCList := make([]*loadbalancer.SVC, 0, len(newSVCMap))
	for hash := range newSVCMap {
		svc := newSVCMap[hash]
		addrStr := svc.Frontend.IP.String()
		portStr := strconv.Itoa(int(svc.Frontend.Port))
		host := net.JoinHostPort(addrStr, portStr)
		svc.Type = flagsCache[host].SVCType()
		svc.TrafficPolicy = flagsCache[host].SVCTrafficPolicy()
		newSVCList = append(newSVCList, &svc)
	}

	return newSVCList, errs
}

// IsMaglevLookupTableRecreated returns true if the maglev lookup BPF map
// was recreated due to the changed M param.
func (*LBBPFMap) IsMaglevLookupTableRecreated(ipv6 bool) bool {
	if ipv6 {
		return maglevRecreatedIPv6
	}
	return maglevRecreatedIPv4
}

// UpsertMaglevLookupTable calculates Maglev lookup table for given backends, and
// inserts into the Maglev BPF map.
func (lbmap *LBBPFMap) UpsertMaglevLookupTable(svcID uint16, backends map[string]uint16, ipv6 bool) error {
	backendNames := make([]string, 0, len(backends))
	for name := range backends {
		backendNames = append(backendNames, name)
	}

	// Maglev algorithm might produce different lookup table for the same
	// set of backends listed in a different order. To avoid that sort
	// backends by name, as the names are the same on all nodes (in opposite
	// to backend IDs which are node-local).
	sort.Strings(backendNames)
	table := maglev.GetLookupTable(backendNames, lbmap.maglevTableSize)
	for i, pos := range table {
		lbmap.maglevBackendIDsBuffer[i] = backends[backendNames[pos]]
	}

	if err := updateMaglevTable(ipv6, svcID, lbmap.maglevBackendIDsBuffer); err != nil {
		return err
	}

	return nil
}
