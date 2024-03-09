package lbmap

import (
	"fmt"
	"net"
	"unsafe"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/bpf"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/loadbalancer"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/types"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/u8proto"
)

const (
	// HealthProbe4MapName is the health datapath map name
	HealthProbe4MapName = "cilium_lb4_health"

	// SockRevNat4MapName is the BPF map name.
	SockRevNat4MapName = "cilium_lb4_reverse_sk"

	// SockRevNat4MapSize is the maximum number of entries in the BPF map.
	SockRevNat4MapSize = 256 * 1024

	// Service4MapV2Name is the name of the IPv4 LB Services v2 BPF map.
	Service4MapV2Name = "cilium_lb4_services_v2"
	// Backend4MapName is the name of the IPv4 LB backends BPF map.
	Backend4MapName = "cilium_lb4_backends"
	// Backend4MapNameV2 is the name of the IPv4 LB backends v2 BPF map.
	Backend4MapNameV2 = "cilium_lb4_backends_v2"
	// RevNat4MapName is the name of the IPv4 LB reverse NAT BPF map.
	RevNat4MapName = "cilium_lb4_reverse_nat"
)

var (
	Service4MapV2 *bpf.Map
	Backend4Map   *bpf.Map
	// Backend4MapV2 is the IPv4 LB backends v2 BPF map.
	Backend4MapV2 *bpf.Map
	// RevNat4Map is the IPv4 LB reverse NAT BPF map.
	RevNat4Map *bpf.Map

	MaxSockRevNat4MapEntries = SockRevNat4MapSize
)

func initSVC(params InitParams) {
	if params.IPv4 {
		Service4MapV2 = bpf.NewMap(Service4MapV2Name,
			bpf.MapTypeHash,
			&Service4Key{},
			int(unsafe.Sizeof(Service4Key{})),
			&Service4Value{},
			int(unsafe.Sizeof(Service4Value{})),
			MaxEntries,
			0, 0,
			bpf.ConvertKeyValue,
		).WithCache().WithPressureMetric()
		Backend4Map = bpf.NewMap(Backend4MapName,
			bpf.MapTypeHash,
			&Backend4Key{},
			int(unsafe.Sizeof(Backend4Key{})),
			&Backend4Value{},
			int(unsafe.Sizeof(Backend4Value{})),
			MaxEntries,
			0, 0,
			bpf.ConvertKeyValue,
		).WithCache().WithPressureMetric()
		Backend4MapV2 = bpf.NewMap(Backend4MapNameV2,
			bpf.MapTypeHash,
			&Backend4KeyV2{},
			int(unsafe.Sizeof(Backend4KeyV2{})), // 使用 sizeof() 函数获取对象的字节大小
			&Backend4Value{},
			int(unsafe.Sizeof(Backend4Value{})),
			MaxEntries,
			0, 0,
			bpf.ConvertKeyValue,
		).WithCache().WithPressureMetric()
		RevNat4Map = bpf.NewMap(RevNat4MapName,
			bpf.MapTypeHash,
			&RevNat4Key{},
			int(unsafe.Sizeof(RevNat4Key{})),
			&RevNat4Value{},
			int(unsafe.Sizeof(RevNat4Value{})),
			MaxEntries,
			0, 0,
			bpf.ConvertKeyValue,
		).WithCache().WithPressureMetric()
	}
}

// Service4Key must match 'struct lb4_key' in "bpf/lib/common.h".
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapKey
type Service4Key struct {
	Address     types.IPv4 `align:"address"`
	Port        uint16     `align:"dport"`
	BackendSlot uint16     `align:"backend_slot"`
	Proto       uint8      `align:"proto"`
	Scope       uint8      `align:"scope"`
	Pad         pad2uint8  `align:"pad"`
}

func (k *Service4Key) String() string {
	//TODO implement me
	panic("implement me")
}

