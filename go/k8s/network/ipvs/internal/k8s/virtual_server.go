package k8s

import (
	"errors"
	"fmt"
	"github.com/moby/ipvs"
	"net"
	"strconv"
	"strings"
	"syscall"
)

// vs in ipvs
type VirtualServer struct {
	Address   net.IP
	Protocol  string // 'tcp', 'udp' or 'SCTP'
	Port      uint16
	Scheduler string // 'rr' 或 'wrr' 等负载均衡策略
	Flags     uint32 // ip hash
	Timeout   uint32
}

func (virtualServer *VirtualServer) String() string {
	return fmt.Sprintf("%s://%s", virtualServer.Protocol, net.JoinHostPort(virtualServer.Address.String(), strconv.Itoa(int(virtualServer.Port))))
}

func (virtualServer *VirtualServer) Equal(other *VirtualServer) bool {
	return virtualServer.Address.Equal(other.Address) &&
		virtualServer.Protocol == other.Protocol &&
		virtualServer.Port == other.Port &&
		virtualServer.Scheduler == other.Scheduler &&
		virtualServer.Flags == other.Flags &&
		virtualServer.Timeout == other.Timeout
}

func toVirtualServerProtocol(protocol uint16) string {
	switch protocol {
	case syscall.IPPROTO_TCP:
		return "TCP"
	case syscall.IPPROTO_UDP:
		return "UDP"
	case syscall.IPPROTO_SCTP:
		return "SCTP"
	}

	return ""
}

func toVirtualServer(service *ipvs.Service) (*VirtualServer, error) {
	virtualServer := &VirtualServer{
		Address:   service.Address,
		Protocol:  toVirtualServerProtocol(service.Protocol),
		Port:      service.Port,
		Scheduler: service.SchedName,
		Timeout:   service.Timeout,
	}

	// TODO: fix flags
	virtualServer.Flags = service.Flags

	if virtualServer.Address == nil {
		if service.AddressFamily == syscall.AF_INET {
			service.Address = net.IPv4zero
		} else if service.AddressFamily == syscall.AF_INET6 {
			service.Address = net.IPv6zero
		}
	}

	return virtualServer, nil
}

// convert VirtualServer obj to Service obj
func toIpvsService(virtualServer *VirtualServer) (*ipvs.Service, error) {
	if virtualServer == nil {
		return nil, errors.New("virtual server should be not empty")
	}

	service := &ipvs.Service{
		Address:   virtualServer.Address,
		Protocol:  toIpvsProtocol(virtualServer.Protocol),
		Port:      virtualServer.Port,
		SchedName: virtualServer.Scheduler,
		Flags:     uint32(virtualServer.Flags),
		Timeout:   virtualServer.Timeout,
	}

	if ipv4 := virtualServer.Address.To4(); ipv4 != nil {
		service.AddressFamily = syscall.AF_INET
		service.Netmask = 0xffffffff
	} else {
		service.AddressFamily = syscall.AF_INET6
		service.Netmask = 128
	}

	return service, nil
}

func toIpvsProtocol(protocol string) uint16 {
	protocol = strings.ToLower(protocol)
	switch protocol {
	case "tcp":
		return uint16(syscall.IPPROTO_TCP)
	case "udp":
		return uint16(syscall.IPPROTO_UDP)
	case "sctp":
		return uint16(syscall.IPPROTO_SCTP)
	}

	return 0
}
