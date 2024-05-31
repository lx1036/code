package neighborsmap

import (
	"unsafe"

	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/config/types"
)

const (
	// Map4Name is the BPF map name.
	Map4Name = "cilium_nodeport_neigh4"

	MaxEntries = 65535
)

// Key4 is the IPv4 for the IP-to-MAC address mappings.
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapKey
type Key4 struct {
	ipv4 types.IPv4
}

// Value is the MAC address for the IP-to-MAC address mappings.
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapValue
type Value struct {
	macaddr types.MACAddr
	pad     uint16
}

// InitMaps creates the nodeport neighbors maps in the kernel.
func InitMaps(ipv4, ipv6 bool) error {
	if ipv4 {
		neigh4Map := bpf.NewMap(Map4Name,
			bpf.MapTypeLRUHash,
			&Key4{},
			int(unsafe.Sizeof(Key4{})),
			&Value{},
			int(unsafe.Sizeof(Value{})),
			MaxEntries,
			0,
			0,
			bpf.ConvertKeyValue,
		)

		_, err := neigh4Map.Create()
		return err
	}

	return nil
}
