package cidr

import (
	"net"
)

type CIDR struct {
	*net.IPNet
}

func NewCIDR(ipnet *net.IPNet) *CIDR {
	if ipnet == nil {
		return nil
	}

	return &CIDR{ipnet}
}
