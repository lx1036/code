package lbmap

import (
	"unsafe"

	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf"

	"github.com/cilium/cilium/pkg/byteorder"
)

const (
	AffinityMatchMapName = "cilium_lb_affinity_match"
	Affinity4MapName     = "cilium_lb4_affinity"
)

// BackendIDByServiceIDSet is the type of a set for checking whether a backend
// belongs to a given service
type BackendIDByServiceIDSet map[uint16]map[uint16]struct{} // svc ID => backend ID

var (
	AffinityMatchMap = bpf.NewMap(
		AffinityMatchMapName,
		bpf.MapTypeHash,
		&AffinityMatchKey{},
		int(unsafe.Sizeof(AffinityMatchKey{})),
		&AffinityMatchValue{},
		int(unsafe.Sizeof(AffinityMatchValue{})),
		MaxEntries,
		0, 0,
		func(key []byte, value []byte, mapKey bpf.MapKey, mapValue bpf.MapValue) (bpf.MapKey, bpf.MapValue, error) {
			aKey, aVal := mapKey.(*AffinityMatchKey), mapValue.(*AffinityMatchValue)

			if _, _, err := bpf.ConvertKeyValue(key, value, aKey, aVal); err != nil {
				return nil, nil, err
			}

			return aKey.ToNetwork(), aVal, nil
		}).WithCache()
	Affinity4Map = bpf.NewMap(
		Affinity4MapName,
		bpf.MapTypeLRUHash,
		&Affinity4Key{},
		int(unsafe.Sizeof(Affinity4Key{})),
		&AffinityValue{},
		int(unsafe.Sizeof(AffinityValue{})),
		MaxEntries,
		0,
		0,
		bpf.ConvertKeyValue,
	)
)

// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapKey
type AffinityMatchKey struct {
	BackendID uint32 `align:"backend_id"`
	RevNATID  uint16 `align:"rev_nat_id"`
	Pad       uint16 `align:"pad"`
}

// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapValue
type AffinityMatchValue struct {
	Pad uint8 `align:"pad"`
}

// NewAffinityMatchKey creates the AffinityMatch key
func NewAffinityMatchKey(revNATID uint16, backendID uint32) *AffinityMatchKey {
	return &AffinityMatchKey{
		BackendID: backendID,
		RevNATID:  revNATID,
	}
}

// ToNetwork returns the key in the network byte order
func (k *AffinityMatchKey) ToNetwork() *AffinityMatchKey {
	n := *k
	// For some reasons rev_nat_index is stored in network byte order in
	// the SVC BPF maps
	n.RevNATID = byteorder.HostToNetwork(n.RevNATID).(uint16)
	return &n
}

// Affinity4Key is the Go representation of lb4_affinity_key
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapKey
type Affinity4Key struct {
	ClientID    uint64 `align:"client_id"`
	RevNATID    uint16 `align:"rev_nat_id"`
	NetNSCookie uint8  `align:"netns_cookie"`
	Pad1        uint8  `align:"pad1"`
	Pad2        uint32 `align:"pad2"`
}

// AffinityValue is the Go representing of lb_affinity_value
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapValue
type AffinityValue struct {
	LastUsed  uint64 `align:"last_used"`
	BackendID uint32 `align:"backend_id"`
	Pad       uint32 `align:"pad"`
}
