package lxcmap

import (
	"unsafe"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/bpf"
)

const (
	MapName = "cilium_lxc"

	// MaxEntries represents the maximum number of endpoints in the map
	MaxEntries = 65535

	// PortMapMax represents the maximum number of Ports Mapping per container.
	PortMapMax = 16
)

var (
	// LXCMap represents the BPF map for endpoints
	LXCMap = bpf.NewMap(MapName,
		bpf.MapTypeHash,
		&EndpointKey{},
		int(unsafe.Sizeof(EndpointKey{})),
		&EndpointInfo{},
		int(unsafe.Sizeof(EndpointInfo{})),
		MaxEntries,
		0, 0,
		bpf.ConvertKeyValue,
	).WithCache().WithPressureMetric()
)

// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapKey
type EndpointKey struct {
	bpf.EndpointKey
}

func (k *EndpointKey) String() string {
	//TODO implement me
	panic("implement me")
}

func (k *EndpointKey) GetKeyPtr() unsafe.Pointer {
	//TODO implement me
	panic("implement me")
}

func (k *EndpointKey) NewValue() bpf.MapValue {
	//TODO implement me
	panic("implement me")
}

func (k *EndpointKey) DeepCopyMapKey() bpf.MapKey {
	//TODO implement me
	panic("implement me")
}

// EndpointInfo represents the value of the endpoints BPF map.
//
// Must be in sync with struct endpoint_info in <bpf/lib/common.h>
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapValue
type EndpointInfo struct {
	IfIndex uint32 `align:"ifindex"`
	Unused  uint16 `align:"unused"`
	LxcID   uint16 `align:"lxc_id"`
	Flags   uint32 `align:"flags"`
	// go alignment
	_       uint32
	MAC     MAC        `align:"mac"`
	NodeMAC MAC        `align:"node_mac"`
	Pad     pad4uint32 `align:"pad"`
}

func (v *EndpointInfo) String() string {
	//TODO implement me
	panic("implement me")
}

func (v *EndpointInfo) GetValuePtr() unsafe.Pointer {
	//TODO implement me
	panic("implement me")
}

func (v *EndpointInfo) DeepCopyMapValue() bpf.MapValue {
	//TODO implement me
	panic("implement me")
}
