package ctmap

import (
	"github.com/cilium/cilium/pkg/metrics"
	"math"
	"unsafe"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/bpf"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/lock"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/logging"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/logging/logfields"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/maps/nat"
)

const (
	// mapCount counts the maximum number of CT maps that one endpoint may
	// access at once.
	mapCount = 4

	// Map names for TCP CT tables are retained from Cilium 1.0 naming
	// scheme to minimize disruption of ongoing connections during upgrade.
	MapNamePrefix     = "cilium_ct"
	MapNameTCP6       = MapNamePrefix + "6_"
	MapNameTCP4       = MapNamePrefix + "4_"
	MapNameTCP6Global = MapNameTCP6 + "global"
	MapNameTCP4Global = MapNameTCP4 + "global"

	// Map names for "any" protocols indicate CT for non-TCP protocols.
	MapNameAny6       = MapNamePrefix + "_any6_"
	MapNameAny4       = MapNamePrefix + "_any4_"
	MapNameAny6Global = MapNameAny6 + "global"
	MapNameAny4Global = MapNameAny4 + "global"

	mapNumEntriesLocal = 64000

	TUPLE_F_OUT     = 0
	TUPLE_F_IN      = 1
	TUPLE_F_RELATED = 2
	TUPLE_F_SERVICE = 4

	// MaxTime specifies the last possible time for GCFilter.Time
	MaxTime = math.MaxUint32

	metricsAlive   = "alive"
	metricsDeleted = "deleted"

	metricsIngress = "ingress"
	metricsEgress  = "egress"
)

var (
	log = logging.DefaultLogger.WithField(logfields.LogSubsys, "map-ct")

	// labelIPv6CTDumpInterrupts marks the count for conntrack dump resets (IPv6).
	labelIPv6CTDumpInterrupts = map[string]string{
		metrics.LabelDatapathArea:   "conntrack",
		metrics.LabelDatapathName:   "dump_interrupts",
		metrics.LabelDatapathFamily: "ipv6",
	}
	// labelIPv4CTDumpInterrupts marks the count for conntrack dump resets (IPv4).
	labelIPv4CTDumpInterrupts = map[string]string{
		metrics.LabelDatapathArea:   "conntrack",
		metrics.LabelDatapathName:   "dump_interrupts",
		metrics.LabelDatapathFamily: "ipv4",
	}

	mapInfo map[mapType]mapAttributes
)

var globalDeleteLock [mapTypeMax]lock.Mutex
var natMapsLock [mapTypeMax]*lock.Mutex

type mapAttributes struct {
	mapKey     bpf.MapKey
	keySize    int
	mapValue   bpf.MapValue
	valueSize  int
	maxEntries int
	parser     bpf.DumpParser
	bpfDefine  string
	natMapLock *lock.Mutex // Serializes concurrent accesses to natMap
	natMap     *nat.Map
}

func setupMapInfo(m mapType, define string, mapKey bpf.MapKey, keySize int, maxEntries int, nat *nat.Map) {
	mapInfo[m] = mapAttributes{
		bpfDefine: define,
		mapKey:    mapKey,
		keySize:   keySize,
		// the value type is CtEntry for all CT maps
		mapValue:   &CtEntry{},
		valueSize:  SizeofCtEntry,
		maxEntries: maxEntries,
		parser:     bpf.ConvertKeyValue,
		natMapLock: natMapsLock[m],
		natMap:     nat,
	}
}

// InitMapInfo builds the information about different CT maps for the
// combination of L3/L4 protocols, using the specified limits on TCP vs non-TCP
// maps.
func InitMapInfo(tcpMaxEntries, anyMaxEntries int, v4, v6, nodeport bool) {
	mapInfo = make(map[mapType]mapAttributes, mapTypeMax)

	global4Map, global6Map := nat.GlobalMaps(v4, v6, nodeport)

	// SNAT also only works if the CT map is global so all local maps will be nil
	natMaps := map[mapType]*nat.Map{
		mapTypeIPv4TCPLocal:  nil,
		mapTypeIPv6TCPLocal:  nil,
		mapTypeIPv4TCPGlobal: global4Map,
		mapTypeIPv6TCPGlobal: global6Map,
		mapTypeIPv4AnyLocal:  nil,
		mapTypeIPv6AnyLocal:  nil,
		mapTypeIPv4AnyGlobal: global4Map,
		mapTypeIPv6AnyGlobal: global6Map,
	}
	global4MapLock := &lock.Mutex{}
	global6MapLock := &lock.Mutex{}
	natMapsLock[mapTypeIPv4TCPGlobal] = global4MapLock
	natMapsLock[mapTypeIPv6TCPGlobal] = global6MapLock
	natMapsLock[mapTypeIPv4AnyGlobal] = global4MapLock
	natMapsLock[mapTypeIPv6AnyGlobal] = global6MapLock

	setupMapInfo(mapTypeIPv4TCPLocal, "CT_MAP_TCP4",
		&CtKey4{}, int(unsafe.Sizeof(CtKey4{})),
		mapNumEntriesLocal, natMaps[mapTypeIPv4TCPLocal])

	//setupMapInfo(mapTypeIPv6TCPLocal, "CT_MAP_TCP6",
	//	&CtKey6{}, int(unsafe.Sizeof(CtKey6{})),
	//	mapNumEntriesLocal, natMaps[mapTypeIPv6TCPLocal])

	setupMapInfo(mapTypeIPv4TCPGlobal, "CT_MAP_TCP4",
		&CtKey4Global{}, int(unsafe.Sizeof(CtKey4Global{})),
		tcpMaxEntries, natMaps[mapTypeIPv4TCPGlobal])

	//setupMapInfo(mapTypeIPv6TCPGlobal, "CT_MAP_TCP6",
	//	&CtKey6Global{}, int(unsafe.Sizeof(CtKey6Global{})),
	//	tcpMaxEntries, natMaps[mapTypeIPv6TCPGlobal])

	setupMapInfo(mapTypeIPv4AnyLocal, "CT_MAP_ANY4",
		&CtKey4{}, int(unsafe.Sizeof(CtKey4{})),
		mapNumEntriesLocal, natMaps[mapTypeIPv4AnyLocal])

	//setupMapInfo(mapTypeIPv6AnyLocal, "CT_MAP_ANY6",
	//	&CtKey6{}, int(unsafe.Sizeof(CtKey6{})),
	//	mapNumEntriesLocal, natMaps[mapTypeIPv6AnyLocal])

	setupMapInfo(mapTypeIPv4AnyGlobal, "CT_MAP_ANY4",
		&CtKey4Global{}, int(unsafe.Sizeof(CtKey4Global{})),
		anyMaxEntries, natMaps[mapTypeIPv4AnyGlobal])

	//setupMapInfo(mapTypeIPv6AnyGlobal, "CT_MAP_ANY6",
	//	&CtKey6Global{}, int(unsafe.Sizeof(CtKey6Global{})),
	//	anyMaxEntries, natMaps[mapTypeIPv6AnyGlobal])
}
