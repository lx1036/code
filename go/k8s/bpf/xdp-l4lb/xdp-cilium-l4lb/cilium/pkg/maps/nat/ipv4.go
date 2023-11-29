package nat

import (
	"fmt"
	"unsafe"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/bpf"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/tuple"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/types"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/u8proto"
)

// NatEntry4 represents an IPv4 entry in the NAT table.
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapValue
type NatEntry4 struct {
	Created   uint64     `align:"created"`
	HostLocal uint64     `align:"host_local"`
	Pad1      uint64     `align:"pad1"`
	Pad2      uint64     `align:"pad2"`
	Addr      types.IPv4 `align:"to_saddr"`
	Port      uint16     `align:"to_sport"`
}

// SizeofNatEntry4 is the size of the NatEntry4 type in bytes.
const SizeofNatEntry4 = int(unsafe.Sizeof(NatEntry4{}))

// GetValuePtr returns the unsafe.Pointer for n.
func (n *NatEntry4) GetValuePtr() unsafe.Pointer { return unsafe.Pointer(n) }

// String returns the readable format.
func (n *NatEntry4) String() string {
	return fmt.Sprintf("Addr=%s Port=%d Created=%d HostLocal=%d\n",
		n.Addr,
		n.Port,
		n.Created,
		n.HostLocal)
}

// Dump dumps NAT entry to string.
func (n *NatEntry4) Dump(key NatKey, start uint64) string {
	var which string

	if key.GetFlags()&tuple.TUPLE_F_IN != 0 {
		which = "DST"
	} else {
		which = "SRC"
	}
	return fmt.Sprintf("XLATE_%s %s:%d Created=%s HostLocal=%d\n",
		which,
		n.Addr,
		n.Port,
		NatDumpCreated(start, n.Created),
		n.HostLocal)
}

// ToHost converts NatEntry4 ports to host byte order.
func (n *NatEntry4) ToHost() NatEntry {
	x := *n
	x.Port = byteorder.NetworkToHost(n.Port).(uint16)
	return &x
}

// NatKey4 is needed to provide NatEntry type to Lookup values
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapKey
type NatKey4 struct {
	tuple.TupleKey4Global
}

// SizeofNatKey4 is the size of the NatKey4 type in bytes.
const SizeofNatKey4 = int(unsafe.Sizeof(NatKey4{}))

// NewValue creates a new bpf.MapValue.
func (k *NatKey4) NewValue() bpf.MapValue { return &NatEntry4{} }

// ToNetwork converts ports to network byte order.
//
// This is necessary to prevent callers from implicitly converting
// the NatKey4 type here into a local key type in the nested
// TupleKey4Global field.
func (k *NatKey4) ToNetwork() NatKey {
	return &NatKey4{
		TupleKey4Global: *k.TupleKey4Global.ToNetwork().(*tuple.TupleKey4Global),
	}
}

// ToHost converts ports to host byte order.
//
// This is necessary to prevent callers from implicitly converting
// the NatKey4 type here into a local key type in the nested
// TupleKey4Global field.
func (k *NatKey4) ToHost() NatKey {
	return &NatKey4{
		TupleKey4Global: *k.TupleKey4Global.ToHost().(*tuple.TupleKey4Global),
	}
}

// GetKeyPtr returns the unsafe.Pointer for k.
func (k *NatKey4) GetKeyPtr() unsafe.Pointer { return unsafe.Pointer(k) }

func (k *NatKey4) GetNextHeader() u8proto.U8proto {
	return k.NextHeader
}
