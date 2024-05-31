package nat

import (
	"github.com/cilium/cilium/pkg/u8proto"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/config/option"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/config/types"
	"unsafe"
)

const (
	// MapNameSnat4Global represents global IPv4 NAT table.
	MapNameSnat4Global = "cilium_snat_v4_external"
)

// NatMap represents a NAT map.
// It also implements the NatMap interface.
type NatMap struct {
	bpf.Map
	v4 bool
}

func NewMap(name string, v4 bool, entries int) *NatMap {
	return &NatMap{
		Map: *bpf.NewMap(
			name,
			bpf.MapTypeLRUHash,
			&NatKey4{},
			int(unsafe.Sizeof(NatKey4{})),
			&NatEntry4{},
			int(unsafe.Sizeof(NatEntry4{})),
			entries,
			0, 0,
			bpf.ConvertKeyValue,
		).WithCache(),
		v4: v4,
	}
}

// GlobalMaps returns all global NAT maps.
func GlobalMaps(ipv4, ipv6 bool) (ipv4Map, ipv6Map *NatMap) {
	entries := option.Config.NATMapEntriesGlobal
	if entries == 0 {
		entries = option.LimitTableMax
	}
	if ipv4 {
		ipv4Map = NewMap(MapNameSnat4Global, true, entries)
	}

	return
}

// NatKey4 is needed to provide NatEntry type to Lookup values.
// represents the key for IPv4 entries in the local BPF conntrack map.
// Address field names are correct for return traffic, i.e., they are reversed
// compared to the original direction traffic.
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapKey
type NatKey4 struct {
	DestAddr   types.IPv4      `align:"daddr"`
	SourceAddr types.IPv4      `align:"saddr"`
	DestPort   uint16          `align:"dport"`
	SourcePort uint16          `align:"sport"`
	NextHeader u8proto.U8proto `align:"nexthdr"`
	Flags      uint8           `align:"flags"`
}

// NatEntry4 represents an IPv4 entry in the NAT table.
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapValue
type NatEntry4 struct {
	Created   uint64     `align:"created"`
	HostLocal uint64     `align:"host_local"`
	Pad1      uint64     `align:"pad1"`
	Pad2      uint64     `align:"pad2"`
	Addr      types.IPv4 `align:"to_saddr"`
	Port      uint16     `align:"to_sport"`
}
