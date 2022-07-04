package bpf

import (
	"net"

	"k8s-lx1036/k8s/network/cilium/cilium/pkg/config/types"
)

// Must be in sync with ENDPOINT_KEY_* in <bpf/lib/common.h>
const (
	EndpointKeyIPv4 uint8 = 1
	EndpointKeyIPv6 uint8 = 2
)

// EndpointKey represents the key value of the endpoints BPF map
//
// Must be in sync with struct endpoint_key in <bpf/lib/common.h>
// +k8s:deepcopy-gen=true
type EndpointKey struct {
	// represents both IPv6 and IPv4 (in the lowest four bytes)
	IP     types.IPv6 `align:"$union0"`
	Family uint8      `align:"family"`
	Key    uint8      `align:"key"`
	Pad2   uint16     `align:"pad5"`
}

func NewEndpointKey(ip net.IP) EndpointKey {
	result := EndpointKey{}

	if ip4 := ip.To4(); ip4 != nil {
		result.Family = EndpointKeyIPv4
		copy(result.IP[:], ip4)
	} else {
		result.Family = EndpointKeyIPv6
		copy(result.IP[:], ip)
	}
	result.Key = 0

	return result
}

// ToIP converts the EndpointKey into a net.IP structure.
func (k EndpointKey) ToIP() net.IP {
	switch k.Family {
	case EndpointKeyIPv4:
		return k.IP[:4]
	case EndpointKeyIPv6:
		return k.IP[:]
	}
	return nil
}
