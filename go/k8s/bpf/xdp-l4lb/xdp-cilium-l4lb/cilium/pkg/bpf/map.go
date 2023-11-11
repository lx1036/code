package bpf

import (
	"github.com/cilium/cilium/pkg/datapath/linux/probes"
	"time"
)

type MapType int

// This enumeration must be in sync with enum bpf_map_type in <linux/bpf.h>
const (
	MapTypeUnspec MapType = iota
	MapTypeHash
	MapTypeArray
	MapTypeProgArray
	MapTypePerfEventArray
	MapTypePerCPUHash
	MapTypePerCPUArray
	MapTypeStackTrace
	MapTypeCgroupArray
	MapTypeLRUHash
	MapTypeLRUPerCPUHash
	MapTypeLPMTrie
	MapTypeArrayOfMaps
	MapTypeHashOfMaps
	MapTypeDevMap
	MapTypeSockMap
	MapTypeCPUMap
	MapTypeXSKMap
	MapTypeSockHash
	// MapTypeMaximum is the maximum supported known map type.
	MapTypeMaximum

	// maxSyncErrors is the maximum consecutive errors syncing before the
	// controller bails out
	maxSyncErrors = 512

	// errorResolverSchedulerMinInterval is the minimum interval for the
	// error resolver to be scheduled. This minimum interval ensures not to
	// overschedule if a large number of updates fail in a row.
	errorResolverSchedulerMinInterval = 5 * time.Second

	// errorResolverSchedulerDelay is the delay to update the controller
	// after determination that a run is needed. The delay allows to
	// schedule the resolver after series of updates have failed.
	errorResolverSchedulerDelay = 200 * time.Millisecond
)

var (
	supportedMapTypes *probes.MapTypes
)

func GetMapType(t MapType) MapType {
	// If the supported map types have not been set, default to the system
	// prober. This path enables unit tests to mock out the supported map
	// types.
	if supportedMapTypes == nil {
		setMapTypesFromProber(probes.NewProbeManager())
	}
	switch t {
	case MapTypeLPMTrie:
		fallthrough
	case MapTypeLRUHash:
		if !supportedMapTypes.HaveLruHashMapType {
			return MapTypeHash
		}
	}
	return t
}

func setMapTypesFromProber(prober prober) {
	features := prober.Probe()
	supportedMapTypes = &features.MapTypes
}

type prober interface {
	// Probe returns the kernel feaures available on machine.
	Probe() probes.Features
}
