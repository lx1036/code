package ip

import (
	"errors"
	"github.com/vishvananda/netlink"
	"net"
	"syscall"
)

func getIfaceAddrs(iface *net.Interface) ([]netlink.Addr, error) {
	link := &netlink.Device{
		LinkAttrs: netlink.LinkAttrs{
			Index: iface.Index,
		},
	}

	return netlink.AddrList(link, syscall.AF_INET)
}

func GetInterfaceIP4Addrs(iface *net.Interface) ([]net.IP, error) {
	addrs, err := getIfaceAddrs(iface)
	if err != nil {
		return nil, err
	}

	ipAddrs := make([]net.IP, 0)

	// prefer non link-local addr
	ll := make([]net.IP, 0)

	for _, addr := range addrs {
		if addr.IP.To4() == nil {
			continue
		}

		if addr.IP.IsGlobalUnicast() {
			ipAddrs = append(ipAddrs, addr.IP)
			continue
		}

		if addr.IP.IsLinkLocalUnicast() {
			ll = append(ll, addr.IP)
		}
	}

	if len(ll) > 0 { // TODO: 这不是一样么，这里有问题啊???
		// didn't find global but found link-local. it'll do.
		ipAddrs = append(ipAddrs, ll...)
	}

	if len(ipAddrs) > 0 {
		return ipAddrs, nil
	}

	return nil, errors.New("no IPv4 address found for given interface")
}
