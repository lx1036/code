package lbmap

import (
	"unsafe"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/bpf"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/types"
)

const (
	SourceRange4MapName = "cilium_lb4_source_range"
	SourceRange6MapName = "cilium_lb6_source_range"
	lpmPrefixLen4       = 16 + 16 // sizeof(SourceRangeKey4.RevNATID)+sizeof(SourceRangeKey4.Pad)
	lpmPrefixLen6       = 16 + 16 // sizeof(SourceRangeKey6.RevNATID)+sizeof(SourceRangeKey6.Pad)
)

var (
	// SourceRange4Map is the BPF map for storing IPv4 service source ranges to
	// check if option.Config.EnableSVCSourceRangeCheck is enabled.
	SourceRange4Map *bpf.Map
)

// initSourceRange creates the BPF maps for storing both IPv4 and IPv6
// service source ranges.
func initSourceRange(params InitParams) {
	if params.IPv4 {
		SourceRange4Map = bpf.NewMap(
			SourceRange4MapName,
			bpf.MapTypeLPMTrie,
			&SourceRangeKey4{}, int(unsafe.Sizeof(SourceRangeKey4{})),
			&SourceRangeValue{}, int(unsafe.Sizeof(SourceRangeValue{})),
			MaxEntries,
			bpf.BPF_F_NO_PREALLOC, 0,
			bpf.ConvertKeyValue,
		).WithCache().WithPressureMetric()
	}
}

// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapKey
type SourceRangeKey4 struct {
	PrefixLen uint32     `align:"lpm_key"`
	RevNATID  uint16     `align:"rev_nat_id"`
	Pad       uint16     `align:"pad"`
	Address   types.IPv4 `align:"addr"`
}

func (s *SourceRangeKey4) String() string {
	//TODO implement me
	panic("implement me")
}

func (s *SourceRangeKey4) GetKeyPtr() unsafe.Pointer {
	//TODO implement me
	panic("implement me")
}

func (s *SourceRangeKey4) NewValue() bpf.MapValue {
	//TODO implement me
	panic("implement me")
}

func (s *SourceRangeKey4) DeepCopyMapKey() bpf.MapKey {
	//TODO implement me
	panic("implement me")
}

// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapValue
type SourceRangeValue struct {
	Pad uint8 // not used
}

func (v *SourceRangeValue) String() string {
	//TODO implement me
	panic("implement me")
}

func (v *SourceRangeValue) GetValuePtr() unsafe.Pointer {
	//TODO implement me
	panic("implement me")
}

func (v *SourceRangeValue) DeepCopyMapValue() bpf.MapValue {
	//TODO implement me
	panic("implement me")
}
