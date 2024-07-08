package addressing

import (
	"net"
)

type CiliumIPv4 []byte

func (ip CiliumIPv4) IP() net.IP {
	return net.IP(ip)
}

// IsSet returns true if the IP is set
func (ip CiliumIPv4) IsSet() bool {
	return len(ip.String()) != 0
}

func (ip CiliumIPv4) String() string {
	if ip == nil {
		return ""
	}

	return net.IP(ip).String()
}
