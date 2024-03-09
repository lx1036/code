package ipcache

import (
	"sync"
	"unsafe"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/bpf"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/logging"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/logging/logfields"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/types"
)

var log = logging.DefaultLogger.WithField(logfields.LogSubsys, "map-ipcache")

const (
	// MaxEntries is the maximum number of keys that can be present in the
	// RemoteEndpointMap.
	MaxEntries = 512000

	// Name is the canonical name for the IPCache map on the filesystem.
	Name = "cilium_ipcache"

	// maxPrefixLengths is an approximation of how many different CIDR
	// prefix lengths may be supported by the BPF datapath without causing
	// BPF code generation to exceed the verifier instruction limit.
	// It applies to Linux versions that lack support for LPM, ie < v4.11.
	//
	// This is based upon the defines in bpf/ep_config.h, which in turn
	// are derived by building the bpf/ directory and running the script
	// test/bpf/verifier-test.sh, then adjusting the number of unique
	// prefix lengths until the script passes.
	maxPrefixLengths6 = 4
	maxPrefixLengths4 = 18
)

var (
	// IPCache is a mapping of all endpoint IPs in the cluster which this
	// Cilium agent is a part of to their corresponding security identities.
	// It is a singleton; there is only one such map per agent.
	IPCache = NewMap(Name)
)

type Map struct {
	bpf.Map

	// detectDeleteSupport is used to initialize 'supportsDelete' the first
	// time that a delete is issued from the datapath.
	detectDeleteSupport sync.Once

	// deleteSupport is set to 'true' initially, then is updated to set
	// whether the underlying kernel supports delete operations on the map
	// the first time that supportsDelete() is called.
	deleteSupport bool
}

func NewMap(name string) *Map {
	return &Map{
		Map: *bpf.NewMap(
			name,
			bpf.MapTypeLPMTrie,
			&Key{},
			int(unsafe.Sizeof(Key{})),
			&RemoteEndpointInfo{},
			int(unsafe.Sizeof(RemoteEndpointInfo{})),
			MaxEntries,
			bpf.BPF_F_NO_PREALLOC, 0,
			bpf.ConvertKeyValue,
		).WithCache().WithPressureMetric(),
		deleteSupport: true,
	}
}

// Reopen attempts to close and re-open the IPCache map at the standard path
// on the filesystem.
func Reopen() error {
	return IPCache.Map.Reopen()
}

// Key implements the bpf.MapKey interface.
//
// Must be in sync with struct ipcache_key in <bpf/lib/maps.h>
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapKey
type Key struct {
	Prefixlen uint32 `align:"lpm_key"`
	Pad1      uint16 `align:"pad1"`
	Pad2      uint8  `align:"pad2"`
	Family    uint8  `align:"family"`
	// represents both IPv6 and IPv4 (in the lowest four bytes)
	IP types.IPv6 `align:"$union0"`
}

func (k *Key) String() string {
	//TODO implement me
	panic("implement me")
}

func (k *Key) GetKeyPtr() unsafe.Pointer {
	//TODO implement me
	panic("implement me")
}

func (k *Key) NewValue() bpf.MapValue {
	//TODO implement me
	panic("implement me")
}

func (k *Key) DeepCopyMapKey() bpf.MapKey {
	//TODO implement me
	panic("implement me")
}

// RemoteEndpointInfo implements the bpf.MapValue interface. It contains the
// security identity of a remote endpoint.
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapValue
type RemoteEndpointInfo struct {
	SecurityIdentity uint32     `align:"sec_label"`
	TunnelEndpoint   types.IPv4 `align:"tunnel_endpoint"`
	Key              uint8      `align:"key"`
}

func (v *RemoteEndpointInfo) String() string {
	//TODO implement me
	panic("implement me")
}

func (v *RemoteEndpointInfo) GetValuePtr() unsafe.Pointer {
	//TODO implement me
	panic("implement me")
}

func (v *RemoteEndpointInfo) DeepCopyMapValue() bpf.MapValue {
	//TODO implement me
	panic("implement me")
}
