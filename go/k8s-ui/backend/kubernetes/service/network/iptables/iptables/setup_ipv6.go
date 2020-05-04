package iptables

const (
	bridgeIPv6Str          = "fe80::1/64"
	ipv6ForwardConfPerm    = 0644
	ipv6ForwardConfDefault = "/proc/sys/net/ipv6/conf/default/forwarding"
	ipv6ForwardConfAll     = "/proc/sys/net/ipv6/conf/all/forwarding"
)
