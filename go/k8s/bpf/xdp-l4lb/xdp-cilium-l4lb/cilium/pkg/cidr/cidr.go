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

// ParseCIDR parses the CIDR string using net.ParseCIDR
func ParseCIDR(str string) (*CIDR, error) {
    _, ipnet, err := net.ParseCIDR(str)
    if err != nil {
        return nil, err
    }
    return NewCIDR(ipnet), nil
}
