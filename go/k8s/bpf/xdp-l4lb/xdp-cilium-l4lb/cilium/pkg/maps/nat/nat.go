package nat

import (
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/bpf"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/option"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/u8proto"
	"strings"
)

const (
	// MapNameSnat4Global represents global IPv4 NAT table.
	MapNameSnat4Global = "cilium_snat_v4_external"
	// MapNameSnat6Global represents global IPv6 NAT table.
	MapNameSnat6Global = "cilium_snat_v6_external"

	// MinPortSnatDefault represents default min port from range.
	MinPortSnatDefault = 1024
	// MaxPortSnatDefault represents default max port from range.
	MaxPortSnatDefault = 65535
)

type NatKey interface {
	bpf.MapKey

	// ToNetwork converts fields to network byte order.
	ToNetwork() NatKey

	// ToHost converts fields to host byte order.
	ToHost() NatKey

	// Dump contents of key to sb. Returns true if successful.
	Dump(sb *strings.Builder, reverse bool) bool

	// GetFlags flags containing the direction of the TupleKey.
	GetFlags() uint8

	// GetNextHeader returns the proto of the NatKey
	GetNextHeader() u8proto.U8proto
}

func NewMap(name string, v4 bool, entries int) *Map {
	var sizeKey, sizeVal int
	var mapKey bpf.MapKey
	var mapValue bpf.MapValue

	if v4 {
		mapKey = &NatKey4{}
		sizeKey = SizeofNatKey4
		mapValue = &NatEntry4{}
		sizeVal = SizeofNatEntry4
	} else {
		//mapKey = &NatKey6{}
		//sizeKey = SizeofNatKey6
		//mapValue = &NatEntry6{}
		//sizeVal = SizeofNatEntry6
	}
	return &Map{
		Map: *bpf.NewMap(
			name,
			bpf.MapTypeLRUHash,
			mapKey,
			sizeKey,
			mapValue,
			sizeVal,
			entries,
			0, 0,
			bpf.ConvertKeyValue,
		).WithCache(),
		v4: v4,
	}
}

type Map struct {
	bpf.Map
	v4 bool
}

func GlobalMaps(ipv4, ipv6, nodeport bool) (ipv4Map, ipv6Map *Map) {
	if !nodeport {
		return
	}
	entries := option.Config.NATMapEntriesGlobal
	if entries == 0 {
		entries = option.LimitTableMax
	}
	if ipv4 {
		ipv4Map = NewMap(MapNameSnat4Global, true, entries)
	}
	if ipv6 {
		ipv6Map = NewMap(MapNameSnat6Global, false, entries)
	}
	return
}
