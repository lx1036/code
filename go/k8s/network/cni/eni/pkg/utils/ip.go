package utils

import (
	"fmt"
	"net"
)

// ToIP parse str to net.IP and return error is parse failed
func ToIP(addr string) (net.IP, error) {
	ip := net.ParseIP(addr)
	if ip == nil {
		return nil, fmt.Errorf("failed to parse ip %s", addr)
	}
	return ip, nil
}

func IPv6(ip net.IP) bool {
	return ip.To4() == nil
}
