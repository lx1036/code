//go:build linux
// +build linux

package k8s

import (
	"fmt"
	"github.com/moby/ipvs"
	log "github.com/sirupsen/logrus"
	"net"
	"strconv"
	"sync"
)

// IPVS required kernel modules.
const (
	// KernelModuleIPVS is the kernel module "ip_vs"
	KernelModuleIPVS string = "ip_vs"

	// KernelModuleIPVSRR is the kernel module "ip_vs_rr"
	KernelModuleIPVSRR string = "ip_vs_rr"

	// KernelModuleIPVSWRR is the kernel module "ip_vs_wrr"
	KernelModuleIPVSWRR string = "ip_vs_wrr"

	// KernelModuleIPVSSH is the kernel module "ip_vs_sh"
	KernelModuleIPVSSH string = "ip_vs_sh"

	// KernelModuleNfConntrackIPV4 is the module "nf_conntrack_ipv4"
	// "nf_conntrack_ipv4" has been removed since v4.19
	KernelModuleNfConntrackIPV4 string = "nf_conntrack_ipv4"

	// KernelModuleNfConntrack is the kernel module "nf_conntrack"
	KernelModuleNfConntrack string = "nf_conntrack"
)

type Interface interface {
	// flush all virtual servers in host
	Flush() error
	// add a non-existing virtual server, if exist, return error
	AddVirtualServer(*VirtualServer) error
	// update existing virtual server, if not exist, return error
	UpdateVirtualServer(*VirtualServer) error
	// delete a existing virtual server, if not exits, return error
	DeleteVirtualServer(*VirtualServer) error
	// given a partial virtual server, get all information. if not exist, return error
	GetVirtualServer(*VirtualServer) (*VirtualServer, error)
	// list the virtual servers
	GetVirtualServers() ([]*VirtualServer, error)

	// add real server into existing virtual server, if not exist, return error
	AddRealServer(*VirtualServer, *RealServer) error
	// update existing real server, if not exist, return error
	UpdateRealServer(*VirtualServer, *RealServer) error
	// delete existing real server, if not exist, return error
	DeleteRealServer(*VirtualServer, *RealServer) error
	// list real servers
	GetRealServers(*VirtualServer) ([]*RealServer, error)
}

type Runner struct {
	Handle *ipvs.Handle
	Mu     *sync.Mutex
}

func (runner *Runner) Flush() error {
	runner.Mu.Lock()
	defer runner.Mu.Unlock()
	return runner.Handle.Flush()
}

func (runner *Runner) GetVirtualServer(vs *VirtualServer) (*VirtualServer, error) {
	service, err := toIpvsService(vs)
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

func (runner *Runner) AddVirtualServer(vs *VirtualServer) error {
	service, err := toIpvsService(vs)
	if err != nil {
		return err
	}

	runner.Mu.Lock()
	defer runner.Mu.Unlock()
	return runner.Handle.NewService(service)
}

func (runner *Runner) UpdateVirtualServer(vs *VirtualServer) error {
	service, err := toIpvsService(vs)
	if err != nil {
		return err
	}

	runner.Mu.Lock()
	defer runner.Mu.Unlock()
	return runner.Handle.UpdateService(service)
}

func (runner *Runner) DeleteVirtualServer(vs *VirtualServer) error {
	service, err := toIpvsService(vs)
	if err != nil {
		return err
	}

	runner.Mu.Lock()
	defer runner.Mu.Unlock()

	return runner.Handle.DelService(service)
}

func (runner *Runner) GetRealServers(vs *VirtualServer) ([]*RealServer, error) {
	service, err := toIpvsService(vs)
	if err != nil {
		return nil, err
	}

	var realServers []*RealServer
	var errs []error
	runner.Mu.Lock()
	defer runner.Mu.Unlock()
	destinations, err := runner.Handle.GetDestinations(service)
	if err != nil {
		return nil, err
	}
	for _, destination := range destinations {
		realServer, err := toRealServer(destination)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		realServers = append(realServers, realServer)
	}

	if len(errs) != 0 {
		for _, err := range errs {
			log.Errorf("[rs]list real servers error: %v", err)
		}
	}

	if len(destinations) == 0 {
		return nil, fmt.Errorf("virtual server %s has empty real servers", net.JoinHostPort(vs.Address.String(), strconv.Itoa(int(vs.Port))))
	}

	return realServers, nil
}

func (runner *Runner) AddRealServer(vs *VirtualServer, rs *RealServer) error {
	service, err := toIpvsService(vs)
	if err != nil {
		return err
	}

	destination, err := toIpvsDestination(rs)
	if err != nil {
		return err
	}

	runner.Mu.Lock()
	defer runner.Mu.Unlock()
	err = runner.Handle.NewDestination(service, destination)
	if err != nil {
		return err
	}

	return nil
}

func (runner *Runner) DeleteRealServer(vs *VirtualServer, rs *RealServer) error {
	service, err := toIpvsService(vs)
	if err != nil {
		return err
	}

	destination, err := toIpvsDestination(rs)
	if err != nil {
		return err
	}

	runner.Mu.Lock()
	defer runner.Mu.Unlock()
	err = runner.Handle.DelDestination(service, destination)
	if err != nil {
		return err
	}

	return nil
}

func (runner *Runner) UpdateRealServer(vs *VirtualServer, rs *RealServer) error {
	service, err := toIpvsService(vs)
	if err != nil {
		return err
	}

	destination, err := toIpvsDestination(rs)
	if err != nil {
		return err
	}

	runner.Mu.Lock()
	defer runner.Mu.Unlock()
	err = runner.Handle.UpdateDestination(service, destination)
	if err != nil {
		return err
	}

	return nil
}

func New() (Interface, error) {
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
