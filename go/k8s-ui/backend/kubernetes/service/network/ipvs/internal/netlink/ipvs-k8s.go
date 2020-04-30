package netlink

import (
	"github.com/moby/ipvs"
	"sync"
)

type Runner struct {
	Handle ipvs.Handle
	Mu     sync.Mutex
}

type VirtualServer struct {
}

type RealServer struct {
}

func (runner *Runner) GetVirtualServer(vs *VirtualServer) (*VirtualServer, error) {

}

func (runner *Runner) GetVirtualServers() ([]*VirtualServer, error) {

}

func (runner *Runner) AddVirtualServer(vs *VirtualServer) error {

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

}
