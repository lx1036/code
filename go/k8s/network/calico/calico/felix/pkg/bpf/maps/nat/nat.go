package nat

import (
	"k8s-lx1036/k8s/network/calico/calico/felix/pkg/bpf"
	"k8s-lx1036/k8s/network/calico/calico/felix/pkg/bpf/maps"

	"golang.org/x/sys/unix"
)

// struct calico_nat_v4_key {
//    uint32_t prefixLen;
//    uint32_t addr; // NBO
//    uint16_t port; // HBO
//    uint8_t protocol;
//    uint32_t saddr;
//    uint8_t pad;
// };
const frontendKeySize = 16

// struct calico_nat_v4_value {
//    uint32_t id;
//    uint32_t count;
//    uint32_t local;
//    uint32_t affinity_timeo;
//    uint32_t flags;
// };
const frontendValueSize = 20

var FrontendMapParameters = maps.MapParameters{
	Filename:   "/sys/fs/bpf/tc/globals/cali_v4_nat_fe",
	Type:       "lpm_trie",
	KeySize:    frontendKeySize,
	ValueSize:  frontendValueSize,
	MaxEntries: 64 * 1024,
	Name:       "cali_v4_nat_fe",
	Flags:      unix.BPF_F_NO_PREALLOC,
	Version:    3,
}

func FrontendMap(mc *bpf.MapContext) *bpf.Map {
	return mc.NewPinnedMap(FrontendMapParameters)
}

// struct calico_nat_secondary_v4_key {
//   uint32_t id;
//   uint32_t ordinal;
// };
const backendKeySize = 8

// struct calico_nat_dest {
//    uint32_t addr;
//    uint16_t port;
//    uint8_t pad[2];
// };
const backendValueSize = 8

var BackendMapParameters = maps.MapParameters{
	Filename:   "/sys/fs/bpf/tc/globals/cali_v4_nat_be",
	Type:       "hash",
	KeySize:    backendKeySize,
	ValueSize:  backendValueSize,
	MaxEntries: 256 * 1024,
	Name:       "cali_v4_nat_be",
	Flags:      unix.BPF_F_NO_PREALLOC,
}

func BackendMap(mc *bpf.MapContext) *bpf.Map {
	return mc.NewPinnedMap(BackendMapParameters)
}

// struct calico_nat {
//	uint32_t addr;
//	uint16_t port;
//	uint8_t  protocol;
//	uint8_t  pad;
// };
const frontendAffKeySize = 8

// struct calico_nat_v4_affinity_key {
//    struct calico_nat_v4 nat_key;
// 	  uint32_t client_ip;
// 	  uint32_t padding;
// };
const affinityKeySize = frontendAffKeySize + 8

// struct calico_nat_v4_affinity_val {
//    struct calico_nat_dest;
//    uint64_t ts;
// };

const affinityValueSize = backendValueSize + 8

// AffinityMapParameters describe the AffinityMap
var AffinityMapParameters = maps.MapParameters{
	Filename:   "/sys/fs/bpf/tc/globals/cali_v4_nat_aff",
	Type:       "lru_hash",
	KeySize:    affinityKeySize,
	ValueSize:  affinityValueSize,
	MaxEntries: 64 * 1024,
	Name:       "cali_v4_nat_aff",
}

func AffinityMap(mc *bpf.MapContext) *bpf.Map {
	return mc.NewPinnedMap(AffinityMapParameters)
}

// struct sendrecv4_key {
// 	uint64_t cookie;
// 	uint32_t ip;
// 	uint32_t port;
// };
//
// struct sendrecv4_val {
// 	uint32_t ip;
// 	uint32_t port;
// };

const sendRecvMsgKeySize = 16
const ctNATsMsgKeySize = 24
const sendRecvMsgValueSize = 8

// SendRecvMsgMapParameters define SendRecvMsgMap
var SendRecvMsgMapParameters = maps.MapParameters{
	Filename:   "/sys/fs/bpf/tc/globals/cali_v4_srmsg",
	Type:       "lru_hash",
	KeySize:    sendRecvMsgKeySize,
	ValueSize:  sendRecvMsgValueSize,
	MaxEntries: 510000,
	Name:       "cali_v4_srmsg",
}

// SendRecvMsgMap tracks reverse translations for sendmsg/recvmsg of
// unconnected UDP
func SendRecvMsgMap(mc *bpf.MapContext) *bpf.Map {
	return mc.NewPinnedMap(SendRecvMsgMapParameters)
}

var CTNATsMapParameters = maps.MapParameters{
	Filename:   "/sys/fs/bpf/tc/globals/cali_v4_ct_nats",
	Type:       "lru_hash",
	KeySize:    ctNATsMsgKeySize,
	ValueSize:  sendRecvMsgValueSize,
	MaxEntries: 10000,
	Name:       "cali_v4_ct_nats",
}

func AllNATsMsgMap(mc *bpf.MapContext) *bpf.Map {
	return mc.NewPinnedMap(CTNATsMapParameters)
}
