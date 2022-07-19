package fragmap

import (
	"unsafe"

	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/config/types"
)

const (
	// MapName is the name of the map used to retrieve L4 ports associated
	// to the datagram to which an IPv4 belongs.
	MapName = "cilium_ipv4_frag_datagrams"
)

// FragmentKey must match 'struct ipv4_frag_id' in "bpf/lib/ipv4.h".
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapKey
type FragmentKey struct {
	destAddr   types.IPv4 `align:"daddr"`
	sourceAddr types.IPv4 `align:"saddr"`
	id         uint16     `align:"id"`
	proto      uint8      `align:"proto"`
	pad        uint8      `align:"pad"`
}

// FragmentValue must match 'struct ipv4_frag_l4ports' in "bpf/lib/ipv4.h".
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapValue
type FragmentValue struct {
	sourcePort uint16 `align:"sport"`
	destPort   uint16 `align:"dport"`
}

// InitMap creates the signal map in the kernel.
func InitMap(mapEntries int) error {
	fragMap := bpf.NewMap(MapName,
		bpf.MapTypeLRUHash,
		&FragmentKey{},
		int(unsafe.Sizeof(FragmentKey{})),
		&FragmentValue{},
		int(unsafe.Sizeof(FragmentValue{})),
		mapEntries,
		0,
		0,
		bpf.ConvertKeyValue,
	)
	_, err := fragMap.Create()
	return err
}
