package driver

// IPvlanDriver INFO: IPVlan Linux docs：https://www.kernel.org/doc/Documentation/networking/ipvlan.txt
type IPvlanDriver struct {
	name string
	ipv4 bool
	ipv6 bool
}

func NewIPVlanDriver(ipv4, ipv6 bool) *IPvlanDriver {
	return &IPvlanDriver{
		name: "IPVLanL2",
		ipv4: ipv4,
		ipv6: ipv6,
	}
}
