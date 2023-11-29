package nat

import (
	"fmt"
	"unsafe"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/bpf"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/tuple"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/types"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/u8proto"
)

// NatKey6 is needed to provide NatEntry type to Lookup values
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapKey
type NatKey6 struct {
	tuple.TupleKey6Global
}

// SizeofNatKey6 is the size of the NatKey6 type in bytes.
const SizeofNatKey6 = int(unsafe.Sizeof(NatKey6{}))

// NewValue creates a new bpf.MapValue.
func (k *NatKey6) NewValue() bpf.MapValue { return &NatEntry6{} }

// ToNetwork converts ports to network byte order.
//
// This is necessary to prevent callers from implicitly converting
// the NatKey6 type here into a local key type in the nested
// TupleKey6Global field.
func (k *NatKey6) ToNetwork() NatKey {
	return &NatKey6{
		TupleKey6Global: *k.TupleKey6Global.ToNetwork().(*tuple.TupleKey6Global),
	}
}

// ToHost converts ports to host byte order.
//
// This is necessary to prevent callers from implicitly converting
// the NatKey6 type here into a local key type in the nested
// TupleKey6Global field.
func (k *NatKey6) ToHost() NatKey {
	return &NatKey6{
		TupleKey6Global: *k.TupleKey6Global.ToHost().(*tuple.TupleKey6Global),
	}
}

// GetKeyPtr returns the unsafe.Pointer for k.
func (k *NatKey6) GetKeyPtr() unsafe.Pointer { return unsafe.Pointer(k) }

func (k *NatKey6) GetNextHeader() u8proto.U8proto {
	return k.NextHeader
}

// NatEntry6 represents an IPv6 entry in the NAT table.
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapValue
type NatEntry6 struct {
	Created   uint64     `align:"created"`
	HostLocal uint64     `align:"host_local"`
	Pad1      uint64     `align:"pad1"`
	Pad2      uint64     `align:"pad2"`
	Addr      types.IPv6 `align:"to_saddr"`
	Port      uint16     `align:"to_sport"`
}

// SizeofNatEntry6 is the size of the NatEntry6 type in bytes.
const SizeofNatEntry6 = int(unsafe.Sizeof(NatEntry6{}))

// GetValuePtr returns the unsafe.Pointer for n.
func (n *NatEntry6) GetValuePtr() unsafe.Pointer { return unsafe.Pointer(n) }

// String returns the readable format.
func (n *NatEntry6) String() string {
	return fmt.Sprintf("Addr=%s Port=%d Created=%d HostLocal=%d\n",
		n.Addr,
		n.Port,
		n.Created,
		n.HostLocal)
}

// Dump dumps NAT entry to string.
func (n *NatEntry6) Dump(key NatKey, start uint64) string {
	var which string

	if key.GetFlags()&tuple.TUPLE_F_IN != 0 {
		which = "DST"
	} else {
		which = "SRC"
	}
	return fmt.Sprintf("XLATE_%s [%s]:%d Created=%s HostLocal=%d\n",
		which,
		n.Addr,
		n.Port,
		NatDumpCreated(start, n.Created),
		n.HostLocal)
}

// ToHost converts NatEntry4 ports to host byte order.
func (n *NatEntry6) ToHost() NatEntry {
	x := *n
	x.Port = byteorder.NetworkToHost(n.Port).(uint16)
	return &x
}
