package lbmap

import (
	"fmt"
	"github.com/cilium/cilium/pkg/logging/logfields"
	log "github.com/sirupsen/logrus"
	"net"

	"github.com/cilium/cilium/pkg/loadbalancer"
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
