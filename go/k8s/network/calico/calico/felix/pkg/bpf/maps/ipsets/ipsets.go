package ipsets

import (
	"golang.org/x/sys/unix"
	"k8s-lx1036/k8s/network/calico/calico/felix/pkg/bpf"
	"k8s-lx1036/k8s/network/calico/calico/felix/pkg/bpf/maps"
)

// IPSetEntrySize WARNING: must be kept in sync with the definitions in bpf/polprog/pol_prog_builder.go.
// WARNING: must be kept in sync with the definitions in bpf/include/policy.h.
// uint32 prefixLen HE  4
// uint64 set_id BE     +8 = 12
// uint32 addr BE       +4 = 16
// uint16 port HE       +2 = 18
// uint8 proto          +1 = 19
// uint8 pad            +1 = 20
const IPSetEntrySize = 20

var MapParameters = maps.MapParameters{
	Filename:   "/sys/fs/bpf/tc/globals/cali_v4_ip_sets",
	Type:       "lpm_trie",
	KeySize:    IPSetEntrySize,
	ValueSize:  4,
	MaxEntries: 1024 * 1024,
	Name:       "cali_v4_ip_sets",
	Flags:      unix.BPF_F_NO_PREALLOC,
}

func Map(mc *bpf.MapContext) *bpf.Map {
	return mc.NewPinnedMap(MapParameters)
}
