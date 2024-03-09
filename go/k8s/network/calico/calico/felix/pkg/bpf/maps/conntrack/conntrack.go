package conntrack

import (
	"k8s-lx1036/k8s/network/calico/calico/felix/pkg/bpf"
	"k8s-lx1036/k8s/network/calico/calico/felix/pkg/bpf/maps"

	"golang.org/x/sys/unix"
)

//	struct calico_ct_key {
//	  uint32_t protocol;
//	  __be32 addr_a, addr_b; // NBO
//	  uint16_t port_a, port_b; // HBO
//	};
const KeySize = 16
const ValueSize = 64
const MaxEntries = 512000

var MapParams = maps.MapParameters{
	Filename:     "/sys/fs/bpf/tc/globals/cali_v4_ct",
	Type:         "hash",
	KeySize:      KeySize,
	ValueSize:    ValueSize,
	MaxEntries:   MaxEntries,
	Name:         "cali_v4_ct",
	Flags:        unix.BPF_F_NO_PREALLOC,
	Version:      2,
	UpdatedByBPF: true,
}

func Map(mc *bpf.MapContext) *bpf.Map {
	return mc.NewPinnedMap(MapParams)
}
