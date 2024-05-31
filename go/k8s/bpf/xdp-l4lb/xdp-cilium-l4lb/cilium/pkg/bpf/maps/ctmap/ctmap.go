package ctmap

import (
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf"
)

const (
	// mapCount counts the maximum number of CT maps that one endpoint may
	// access at once.
	mapCount = 4

	MapNamePrefix     = "cilium_ct"
	MapNameTCP4       = MapNamePrefix + "4_"
	MapNameTCP4Global = MapNameTCP4 + "global" // cilium_ct4_global
	MapNameAny4       = MapNamePrefix + "_any4_"
	MapNameAny4Global = MapNameAny4 + "global" // cilium_ct_any4_global

)

var (
	mapInfo map[mapType]mapAttributes
)

type mapAttributes struct {
	mapKey     bpf.MapKey
	keySize    int
	mapValue   bpf.MapValue
	valueSize  int
	maxEntries int
	parser     bpf.DumpParser
	bpfDefine  string
	natMap     NatMap
}

// CtEndpoint represents an endpoint for the functions required to manage
// conntrack maps for the endpoint.
type CtEndpoint interface {
	GetID() uint64
}

// CTMap represents an instance of a BPF connection tracking map.
type CTMap struct {
	bpf.Map

	mapType mapType
	// define maps to the macro used in the datapath portion for the map
	// name, for example 'CT_MAP4'.
	define string
}

func newMap(mapName string, m mapType) *CTMap {
	result := &CTMap{
		Map: *bpf.NewMap(mapName,
			bpf.MapTypeLRUHash,
			mapInfo[m].mapKey,
			mapInfo[m].keySize,
			mapInfo[m].mapValue,
			mapInfo[m].valueSize,
			mapInfo[m].maxEntries,
			0, 0,
			mapInfo[m].parser,
		),
		mapType: m,
		define:  mapInfo[m].bpfDefine,
	}
	return result
}

func LocalMaps(e CtEndpoint, ipv4, ipv6 bool) []*CTMap {
	return maps(e, ipv4, ipv6)
}

// GlobalMaps cilium_ct4_global/cilium_ct_any4_global
func GlobalMaps(ipv4, ipv6 bool) []*CTMap {
	return maps(nil, ipv4, ipv6)
}

func maps(e CtEndpoint, ipv4, ipv6 bool) []*CTMap {
	result := make([]*CTMap, 0, mapCount)
	if e == nil {
		if ipv4 {
			result = append(result, newMap(MapNameTCP4Global, mapTypeIPv4TCPGlobal))
			result = append(result, newMap(MapNameAny4Global, mapTypeIPv4AnyGlobal))
		}
	} else {
		if ipv4 {
			result = append(result, newMap(bpf.LocalMapName(MapNameTCP4, uint16(e.GetID())),
				mapTypeIPv4TCPLocal))
			result = append(result, newMap(bpf.LocalMapName(MapNameAny4, uint16(e.GetID())),
				mapTypeIPv4AnyLocal))
		}
	}
	return result
}
