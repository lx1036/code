package arp

import (
	"k8s-lx1036/k8s/network/calico/calico/felix/pkg/bpf"
	"k8s-lx1036/k8s/network/calico/calico/felix/pkg/bpf/maps"
)

const KeySize = 8
const ValueSize = 12

var MapParams = maps.MapParameters{
	Filename:   "/sys/fs/bpf/tc/globals/cali_v4_arp",
	Type:       "lru_hash",
	KeySize:    KeySize,
	ValueSize:  ValueSize,
	MaxEntries: 10000, // max number of nodes that can forward nodeports to a single node
	Name:       "cali_v4_arp",
	Version:    2,
}

func Map(mc *bpf.MapContext) *bpf.Map {
	return mc.NewPinnedMap(MapParams)
}