func (k *Service4Key) GetKeyPtr() unsafe.Pointer {
	//TODO implement me
	panic("implement me")
}

func (k *Service4Key) NewValue() bpf.MapValue {
	//TODO implement me
	panic("implement me")
}

func (k *Service4Key) DeepCopyMapKey() bpf.MapKey {
	//TODO implement me
	panic("implement me")
}

func (k *Service4Key) IsIPv6() bool {
	//TODO implement me
	panic("implement me")
}

func (k *Service4Key) IsSurrogate() bool {
	//TODO implement me
	panic("implement me")
}

func (k *Service4Key) Map() *bpf.Map {
	return Service4MapV2
}

func (k *Service4Key) SetBackendSlot(slot int) {
	//TODO implement me
	panic("implement me")
}

func (k *Service4Key) GetBackendSlot() int {
	//TODO implement me
	panic("implement me")
}

func (k *Service4Key) SetScope(scope uint8) {
	//TODO implement me
	panic("implement me")
}

func (k *Service4Key) GetScope() uint8 {
	//TODO implement me
	panic("implement me")
}

func (k *Service4Key) GetAddress() net.IP {
	//TODO implement me
	panic("implement me")
}

func (k *Service4Key) GetPort() uint16 {
	//TODO implement me
	panic("implement me")
}

func (k *Service4Key) RevNatValue() RevNatValue {
	//TODO implement me
	panic("implement me")
}

func (k *Service4Key) MapDelete() error {
	//TODO implement me
	panic("implement me")
}

func (k *Service4Key) ToNetwork() ServiceKey {
	//TODO implement me
	panic("implement me")
}

func (k *Service4Key) ToHost() ServiceKey {
	//TODO implement me
	panic("implement me")
}

func NewService4Key(ip net.IP, port uint16, proto u8proto.U8proto, scope uint8, slot uint16) *Service4Key {
	key := Service4Key{
		Port:        port,
		Proto:       uint8(proto),
		Scope:       scope,
		BackendSlot: slot,
	}

	copy(key.Address[:], ip.To4())

	return &key
}

// Service4Value must match 'struct lb4_service' in "bpf/lib/common.h".
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapValue
type Service4Value struct {
	BackendID uint32    `align:"backend_id"`
	Count     uint16    `align:"count"`
	RevNat    uint16    `align:"rev_nat_index"`
	Flags     uint8     `align:"flags"`
	Flags2    uint8     `align:"flags2"`
	Pad       pad2uint8 `align:"pad"`
}

func (v *Service4Value) String() string {
	//TODO implement me
	panic("implement me")
}

func (v *Service4Value) GetValuePtr() unsafe.Pointer {
	//TODO implement me
	panic("implement me")
}

func (v *Service4Value) DeepCopyMapValue() bpf.MapValue {
	//TODO implement me
	panic("implement me")
}

type pad2uint8 [2]uint8

func (in *pad2uint8) DeepCopyInto(out *pad2uint8) {
	copy(out[:], in[:])
	return
}

type Backend4 struct {
	Key   *Backend4Key
	Value *Backend4Value
}

func (b *Backend4) Map() *bpf.Map {
	//TODO implement me
	panic("implement me")
}

func (b *Backend4) GetKey() BackendKey {
	//TODO implement me
	panic("implement me")
}

func (b *Backend4) GetValue() BackendValue {
	//TODO implement me
	panic("implement me")
}

func NewBackend4(id loadbalancer.BackendID, ip net.IP, port uint16, proto u8proto.U8proto) (*Backend4, error) {
	val, err := NewBackend4Value(ip, port, proto)
	if err != nil {
		return nil, err
	}

	return &Backend4{
		Key:   NewBackend4Key(id),
		Value: val,
	}, nil
}

// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapKey
type Backend4Key struct {
	ID loadbalancer.BackendID
}

func NewBackend4Key(id loadbalancer.BackendID) *Backend4Key {
	return &Backend4Key{ID: id}
}

