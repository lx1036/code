// +build linux

package netlink

import (
	"errors"
	"fmt"
	"github.com/moby/ipvs"
	log "github.com/sirupsen/logrus"
	"net"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

type Runner struct {
	Handle *ipvs.Handle
	Mu     *sync.Mutex
}

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

// ServiceFlags is used to specify session affinity, ip hash etc.
type ServiceFlags uint32

type RealServer struct {
	Address            net.IP
	Port               uint16
	Weight             int
	ActiveConnection   int
	InActiveConnection int
}

func toIPVSProtocol(protocol string) uint16 {
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

func toIPVSService(virtualServer *VirtualServer) (*ipvs.Service, error) {
	if virtualServer == nil {
		return nil, errors.New("virtual server should be not empty")
	}

	service := &ipvs.Service{
		Address:   virtualServer.Address,
		Protocol:  toIPVSProtocol(virtualServer.Protocol),
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

func (runner *Runner) GetVirtualServer(vs *VirtualServer) (*VirtualServer, error) {
	service, err := toIPVSService(vs)
	if err != nil {
		return nil, err
	}

	runner.Mu.Lock()
	service, err = runner.Handle.GetService(service)
	runner.Mu.Unlock()
	if err != nil {
		return nil, err
	}

	virtualServer, err := toVirtualServer(service)
	if err != nil {
		return nil, err
	}

	return virtualServer, nil
}

func (runner *Runner) GetVirtualServers() ([]*VirtualServer, error) {
	var services []*ipvs.Service
	var err error

	runner.Mu.Lock()
	services, err = runner.Handle.GetServices()
	runner.Mu.Unlock()
	if err != nil {
		return nil, err
	}

	var virtualServers []*VirtualServer
	for _, service := range services {
		virtualServer, err := toVirtualServer(service)
		if err != nil {
			log.Warnf("service [%s] can't be converted to virtual server", service.Address)
			continue
		}
		virtualServers = append(virtualServers, virtualServer)
	}

	return virtualServers, nil
}

/*func (runner *Runner) AddVirtualServer(vs *VirtualServer) error {

}

func (runner *Runner) UpdateVirtualServer(vs *VirtualServer) error {

}

func (runner *Runner) DeleteVirtualServer(vs *VirtualServer) error {

}

func (runner *Runner) GetRealServers(vs *VirtualServer) ([]*RealServer, error) {

}

func (runner *Runner) AddRealServer(vs *VirtualServer, rs *RealServer) error {

}

func (runner *Runner) DeleteRealServer(vs *VirtualServer, rs *RealServer) error {

}

func (runner *Runner) UpdateRealServer(vs *VirtualServer, rs *RealServer) error {

}*/

func New() (*Runner, error) {
	handle, err := ipvs.New("")
	if err != nil {
		return nil, err
	}

	mu := &sync.Mutex{}

	runner := &Runner{
		Handle: handle,
		Mu:     mu,
	}

	return runner, nil
}
