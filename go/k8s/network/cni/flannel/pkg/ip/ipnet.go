package ip

import (
	"bytes"
	"errors"
	"fmt"
	"net"
)

type IP4 uint32

func FromIP(ip net.IP) IP4 {
	ipv4 := ip.To4()

	if ipv4 == nil {
		panic("Address is not an IPv4 address")
	}

	return FromBytes(ipv4)
}

func (ip IP4) Octets() (a, b, c, d byte) {
	a, b, c, d = byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip)
	return
}

func (ip IP4) StringSep(sep string) string {
	a, b, c, d := ip.Octets()
	return fmt.Sprintf("%d%s%d%s%d%s%d", a, sep, b, sep, c, sep, d)
}

func (ip IP4) String() string {
	return ip.ToIP().String()
}

func (ip IP4) ToIP() net.IP {
	return net.IPv4(ip.Octets())
}

func FromBytes(ip []byte) IP4 {
	return IP4(uint32(ip[3]) |
		(uint32(ip[2]) << 8) |
		(uint32(ip[1]) << 16) |
		(uint32(ip[0]) << 24))
}

func FromIP(ip net.IP) IP4 {
	ipv4 := ip.To4()

	if ipv4 == nil {
		panic("Address is not an IPv4 address")
	}

	return FromBytes(ipv4)
}

func ParseIP4(s string) (IP4, error) {
	ip := net.ParseIP(s)
	if ip == nil {
		return IP4(0), errors.New("invalid IP address format")
	}
	return FromIP(ip), nil
}

type IP4Net struct {
	IP        IP4
	PrefixLen uint
}

func (n *IP4Net) IncrementIP() {
	n.IP++
}

func (n IP4Net) ToIPNet() *net.IPNet {
	return &net.IPNet{
		IP:   n.IP.ToIP(),
		Mask: net.CIDRMask(int(n.PrefixLen), 32),
	}
}

func (n IP4Net) Contains(ip IP4) bool {
	return (uint32(n.IP) & n.Mask()) == (uint32(ip) & n.Mask())
}

func (n IP4Net) Equal(other IP4Net) bool {
	return n.IP == other.IP && n.PrefixLen == other.PrefixLen
}

func (n IP4Net) Mask() uint32 {
	var ones uint32 = 0xFFFFFFFF
	return ones << (32 - n.PrefixLen)
}

func (n IP4Net) Next() IP4Net {
	return IP4Net{
		n.IP + (1 << (32 - n.PrefixLen)),
		n.PrefixLen,
	}
}

func (n *IP4Net) UnmarshalJSON(j []byte) error {
	j = bytes.Trim(j, "\"")
	if _, val, err := net.ParseCIDR(string(j)); err != nil {
		return err
	} else {
		*n = FromIPNet(val)
		return nil
	}
}

func (n IP4Net) String() string {
	return fmt.Sprintf("%s/%d", n.IP.String(), n.PrefixLen)
}

func (n IP4Net) StringSep(octetSep, prefixSep string) string {
	return fmt.Sprintf("%s%s%d", n.IP.StringSep(octetSep), prefixSep, n.PrefixLen)
}

func FromIPNet(n *net.IPNet) IP4Net {
	prefixLen, _ := n.Mask.Size()
	return IP4Net{
		FromIP(n.IP),
		uint(prefixLen),
	}
}
