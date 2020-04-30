// +build linux

package netlink

import (
	"errors"
	libipvs "github.com/moby/ipvs"
	"net"
	"strings"
	"sync"
	"syscall"
	"testing"
)

const (
	// FlagPersistent specify IPVS service session affinity
	FlagPersistent = 0x1
	// FlagHashed specify IPVS service hash flag
	FlagHashed = 0x2
	// IPVSProxyMode is match set up cluster with ipvs proxy model
	IPVSProxyMode = "ipvs"
)

// VirtualServer is an user-oriented definition of an IPVS virtual server in its entirety.
type VirtualServer struct {
	Address   net.IP
	Protocol  string
	Port      uint16
	Scheduler string // 负载均衡策略，如 'wrr'
	Flags     ServiceFlags
	Timeout   uint32
}

// ServiceFlags is used to specify session affinity, ip hash etc.
type ServiceFlags uint32

func TestAddVirtualServer(test *testing.T) {
	vs := &VirtualServer{
		Address:   net.ParseIP("192.168.31.25"),
		Protocol:  "TCP",
		Port:      8088,
		Scheduler: "wrr",
		Flags:     ServiceFlags(FlagPersistent),
		Timeout:   0,
	}
	service, err := toIPVSService(vs)
	if err != nil {
		panic(err)
	}
	handle, err := libipvs.New("")
	if err != nil {
		panic(err)
	}

	mu := sync.Mutex{}
	mu.Lock()
	defer mu.Unlock()
	err = handle.NewService(service)
	if err != nil {
		panic(err)
	}
}

// toIPVSService converts a VirtualServer to the equivalent IPVS Service structure.
func toIPVSService(vs *VirtualServer) (*libipvs.Service, error) {
	if vs == nil {
		return nil, errors.New("virtual server should not be empty")
	}
	ipvsSvc := &libipvs.Service{
		Address:       vs.Address,
		Protocol:      stringToProtocol(vs.Protocol),
		Port:          vs.Port,
		SchedName:     vs.Scheduler,
		Flags:         uint32(vs.Flags),
		Timeout:       vs.Timeout,
		AddressFamily: syscall.AF_INET,
		Netmask:       0xffffffff,
		PEName:        "",
	}

	if ip4 := vs.Address.To4(); ip4 != nil {
		ipvsSvc.AddressFamily = syscall.AF_INET
		ipvsSvc.Netmask = 0xffffffff
	} else {
		ipvsSvc.AddressFamily = syscall.AF_INET6
		ipvsSvc.Netmask = 128
	}
	return ipvsSvc, nil
}

// stringToProtocolType returns the protocol type for the given name
func stringToProtocol(protocol string) uint16 {
	switch strings.ToLower(protocol) {
	case "tcp":
		return uint16(syscall.IPPROTO_TCP)
	case "udp":
		return uint16(syscall.IPPROTO_UDP)
	case "sctp":
		return uint16(syscall.IPPROTO_SCTP)
	}
	return uint16(0)
}
