package lbmap

import (
	"fmt"
	"net"
	"unsafe"

	"github.com/cilium/cilium/pkg/u8proto"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/config/types"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/k8s/loadbalancer"
)

var (
	Service4MapV2 = bpf.NewMap("cilium_lb4_services_v2",
		bpf.MapTypeHash,
		&*Service4Key{},
		int(unsafe.Sizeof(*Service4Key{})),
		&Service4Value{},
		int(unsafe.Sizeof(Service4Value{})),
		MaxEntries,
		0, 0,
		func(key []byte, value []byte, mapKey bpf.MapKey, mapValue bpf.MapValue) (bpf.MapKey, bpf.MapValue, error) {
			svcKey, svcVal := mapKey.(**Service4Key), mapValue.(*Service4Value)

			if _, _, err := bpf.ConvertKeyValue(key, value, svcKey, svcVal); err != nil {
				return nil, nil, err
			}

			return svcKey.ToNetwork(), svcVal.ToNetwork(), nil
		}).WithCache()

	Backend4Map = bpf.NewMap("cilium_lb4_backends",
		bpf.MapTypeHash,
		&Backend4Key{},
		int(unsafe.Sizeof(Backend4Key{})),
		&Backend4Value{},
		int(unsafe.Sizeof(Backend4Value{})),
		MaxEntries,
		0, 0,
		func(key []byte, value []byte, mapKey bpf.MapKey, mapValue bpf.MapValue) (bpf.MapKey, bpf.MapValue, error) {
			backendVal := mapValue.(*Backend4Value)

			if _, _, err := bpf.ConvertKeyValue(key, value, mapKey, backendVal); err != nil {
				return nil, nil, err
			}

			return mapKey, backendVal.ToNetwork(), nil
		}).WithCache()

	RevNat4Map = bpf.NewMap("cilium_lb4_reverse_nat",
		bpf.MapTypeHash,
		&RevNat4Key{},
		int(unsafe.Sizeof(RevNat4Key{})),
		&RevNat4Value{},
		int(unsafe.Sizeof(RevNat4Value{})),
		MaxEntries,
		0, 0,
		func(key []byte, value []byte, mapKey bpf.MapKey, mapValue bpf.MapValue) (bpf.MapKey, bpf.MapValue, error) {
			revKey, revNat := mapKey.(*RevNat4Key), mapValue.(*RevNat4Value)

			if _, _, err := bpf.ConvertKeyValue(key, value, revKey, revNat); err != nil {
				return nil, nil, err
			}

			return revKey.ToNetwork(), revNat.ToNetwork(), nil
		}).WithCache()
)

// Service4Key must match 'struct lb4_key' in "bpf/lib/common.h".
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapKey
type Service4Key struct {
	Address types.IPv4 `align:"address"`
	Port    uint16     `align:"dport"`
	Slave   uint16     `align:"slave"`
	Proto   uint8      `align:"proto"`
	Scope   uint8      `align:"scope"`
	Pad     pad2uint8  `align:"pad"`
}

func NewService4Key(ip net.IP, port uint16, proto u8proto.U8proto, scope uint8, slave uint16) *Service4Key {
	key := Service4Key{
		Port:  port,
		Proto: uint8(proto),
		Scope: scope,
		Slave: slave,
	}

	copy(key.Address[:], ip.To4())

	return &key
}

func (k *Service4Key) String() string {
	if k.Scope == loadbalancer.ScopeInternal {
		return fmt.Sprintf("%s:%d/i", k.Address, k.Port)
	} else {
		return fmt.Sprintf("%s:%d", k.Address, k.Port)
	}
}

func (k *Service4Key) GetKeyPtr() unsafe.Pointer {
	return unsafe.Pointer(k)
}

func (k *Service4Key) NewValue() bpf.MapValue {
	return &Service4Value{}
}

func (k *Service4Key) DeepCopyMapKey() bpf.MapKey {
	//TODO implement me
	panic("implement me")
}

// Service4Value must match 'struct lb4_service' in "bpf/lib/common.h".
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapValue
type Service4Value struct {
	BackendID uint32 `align:"backend_id"`
	Count     uint16 `align:"count"`
	RevNat    uint16 `align:"rev_nat_index"`
	Flags     uint8
	Pad       pad3uint8 `align:"pad"`
}

func (s *Service4Value) String() string {
	return fmt.Sprintf("%d (%d) [FLAGS: 0x%x]", s.BackendID, s.RevNat, s.Flags)
}

func (s *Service4Value) GetValuePtr() unsafe.Pointer { return unsafe.Pointer(s) }

func (s *Service4Value) SetCount(count int)   { s.Count = uint16(count) }
func (s *Service4Value) GetCount() int        { return int(s.Count) }
func (s *Service4Value) SetRevNat(id int)     { s.RevNat = uint16(id) }
func (s *Service4Value) GetRevNat() int       { return int(s.RevNat) }
func (s *Service4Value) RevNatKey() RevNatKey { return &RevNat4Key{s.RevNat} }
func (s *Service4Value) SetFlags(flags uint8) { s.Flags = flags }
func (s *Service4Value) GetFlags() uint8      { return s.Flags }
func (s *Service4Value) SetSessionAffinityTimeoutSec(t uint32) {
	// Go doesn't support union types, so we use BackendID to access the
	// lb4_service.affinity_timeout field
	s.BackendID = t
}
func (s *Service4Value) SetBackendID(id loadbalancer.BackendID) {
	s.BackendID = uint32(id)
}
func (s *Service4Value) GetBackendID() loadbalancer.BackendID {
	return loadbalancer.BackendID(s.BackendID)
}
func (s *Service4Value) ToNetwork() Service4Value {
	n := *s
	n.RevNat = byteorder.HostToNetwork(n.RevNat).(uint16)
	return &n
}

// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapKey
type Backend4Key struct {
	ID loadbalancer.BackendID
}

func NewBackend4Key(id loadbalancer.BackendID) *Backend4Key {
	return &Backend4Key{ID: id}
}

func (k *Backend4Key) String() string                  { return fmt.Sprintf("%d", k.ID) }
func (k *Backend4Key) GetKeyPtr() unsafe.Pointer       { return unsafe.Pointer(k) }
func (k *Backend4Key) NewValue() bpf.MapValue          { return &Backend4Value{} }
func (k *Backend4Key) Map() *bpf.Map                   { return Backend4Map }
func (k *Backend4Key) SetID(id loadbalancer.BackendID) { k.ID = id }
func (k *Backend4Key) GetID() loadbalancer.BackendID   { return k.ID }

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

func (v *Backend4Value) String() string {
	return fmt.Sprintf("%s://%s:%d", v.Proto, v.Address, v.Port)
}

func (v *Backend4Value) GetValuePtr() unsafe.Pointer { return unsafe.Pointer(v) }

func (b *Backend4Value) GetAddress() net.IP { return b.Address.IP() }
func (b *Backend4Value) GetPort() uint16    { return b.Port }

func (v *Backend4Value) ToNetwork() *Backend4Value {
	n := *v
	n.Port = byteorder.HostToNetwork(n.Port).(uint16)
	return &n
}

// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapKey
type RevNat4Key struct {
	Key uint16
}

func NewRevNat4Key(value uint16) *RevNat4Key {
	return &RevNat4Key{value}
}

func (k *RevNat4Key) Map() *bpf.Map             { return RevNat4Map }
func (k *RevNat4Key) NewValue() bpf.MapValue    { return &RevNat4Value{} }
func (k *RevNat4Key) GetKeyPtr() unsafe.Pointer { return unsafe.Pointer(k) }
func (k *RevNat4Key) String() string            { return fmt.Sprintf("%d", k.Key) }
func (k *RevNat4Key) GetKey() uint16            { return k.Key }

// ToNetwork converts RevNat4Key to network byte order.
func (k *RevNat4Key) ToNetwork() RevNatKey {
	n := *k
	n.Key = byteorder.HostToNetwork(n.Key).(uint16)
	return &n
}

// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapValue
type RevNat4Value struct {
	Address types.IPv4 `align:"address"`
	Port    uint16     `align:"port"`
}

func (v *RevNat4Value) GetValuePtr() unsafe.Pointer { return unsafe.Pointer(v) }

// ToNetwork converts RevNat4Value to network byte order.
func (v *RevNat4Value) ToNetwork() *RevNat4Value {
	n := *v
	n.Port = byteorder.HostToNetwork(n.Port).(uint16)
	return &n
}

func (v *RevNat4Value) String() string {
	return fmt.Sprintf("%s:%d", v.Address, v.Port)
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

// GetKeyPtr returns the unsafe pointer to the BPF key
func (k *SockRevNat4Key) GetKeyPtr() unsafe.Pointer { return unsafe.Pointer(k) }

// GetValuePtr returns the unsafe pointer to the BPF value
func (v *SockRevNat4Value) GetValuePtr() unsafe.Pointer { return unsafe.Pointer(v) }

// String converts the key into a human readable string format.
func (k *SockRevNat4Key) String() string {
	return fmt.Sprintf("[%s]:%d, %d", k.address, k.port, k.cookie)
}

// String converts the value into a human readable string format.
func (v *SockRevNat4Value) String() string {
	return fmt.Sprintf("[%s]:%d, %d", v.address, v.port, v.revNatIndex)
}

// NewValue returns a new empty instance of the structure representing the BPF
// map value.
func (k SockRevNat4Key) NewValue() bpf.MapValue { return &SockRevNat4Value{} }

const (
	// SockRevNat4MapName is the BPF map name.
	SockRevNat4MapName = "cilium_lb4_reverse_sk"

	// SockRevNat4MapSize is the maximum number of entries in the BPF map.
	SockRevNat4MapSize = 256 * 1024
)

// CreateSockRevNat4Map creates the reverse NAT sock map.
func CreateSockRevNat4Map() error {
	sockRevNat4Map := bpf.NewMap(SockRevNat4MapName,
		bpf.MapTypeLRUHash,
		&SockRevNat4Key{},
		int(unsafe.Sizeof(SockRevNat4Key{})),
		&SockRevNat4Value{},
		int(unsafe.Sizeof(SockRevNat4Value{})),
		MaxSockRevNat4MapEntries,
		0,
		0,
		bpf.ConvertKeyValue,
	)
	_, err := sockRevNat4Map.Create()
	return err
}
