package lbmap

import (
	"unsafe"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/bpf"
)

const (
	AffinityMatchMapName = "cilium_lb_affinity_match"
	Affinity4MapName     = "cilium_lb4_affinity"
	Affinity6MapName     = "cilium_lb6_affinity"
)

var (
	// AffinityMatchMap is the BPF map to implement session affinity.
	AffinityMatchMap *bpf.Map
	Affinity4Map     *bpf.Map
)

func initAffinity(params InitParams) {
	AffinityMatchMap = bpf.NewMap(
		AffinityMatchMapName,
		bpf.MapTypeHash,
		&AffinityMatchKey{},
		int(unsafe.Sizeof(AffinityMatchKey{})),
		&AffinityMatchValue{},
		int(unsafe.Sizeof(AffinityMatchValue{})),
		MaxEntries,
		0, 0,
		bpf.ConvertKeyValue,
	).WithCache().WithPressureMetric()

	if params.IPv4 {
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
	}
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

func (k *Affinity4Key) String() string {
	//TODO implement me
	panic("implement me")
}

func (k *Affinity4Key) GetKeyPtr() unsafe.Pointer {
	//TODO implement me
	panic("implement me")
}

func (k *Affinity4Key) NewValue() bpf.MapValue {
	//TODO implement me
	panic("implement me")
}

func (k *Affinity4Key) DeepCopyMapKey() bpf.MapKey {
	//TODO implement me
	panic("implement me")
}

// AffinityValue is the Go representing of lb_affinity_value
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapValue
type AffinityValue struct {
	LastUsed  uint64 `align:"last_used"`
	BackendID uint32 `align:"backend_id"`
	Pad       uint32 `align:"pad"`
}

func (v *AffinityValue) String() string {
	//TODO implement me
	panic("implement me")
}

func (v *AffinityValue) GetValuePtr() unsafe.Pointer {
	//TODO implement me
	panic("implement me")
}

func (v *AffinityValue) DeepCopyMapValue() bpf.MapValue {
	//TODO implement me
	panic("implement me")
}

// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapKey
type AffinityMatchKey struct {
	BackendID uint32 `align:"backend_id"`
	RevNATID  uint16 `align:"rev_nat_id"`
	Pad       uint16 `align:"pad"`
}

func (k *AffinityMatchKey) String() string {
	//TODO implement me
	panic("implement me")
}

func (k *AffinityMatchKey) GetKeyPtr() unsafe.Pointer {
	//TODO implement me
	panic("implement me")
}

func (k *AffinityMatchKey) NewValue() bpf.MapValue {
	//TODO implement me
	panic("implement me")
}

func (k *AffinityMatchKey) DeepCopyMapKey() bpf.MapKey {
	//TODO implement me
	panic("implement me")
}

// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapValue
type AffinityMatchValue struct {
	Pad uint8 `align:"pad"`
}

func (v *AffinityMatchValue) String() string {
	//TODO implement me
	panic("implement me")
}

func (v *AffinityMatchValue) GetValuePtr() unsafe.Pointer {
	//TODO implement me
	panic("implement me")
}

func (v *AffinityMatchValue) DeepCopyMapValue() bpf.MapValue {
	//TODO implement me
	panic("implement me")
}
