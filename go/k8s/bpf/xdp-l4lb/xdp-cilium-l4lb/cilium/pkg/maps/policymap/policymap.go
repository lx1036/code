package policymap

import (
    "fmt"
    "strconv"
    "unsafe"

    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/bpf"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/byteorder"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/policy/trafficdirection"
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

func (key *PolicyKey) String() string {
    trafficDirectionString := (trafficdirection.TrafficDirection)(key.TrafficDirection).String()
    if key.DestPort != 0 {
        return fmt.Sprintf("%s: %d %d/%d", trafficDirectionString, key.Identity,
            byteorder.NetworkToHost16(key.DestPort), key.Nexthdr)
    }
    return fmt.Sprintf("%s: %d", trafficDirectionString, key.Identity)
}

func (key *PolicyKey) GetKeyPtr() unsafe.Pointer {
    return unsafe.Pointer(key)
}

func (key *PolicyKey) NewValue() bpf.MapValue {
    return &PolicyEntry{}
}

func (p PolicyKey) DeepCopyMapKey() bpf.MapKey {
    //TODO implement me
    panic("implement me")
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

func (pe *PolicyEntry) String() string {
    return fmt.Sprintf("%d %d %d", pe.ProxyPort, pe.Packets, pe.Bytes)
}

func (pe *PolicyEntry) GetValuePtr() unsafe.Pointer {
    return unsafe.Pointer(pe)
}

func (pe PolicyEntry) DeepCopyMapValue() bpf.MapValue {
    //TODO implement me
    panic("implement me")
}

// OpenOrCreate opens (or creates) a policy map at the specified path, which
// is used to govern which peer identities can communicate with the endpoint
// protected by this map.
func OpenOrCreate(path string) (*PolicyMap, bool, error) {
    m := newMap(path)
    isNewMap, err := m.OpenOrCreate()
    return m, isNewMap, err
}

// CallKey is the index into the prog array map.
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapKey
type CallKey struct {
    index uint32
}

func (k *CallKey) DeepCopyMapKey() bpf.MapKey {
    //TODO implement me
    panic("implement me")
}

// GetKeyPtr returns the unsafe pointer to the BPF key
func (k *CallKey) GetKeyPtr() unsafe.Pointer { return unsafe.Pointer(k) }

// NewValue returns a new empty instance of the structure representing the BPF
// map value.
func (k CallKey) NewValue() bpf.MapValue {
    return &CallValue{}
}

// String converts the key into a human readable string format.
func (k *CallKey) String() string {
    return strconv.FormatUint(uint64(k.index), 10)
}

// CallValue is the program ID in the prog array map.
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapValue
type CallValue struct {
    progID uint32
}

// String converts the value into a human readable string format.
func (v *CallValue) String() string {
    return strconv.FormatUint(uint64(v.progID), 10)
}

// GetValuePtr returns the unsafe pointer to the BPF value
func (v *CallValue) GetValuePtr() unsafe.Pointer {
    return unsafe.Pointer(v)
}

func (c CallValue) DeepCopyMapValue() bpf.MapValue {
    //TODO implement me
    panic("implement me")
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
