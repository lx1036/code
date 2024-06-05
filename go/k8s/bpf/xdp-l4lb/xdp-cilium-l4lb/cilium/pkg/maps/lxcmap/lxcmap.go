package lxcmap

import (
    "fmt"
    "net"
    "unsafe"

    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/addressing"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/bpf"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/mac"
)

const (
    MapName = "cilium_lxc"

    // MaxEntries represents the maximum number of endpoints in the map
    MaxEntries = 65535

    // PortMapMax represents the maximum number of Ports Mapping per container.
    PortMapMax = 16

    // EndpointFlagHost indicates that this endpoint represents the host
    // #define ENDPOINT_F_HOST	1 /* Special endpoint representing local host */
    EndpointFlagHost = 1
)

var (
    // LXCMap represents the BPF map for endpoints
    LXCMap = bpf.NewMap(MapName,
        bpf.MapTypeHash,
        &EndpointKey{},
        int(unsafe.Sizeof(EndpointKey{})),
        &EndpointInfo{},
        int(unsafe.Sizeof(EndpointInfo{})),
        MaxEntries,
        0, 0,
        bpf.ConvertKeyValue,
    ).WithCache().WithPressureMetric()
)

// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapKey
type EndpointKey struct {
    bpf.EndpointKey
}

// NewValue returns a new empty instance of the structure representing the BPF
// map value
func (k EndpointKey) NewValue() bpf.MapValue {
    return &EndpointInfo{}
}

func (k *EndpointKey) DeepCopyMapKey() bpf.MapKey {
    //TODO implement me
    panic("implement me")
}

// NewEndpointKey returns an EndpointKey based on the provided IP address. The
// address family is automatically detected
func NewEndpointKey(ip net.IP) *EndpointKey {
    return &EndpointKey{
        EndpointKey: bpf.NewEndpointKey(ip),
    }
}

// EndpointInfo represents the value of the endpoints BPF map.
//
// Must be in sync with struct endpoint_info in <bpf/lib/maps.h>
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapValue
type EndpointInfo struct {
    IfIndex uint32 `align:"ifindex"`
    Unused  uint16 `align:"unused"`
    LxcID   uint16 `align:"lxc_id"`
    Flags   uint32 `align:"flags"`
    // go alignment
    _       uint32
    MAC     mac.Uint64MAC `align:"mac"`
    NodeMAC mac.Uint64MAC `align:"node_mac"`
    Pad     pad4uint32    `align:"pad"`
}

func (v *EndpointInfo) String() string {
    if v.Flags&EndpointFlagHost != 0 {
        return "(localhost)"
    }

    return fmt.Sprintf("id=%-5d flags=0x%04X ifindex=%-3d mac=%s nodemac=%s",
        v.LxcID,
        v.Flags,
        v.IfIndex,
        v.MAC,
        v.NodeMAC,
    )
}

func (v *EndpointInfo) GetValuePtr() unsafe.Pointer {
    return unsafe.Pointer(v)
}

func (v *EndpointInfo) DeepCopyMapValue() bpf.MapValue {
    //TODO implement me
    panic("implement me")
}

type pad4uint32 [4]uint32

// DeepCopyInto is a deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *pad4uint32) DeepCopyInto(out *pad4uint32) {
    copy(out[:], in[:])
    return
}

// EndpointFrontend is the interface to implement for an object to synchronize
// with the endpoint BPF map.
type EndpointFrontend interface {
    LXCMac() mac.MAC
    GetNodeMAC() mac.MAC
    GetIfIndex() int
    GetID() uint64
    IPv4Address() addressing.CiliumIPv4
    //IPv6Address() addressing.CiliumIPv6
}

// WriteEndpoint updates the BPF map with the endpoint information and links
// the endpoint information to all keys provided.
func WriteEndpoint(f EndpointFrontend) error {
    info, err := GetBPFValue(f)
    if err != nil {
        return err
    }

    // FIXME: Revert on failure
    for _, key := range GetBPFKeys(f) {
        if err := LXCMap.Update(key, info); err != nil {
            return err
        }
    }

    return nil
}

// GetBPFKeys returns all keys which should represent this endpoint in the BPF
// endpoints map
func GetBPFKeys(e EndpointFrontend) []*EndpointKey {
    var keys []*EndpointKey
    if e.IPv4Address().IsSet() {
        keys = append(keys, NewEndpointKey(e.IPv4Address().IP()))
    }

    return keys
}

// GetBPFValue returns the value which should represent this endpoint in the
// BPF endpoints map
// Must only be called if init() succeeded.
func GetBPFValue(e EndpointFrontend) (*EndpointInfo, error) {
    lxcMac, err := e.LXCMac().Uint64()
    if err != nil {
        return nil, fmt.Errorf("invalid LXC MAC: %v", err)
    }

    nodeMAC, err := e.GetNodeMAC().Uint64()
    if err != nil {
        return nil, fmt.Errorf("invalid node MAC: %v", err)
    }

    info := &EndpointInfo{
        IfIndex: uint32(e.GetIfIndex()),
        // Store security identity in network byte order so it can be
        // written into the packet without an additional byte order
        // conversion.
        LxcID:   uint16(e.GetID()),
        MAC:     lxcMac,
        NodeMAC: nodeMAC,
    }

    return info, nil
}
