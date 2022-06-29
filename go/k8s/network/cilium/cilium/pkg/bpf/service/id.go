package service

import (
	"fmt"
	"github.com/cilium/cilium/pkg/logging/logfields"
	log "github.com/sirupsen/logrus"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/k8s/loadbalancer"
	"sync"
)

const (
	// FirstFreeServiceID is the first ID for which the services should be assigned.
	FirstFreeServiceID = uint32(1)

	// MaxSetOfServiceID is maximum number of set of service IDs that can be stored
	// in the kvstore or the local ID allocator.
	MaxSetOfServiceID = uint32(0xFFFF)

	// FirstFreeBackendID is the first ID for which the backend should be assigned.
	// BPF datapath assumes that backend_id cannot be 0.
	FirstFreeBackendID = uint32(1)

	// MaxSetOfBackendID is maximum number of set of backendIDs IDs that can be
	// stored in the local ID allocator.
	MaxSetOfBackendID = uint32(0xFFFF)
)

var (
	serviceIDAlloc = NewIDAllocator(FirstFreeServiceID, MaxSetOfServiceID)
	backendIDAlloc = NewIDAllocator(FirstFreeBackendID, MaxSetOfBackendID)
)

// IDAllocator contains an internal state of the ID allocator.
type IDAllocator struct {
	// Protects entitiesID, entities, nextID and maxID
	sync.RWMutex

	// entitiesID is a map of all entities indexed by service or backend ID
	entitiesID map[uint32]*loadbalancer.L3n4AddrID

	// entities is a map of all entities indexed by L3n4Addr.StringID()
	entities map[string]uint32

	// nextID is the next ID to attempt to allocate
	nextID uint32

	// maxID is the maximum ID available for allocation
	maxID uint32

	// initNextID is the initial nextID
	initNextID uint32

	// initMaxID is the initial maxID
	initMaxID uint32
}

// NewIDAllocator creates a new ID allocator instance.
func NewIDAllocator(nextID uint32, maxID uint32) *IDAllocator {
	return &IDAllocator{
		entitiesID: map[uint32]*loadbalancer.L3n4AddrID{},
		entities:   map[string]uint32{},
		nextID:     nextID,
		maxID:      maxID,
		initNextID: nextID,
		initMaxID:  maxID,
	}
}

func (alloc *IDAllocator) acquireLocalID(svc loadbalancer.L3n4Addr, desiredID uint32) (*loadbalancer.L3n4AddrID, error) {
	alloc.Lock()
	defer alloc.Unlock()

	if svcID, ok := alloc.entities[svc.StringID()]; ok {
		if svc, ok := alloc.entitiesID[svcID]; ok {
			return svc, nil
		}
	}

	if desiredID != 0 {
		foundSVC, ok := alloc.entitiesID[desiredID]
		if !ok {
			return alloc.addID(svc, desiredID), nil
		}
		return nil, fmt.Errorf("service ID %d is already registered to %q",
			desiredID, foundSVC)
	}

	startingID := alloc.nextID
	rollover := false
	for {
		if alloc.nextID == startingID && rollover {
			break
		} else if alloc.nextID == alloc.maxID {
			alloc.nextID = FirstFreeServiceID
			rollover = true
		}

		if _, ok := alloc.entitiesID[alloc.nextID]; !ok {
			svcID := alloc.addID(svc, alloc.nextID)
			alloc.nextID++
			return svcID, nil
		}

		alloc.nextID++
	}

	return nil, fmt.Errorf("no service ID available")
}

func (alloc *IDAllocator) deleteLocalID(id uint32) error {
	alloc.Lock()
	defer alloc.Unlock()

	if svc, ok := alloc.entitiesID[id]; ok {
		delete(alloc.entitiesID, id)
		delete(alloc.entities, svc.StringID())
	}

	return nil
}

// RestoreBackendID tries to restore the given local ID for the given backend.
//
// If ID cannot be restored (ID already taken), returns an error.
func RestoreBackendID(l3n4Addr loadbalancer.L3n4Addr, id loadbalancer.BackendID) error {
	newID, err := restoreBackendID(l3n4Addr, id)
	if err != nil {
		return err
	}

	// This shouldn't happen (otherwise, there is a bug in the code). But maybe it makes sense to delete all svc v2 in this case.
	if newID != id {
		DeleteBackendID(newID)
		return fmt.Errorf("restored backend ID for %+v does not match (%d != %d)",
			l3n4Addr, newID, id)
	}

	return nil
}

func restoreBackendID(l3n4Addr loadbalancer.L3n4Addr, id loadbalancer.BackendID) (loadbalancer.BackendID, error) {
	l3n4AddrID, err := backendIDAlloc.acquireLocalID(l3n4Addr, uint32(id))
	if err != nil {
		return 0, err
	}
	return loadbalancer.BackendID(l3n4AddrID.ID), nil
}

// DeleteBackendID releases the given backend ID.
// TODO maybe provide l3n4Addr as an arg for the extra safety.
func DeleteBackendID(id loadbalancer.BackendID) {
	backendIDAlloc.deleteLocalID(uint32(id))
}

// RestoreID restores  previously used service ID
func RestoreID(l3n4Addr loadbalancer.L3n4Addr, id uint32) (*loadbalancer.L3n4AddrID, error) {
	log.WithField(logfields.L3n4Addr, logfields.Repr(l3n4Addr)).Debug("Restoring service")

	return serviceIDAlloc.acquireLocalID(l3n4Addr, id)
}
