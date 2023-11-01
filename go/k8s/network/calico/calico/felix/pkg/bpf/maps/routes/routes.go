package routes

import (
	"k8s-lx1036/k8s/network/calico/calico/felix/pkg/bpf"
	"k8s-lx1036/k8s/network/calico/calico/felix/pkg/bpf/maps"

	"golang.org/x/sys/unix"
)

// struct cali_rt_key {
// __u32 mask;
// __be32 addr; // NBO
// };
const KeySize = 8

//	struct cali_rt_value {
//	  __u32 flags;
//	  union {
//	    __u32 next_hop;
//	    __u32 ifIndex;
//	  };
//	};
const ValueSize = 8

var MapParameters = maps.MapParameters{
	Filename:   "/sys/fs/bpf/tc/globals/cali_v4_routes",
	Type:       "lpm_trie",
	KeySize:    KeySize,
	ValueSize:  ValueSize,
	MaxEntries: 256 * 1024,
	Name:       "cali_v4_routes",
	Flags:      unix.BPF_F_NO_PREALLOC,
}

func Map(mc *bpf.MapContext) *bpf.Map {
	return mc.NewPinnedMap(MapParameters)
}
