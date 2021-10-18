package types

import (
	"net"
)

type IPNetSet struct {
	IPv4 *net.IPNet
	IPv6 *net.IPNet
}

// IPSet is the type hole both ipv4 and ipv6 net.IP
type IPSet struct {
	IPv4 net.IP
	IPv6 net.IP
}

type ENI struct {
	ID               string
	MAC              string
	MaxIPs           int
	SecurityGroupIDs []string

	Trunk bool

	PrimaryIP IPSet
	GatewayIP IPSet

	VSwitchCIDR IPNetSet

	VSwitchID string
}
