package lbmap

import (
	"fmt"
	"github.com/cilium/cilium/pkg/logging/logfields"
	log "github.com/sirupsen/logrus"
	"net"
	"strconv"

	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/k8s/loadbalancer"

	"github.com/cilium/cilium/pkg/u8proto"
)

const (
	// Maximum number of entries in each hashtable
	MaxEntries = 65536
)

// LBBPFMap is an implementation of the LBMap interface.
type LBBPFMap struct{}

func (*LBBPFMap) UpdateOrInsertService(
	svcID uint16, svcIP net.IP, svcPort uint16,
	backendIDs []uint16, prevBackendCount int,
	ipv6 bool, svcType loadbalancer.SVCType, svcLocal bool,
	svcScope uint8, sessionAffinity bool,
	sessionAffinityTimeoutSec uint32) error {

	var svcKey Service4Key

	if svcID == 0 {
		return fmt.Errorf("Invalid svc ID 0")
	}

	svcKey = NewService4Key(svcIP, svcPort, u8proto.ANY, svcScope, 0)

	slot := 1
	svcVal := svcKey.NewValue().(*Service4Value)
	for _, backendID := range backendIDs {
		if backendID == 0 {
			return fmt.Errorf("Invalid backend ID 0")
		}
		svcVal.SetBackendID(loadbalancer.BackendID(backendID))
		svcVal.SetRevNat(int(svcID))
		svcKey.SetSlave(slot) // TODO(brb) Rename to SetSlot
		if err := updateServiceEndpoint(svcKey, svcVal); err != nil {
			return fmt.Errorf("Unable to update service entry %+v => %+v: %s",
				svcKey, svcVal, err)
		}
		slot++
	}

	zeroValue := svcKey.NewValue().(ServiceValue)
	zeroValue.SetRevNat(int(svcID)) // TODO change to uint16
	revNATKey := zeroValue.RevNatKey()
	revNATValue := svcKey.RevNatValue()
	if err := updateRevNatLocked(revNATKey, revNATValue); err != nil {
		return fmt.Errorf("Unable to update reverse NAT %+v => %+v: %s", revNATKey, revNATValue, err)
	}

	if err := updateMasterService(svcKey, len(backendIDs), int(svcID), svcType, svcLocal,
		sessionAffinity, sessionAffinityTimeoutSec); err != nil {
		deleteRevNatLocked(revNATKey)
		return fmt.Errorf("Unable to update service %+v: %s", svcKey, err)
	}

	for i := slot; i <= prevBackendCount; i++ {
		svcKey.SetSlave(i)
		if err := deleteServiceLocked(svcKey); err != nil {
			log.WithFields(log.Fields{
				logfields.ServiceKey: svcKey,
				logfields.SlaveSlot:  svcKey.GetSlave(),
			}).WithError(err).Warn("Unable to delete service entry from BPF map")
		}
	}

	return nil
}

// DeleteService removes given service from a BPF map.
func (*LBBPFMap) DeleteService(svc loadbalancer.L3n4AddrID, backendCount int) error {
	if svc.ID == 0 {
		return fmt.Errorf("Invalid svc ID 0")
	}

	svcKey := NewService4Key(svc.IP, svc.Port, u8proto.ANY, svc.Scope, 0)
	revNATKey := NewRevNat4Key(uint16(svc.ID))

	for slot := 0; slot <= backendCount; slot++ {
		svcKey.SetSlave(slot)
		if err := svcKey.MapDelete(); err != nil {
			return fmt.Errorf("Unable to delete service entry %+v: %s", svcKey, err)
		}
	}

	if err := deleteRevNatLocked(revNATKey); err != nil {
		return fmt.Errorf("Unable to delete revNAT entry %+v: %s", revNATKey, err)
	}

	return nil
}

// DumpBackendMaps dumps the backend entries from the BPF maps.
func (*LBBPFMap) DumpBackendMaps() ([]*loadbalancer.Backend, error) {
	backendValueMap := map[loadbalancer.BackendID]Backend4Value{}
	lbBackends := []*loadbalancer.Backend{}

	parseBackendEntries := func(key bpf.MapKey, value bpf.MapValue) {
		// No need to deep copy the key because we are using the ID which
		// is a value.
		backendKey := key.(Backend4Key)
		backendValue := value.DeepCopyMapValue().(Backend4Value)
		backendValueMap[backendKey.GetID()] = backendValue
	}

	err := Backend4Map.DumpWithCallback(parseBackendEntries)
	if err != nil {
		return nil, fmt.Errorf("Unable to dump lb4 backends map: %s", err)
	}

	for backendID, backendVal := range backendValueMap {
		ip := backendVal.GetAddress()
		port := backendVal.GetPort()
		proto := loadbalancer.NONE
		lbBackend := loadbalancer.NewBackend(backendID, proto, ip, port)
		lbBackends = append(lbBackends, lbBackend)
	}

	return lbBackends, nil
}

