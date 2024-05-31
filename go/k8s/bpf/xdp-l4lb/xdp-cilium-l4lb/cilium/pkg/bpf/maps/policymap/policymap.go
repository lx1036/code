package policymap

import (
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf"
	"unsafe"
)

// cilium_policy_{EndpointID} 这些 BPF maps 对象，挂载的 BPF 虚拟文件系统在 /sys/fs/bpf/tc/globals/cilium_policy_xxx

const (
	PolicyCallMapName = "cilium_call_policy"

	MapName = "cilium_policy_"

	MaxEntries = 65535

	// PolicyCallMaxEntries is the upper limit of entries in the program
	// array for the tail calls to jump into the endpoint specific policy
	// programs. This number *MUST* be identical to the maximum endpoint ID.
	PolicyCallMaxEntries = ^uint16(0)
)

type PolicyMap struct {
	*bpf.Map
}

func newMap(path string) *PolicyMap {
	mapType := bpf.MapTypeHash
	flags := bpf.GetPreAllocateMapFlags(mapType)
	return &PolicyMap{
		Map: bpf.NewMap(
			path,
			mapType,
			&PolicyKey{},
			int(unsafe.Sizeof(PolicyKey{})),
			&PolicyEntry{},
			int(unsafe.Sizeof(PolicyEntry{})),
			MaxEntries,
			flags, 0,
			bpf.ConvertKeyValue,
		),
	}
}

// Create creates a policy map at the specified path.
// 每一个 endpoint 一个 BPF policymap，所以不用全局变量
func Create(path string) (bool, error) {
	m := newMap(path)
	return m.Create()
}

// PolicyKey represents a key in the BPF policy map for an endpoint. It must
// match the layout of policy_key in bpf/lib/common.h.
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapKey
type PolicyKey struct {
	Identity         uint32 `align:"sec_label"`
	DestPort         uint16 `align:"dport"` // In network byte-order
	Nexthdr          uint8  `align:"protocol"`
	TrafficDirection uint8  `align:"egress"`
}

// PolicyEntry represents an entry in the BPF policy map for an endpoint. It must
// match the layout of policy_entry in bpf/lib/common.h.
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapValue
type PolicyEntry struct {
	ProxyPort uint16 `align:"proxy_port"` // In network byte-order
	Pad0      uint16 `align:"pad0"`
	Pad1      uint16 `align:"pad1"`
	Pad2      uint16 `align:"pad2"`
	Packets   uint64 `align:"packets"`
	Bytes     uint64 `align:"bytes"`
}

// CallKey is the index into the prog array map.
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapKey
type CallKey struct {
	index uint32
}

// CallValue is the program ID in the prog array map.
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapValue
type CallValue struct {
	progID uint32
}

// InitCallMap creates the policy call map in the kernel.
// /sys/fs/bpf/tc/globals/cilium_call_policy
func InitCallMap() error {
	policyCallMap := bpf.NewMap(PolicyCallMapName,
		bpf.MapTypeProgArray,
		&CallKey{},
		int(unsafe.Sizeof(CallKey{})),
		&CallValue{},
		int(unsafe.Sizeof(CallValue{})),
		int(PolicyCallMaxEntries),
		0,
		0,
		bpf.ConvertKeyValue,
	)
	_, err := policyCallMap.Create()
	return err
}
