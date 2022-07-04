package lxcmap

import (
	"fmt"
	"net"
	"unsafe"

	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf"
)

const (
	MapName = "cilium_lxc"

	// MaxEntries represents the maximum number of endpoints in the map
	MaxEntries = 65535

	// PortMapMax represents the maximum number of Ports Mapping per container.
	PortMapMax = 16

	// EndpointFlagHost indicates that this endpoint represents the host
	EndpointFlagHost = 1
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
	).WithCache()
)

// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapKey
type EndpointKey struct {
	bpf.EndpointKey
}

// NewEndpointKey returns an EndpointKey based on the provided IP address. The
// address family is automatically detected
func NewEndpointKey(ip net.IP) *EndpointKey {
	return &EndpointKey{
		EndpointKey: bpf.NewEndpointKey(ip),
	}
}

// MAC is the __u64 representation of a MAC address.
type MAC uint64

func (m MAC) String() string {
	return fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X",
		uint64(m&0x0000000000FF),
		uint64((m&0x00000000FF00)>>8),
		uint64((m&0x000000FF0000)>>16),
		uint64((m&0x0000FF000000)>>24),
		uint64((m&0x00FF00000000)>>32),
		uint64((m&0xFF0000000000)>>40),
	)
}

type pad4uint32 [4]uint32

// DeepCopyInto is a deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *pad4uint32) DeepCopyInto(out *pad4uint32) {
	copy(out[:], in[:])
	return
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

// IsHost returns true if the EndpointInfo represents a host IP
func (v *EndpointInfo) IsHost() bool {
	return v.Flags&EndpointFlagHost != 0
}

// DumpToMap dumps the contents of the lxcmap into a map and returns it
func DumpToMap() (map[string]*EndpointInfo, error) {
	m := map[string]*EndpointInfo{}
	callback := func(key bpf.MapKey, value bpf.MapValue) {
		if info, ok := value.DeepCopyMapValue().(*EndpointInfo); ok {
			if endpointKey, ok := key.(*EndpointKey); ok {
				m[endpointKey.ToIP().String()] = info
			}
		}
	}

	if err := LXCMap.DumpWithCallback(callback); err != nil {
		return nil, fmt.Errorf("unable to read BPF endpoint list: %s", err)
	}

	return m, nil
}

// DeleteEntry deletes a single map entry
func DeleteEntry(ip net.IP) error {
	return LXCMap.Delete(NewEndpointKey(ip))
}