// DumpServiceMaps dumps the services from the BPF maps.
func (*LBBPFMap) DumpServiceMaps() ([]*loadbalancer.SVC, []error) {
	var errors []error
	backendValueMap := map[loadbalancer.BackendID]Backend4Value{}
	newSVCMap := svcMap{}
	flagsCache := map[string]loadbalancer.ServiceFlags{}

	parseBackendEntries := func(key bpf.MapKey, value bpf.MapValue) {
		backendKey := key.(Backend4Key)
		backendValue := value.DeepCopyMapValue().(Backend4Value)
		backendValueMap[backendKey.GetID()] = backendValue
	}

	parseSVCEntries := func(key bpf.MapKey, value bpf.MapValue) {
		svcKey := key.DeepCopyMapKey().(Service4Key)
		svcValue := value.DeepCopyMapValue().(Service4Value)

		fe := svcFrontend(svcKey, svcValue)

		// Create master entry in case there are no backends.
		if svcKey.GetSlave() == 0 {
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
			errors = append(errors, fmt.Errorf("backend %d not found", backendID))
			return
		}

		be := svcBackend(backendID, backendValue)
		newSVCMap.addFEnBE(fe, be, svcKey.GetSlave())
	}

	// TODO optimization: instead of dumping the backend map, we can
	// pass its content to the function.
	err := Backend4Map.DumpWithCallback(parseBackendEntries)
	if err != nil {
		errors = append(errors, err)
	}
	err = Service4MapV2.DumpWithCallback(parseSVCEntries)
	if err != nil {
		errors = append(errors, err)
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

	return newSVCList, errors
}

// DumpAffinityMatches returns the affinity match map represented as a nested
// map which first key is svc ID and the second - backend ID.
func (*LBBPFMap) DumpAffinityMatches() (BackendIDByServiceIDSet, error) {
	matches := BackendIDByServiceIDSet{}

	parse := func(key bpf.MapKey, value bpf.MapValue) {
		matchKey := key.DeepCopyMapKey().(*AffinityMatchKey)
		svcID := matchKey.RevNATID
		backendID := uint16(matchKey.BackendID) // currently backend_id is u16

		if _, ok := matches[svcID]; !ok {
			matches[svcID] = map[uint16]struct{}{}
		}
		matches[svcID][backendID] = struct{}{}
	}

	err := AffinityMatchMap.DumpWithCallback(parse)
	if err != nil {
		return nil, err
	}

	return matches, nil
}

// DeleteAffinityMatch removes the affinity match for the given svc and backend ID
// tuple from the BPF map
func (*LBBPFMap) DeleteAffinityMatch(revNATID uint16, backendID uint16) error {
	return AffinityMatchMap.Delete(NewAffinityMatchKey(revNATID, uint32(backendID)).ToNetwork())
}

// DeleteBackendByID removes a backend identified with the given ID from a BPF map.
func (*LBBPFMap) DeleteBackendByID(id uint16, ipv6 bool) error {
	var key *Backend4Key

	if id == 0 {
		return fmt.Errorf("invalid backend ID 0")
	}

	key = NewBackend4Key(loadbalancer.BackendID(id))

	if err := key.Map().Delete(key); err != nil {
		return fmt.Errorf("unable to delete backend %d (%t): %s", id, ipv6, err)
	}

	return nil
}

type svcMap map[string]loadbalancer.SVC

func svcFrontend(svcKey Service4Key, svcValue Service4Value) *loadbalancer.L3n4AddrID {
	feL3n4Addr := loadbalancer.NewL3n4Addr(loadbalancer.NONE, svcKey.GetAddress(), svcKey.GetPort(), svcKey.GetScope())
	feL3n4AddrID := &loadbalancer.L3n4AddrID{
		L3n4Addr: *feL3n4Addr,
		ID:       loadbalancer.ID(svcValue.GetRevNat()),
	}
	return feL3n4AddrID
}

func svcBackend(backendID loadbalancer.BackendID, backend Backend4Value) *loadbalancer.Backend {
	beIP := backend.GetAddress()
	bePort := backend.GetPort()
	beProto := loadbalancer.NONE
	beBackend := loadbalancer.NewBackend(backendID, beProto, beIP, bePort)
	return beBackend
}
