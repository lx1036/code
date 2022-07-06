package endpointpolicymap

import (
	log "github.com/sirupsen/logrus"
	"sync"
	"unsafe"

	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf/maps/policymap"
)

// INFO: 查看 BPF 代码里 lib/eps.h::lookup_ip4_endpoint_policy_map(__u32 ip)

const (
	MapName      = "cilium_ep_to_policy"
	innerMapName = "ep-policy-inner-map"

	// MaxEntries represents the maximum number of endpoints in the map
	MaxEntries = 65536
)

var (
	buildMap sync.Once

	// EndpointPolicyMap is the global singleton of the endpoint policy map.
	EndpointPolicyMap *bpf.Map
)

// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapKey
type EndpointKey struct{ bpf.EndpointKey }

// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapValue
type EPPolicyValue struct{ Fd uint32 }

// CreateEPPolicyMap will create both the innerMap (needed for map in map types) and
// then after BPFFS is mounted create the epPolicyMap. We only create the innerFd once
// to avoid having multiple inner maps.
func CreateEPPolicyMap() {
	if err := CreateWithName(MapName); err != nil {
		log.WithError(err).Warning("Unable to open or create endpoint policy map")
	}
}

// CreateWithName INFO: 这里有个特殊的点是，用的 bpf map 为 ep-policy-inner-map hash map
func CreateWithName(mapName string) error {
	buildMap.Do(func() {
		mapType := bpf.MapTypeHash
		fd, err := bpf.CreateMap(mapType,
			uint32(unsafe.Sizeof(policymap.PolicyKey{})),
			uint32(unsafe.Sizeof(policymap.PolicyEntry{})),
			uint32(policymap.MaxEntries),
			bpf.GetPreAllocateMapFlags(mapType),
			0, innerMapName)

		if err != nil {
			log.WithError(err).Fatal("unable to create EP to policy map")
			return
		}

		EndpointPolicyMap = bpf.NewMap(mapName,
			bpf.MapTypeHashOfMaps,
			&EndpointKey{},
			int(unsafe.Sizeof(EndpointKey{})),
			&EPPolicyValue{},
			int(unsafe.Sizeof(EPPolicyValue{})),
			MaxEntries,
			0,
			0,
			bpf.ConvertKeyValue,
		).WithCache()
		EndpointPolicyMap.InnerID = uint32(fd) // 这里特殊
	})

	_, err := EndpointPolicyMap.OpenOrCreate()
	return err
}