func (b *Backend4Key) String() string {
	//TODO implement me
	panic("implement me")
}

func (b *Backend4Key) GetKeyPtr() unsafe.Pointer {
	//TODO implement me
	panic("implement me")
}

func (b *Backend4Key) NewValue() bpf.MapValue {
	//TODO implement me
	panic("implement me")
}

func (b *Backend4Key) DeepCopyMapKey() bpf.MapKey {
	//TODO implement me
	panic("implement me")
}

// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapKey
type Backend4KeyV2 struct {
	ID uint32
}

func (b *Backend4KeyV2) String() string {
	//TODO implement me
	panic("implement me")
}

func (b *Backend4KeyV2) GetKeyPtr() unsafe.Pointer {
	//TODO implement me
	panic("implement me")
}

func (b *Backend4KeyV2) NewValue() bpf.MapValue {
	//TODO implement me
	panic("implement me")
}

func (b *Backend4KeyV2) DeepCopyMapKey() bpf.MapKey {
	//TODO implement me
	panic("implement me")
}

// Backend4Value must match 'struct lb4_backend' in "bpf/lib/common.h".
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapValue
type Backend4Value struct {
	Address types.IPv4      `align:"address"`
	Port    uint16          `align:"port"`
	Proto   u8proto.U8proto `align:"proto"`
	Pad     uint8           `align:"pad"`
}

func NewBackend4Value(ip net.IP, port uint16, proto u8proto.U8proto) (*Backend4Value, error) {
	ip4 := ip.To4()
	if ip4 == nil {
		return nil, fmt.Errorf("Not an IPv4 address")
	}

	val := Backend4Value{
		Port:  port,
		Proto: proto,
	}
	copy(val.Address[:], ip.To4())

	return &val, nil
}

func (b *Backend4Value) String() string {
	//TODO implement me
	panic("implement me")
}

func (b *Backend4Value) GetValuePtr() unsafe.Pointer {
	//TODO implement me
	panic("implement me")
}

func (b *Backend4Value) DeepCopyMapValue() bpf.MapValue {
	//TODO implement me
	panic("implement me")
}

// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapKey
type RevNat4Key struct {
	Key uint16
}

func (r RevNat4Key) String() string {
	//TODO implement me
	panic("implement me")
}

func (r RevNat4Key) GetKeyPtr() unsafe.Pointer {
	//TODO implement me
	panic("implement me")
}

func (r RevNat4Key) NewValue() bpf.MapValue {
	//TODO implement me
	panic("implement me")
}

func (r RevNat4Key) DeepCopyMapKey() bpf.MapKey {
	//TODO implement me
	panic("implement me")
}

func NewRevNat4Key(value uint16) *RevNat4Key {
	return &RevNat4Key{value}
}

// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapValue
type RevNat4Value struct {
	Address types.IPv4 `align:"address"`
	Port    uint16     `align:"port"`
}

func (r *RevNat4Value) String() string {
	//TODO implement me
	panic("implement me")
}

func (r *RevNat4Value) GetValuePtr() unsafe.Pointer {
	//TODO implement me
	panic("implement me")
}

func (r *RevNat4Value) DeepCopyMapValue() bpf.MapValue {
	//TODO implement me
	panic("implement me")
}

// SockRevNat4Key is the tuple with address, port and cookie used as key in
// the reverse NAT sock map.
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapKey
type SockRevNat4Key struct {
	cookie  uint64     `align:"cookie"`
	address types.IPv4 `align:"address"`
	port    int16      `align:"port"`
	pad     int16      `align:"pad"`
}

// SockRevNat4Value is an entry in the reverse NAT sock map.
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapValue
type SockRevNat4Value struct {
	address     types.IPv4 `align:"address"`
	port        int16      `align:"port"`
	revNatIndex uint16     `align:"rev_nat_index"`
}
