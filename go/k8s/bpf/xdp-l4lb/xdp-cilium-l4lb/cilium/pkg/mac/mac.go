package mac

import (
	"fmt"
	"net"
)

// Uint64MAC is the __u64 representation of a MAC address.
// It corresponds to the C mac_t type used in bpf/.
type Uint64MAC uint64

// MAC address is an net.HardwareAddr encapsulation to force cilium to only use MAC-48.
type MAC net.HardwareAddr

func (m MAC) String() string {
	return net.HardwareAddr(m).String()
}

// ParseMAC parses s only as an IEEE 802 MAC-48.
func ParseMAC(s string) (MAC, error) {
	ha, err := net.ParseMAC(s)
	if err != nil {
		return nil, err
	}
	if len(ha) != 6 {
		return nil, fmt.Errorf("invalid MAC address %s", s)
	}

	return MAC(ha), nil
}

// Uint64 returns the MAC in uint64 format. The MAC is represented as little-endian in
// the returned value.
// Example:
//
//	m := MAC([]{0x11, 0x12, 0x23, 0x34, 0x45, 0x56})
//	v, err := m.Uint64()
//	fmt.Printf("0x%X", v) // 0x564534231211
func (m MAC) Uint64() (Uint64MAC, error) {
	if len(m) != 6 {
		return 0, fmt.Errorf("invalid MAC address %s", m.String())
	}

	res := uint64(m[5])<<40 | uint64(m[4])<<32 | uint64(m[3])<<24 |
		uint64(m[2])<<16 | uint64(m[1])<<8 | uint64(m[0])
	return Uint64MAC(res), nil
}
