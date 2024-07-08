package route

import (
	"net"

	"github.com/vishvananda/netlink"
)

type Route struct {
	Prefix   net.IPNet
	Nexthop  *net.IP
	Local    net.IP
	Device   string
	MTU      int
	Priority int
	Proto    int
	Scope    netlink.Scope
	Table    int
	Type     int
}
