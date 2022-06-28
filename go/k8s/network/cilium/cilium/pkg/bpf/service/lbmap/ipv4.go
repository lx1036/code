package lbmap

import (
	"fmt"
	"github.com/cilium/cilium/pkg/bpf"
	"github.com/cilium/cilium/pkg/loadbalancer"
	"net"
	"unsafe"

	"github.com/cilium/cilium/pkg/u8proto"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/types"
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
