package ctmap

import "fmt"

// mapType is a type of connection tracking map.
type mapType int

const (
	// mapTypeIPv4TCPLocal and friends are map types which correspond to a
	// combination of the following attributes:
	// * IPv4 or IPv6;
	// * TCP or non-TCP (shortened to Any)
	// * Local (endpoint-specific) or global (endpoint-oblivious).
	mapTypeIPv4TCPLocal mapType = iota
	mapTypeIPv6TCPLocal
	mapTypeIPv4TCPGlobal
	mapTypeIPv6TCPGlobal
	mapTypeIPv4AnyLocal
	mapTypeIPv6AnyLocal
	mapTypeIPv4AnyGlobal
	mapTypeIPv6AnyGlobal
	mapTypeMax
)

// String renders the map type into a user-readable string.
func (m mapType) String() string {
	switch m {
	case mapTypeIPv4TCPLocal:
		return "Local IPv4 TCP CT map"
	case mapTypeIPv6TCPLocal:
		return "Local IPv6 TCP CT map"
	case mapTypeIPv4TCPGlobal:
		return "Global IPv4 TCP CT map"
	case mapTypeIPv6TCPGlobal:
		return "Global IPv6 TCP CT map"
	case mapTypeIPv4AnyLocal:
		return "Local IPv4 non-TCP CT map"
	case mapTypeIPv6AnyLocal:
		return "Local IPv6 non-TCP CT map"
	case mapTypeIPv4AnyGlobal:
		return "Global IPv4 non-TCP CT map"
	case mapTypeIPv6AnyGlobal:
		return "Global IPv6 non-TCP CT map"
	}
	return fmt.Sprintf("Unknown (%d)", int(m))
}
