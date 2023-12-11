package ctmap

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
