package vxlan

import (
	"context"
	"k8s-lx1036/k8s/network/cni/flannel/pkg/ip"
	"net"
	"sync"

	"k8s-lx1036/k8s/network/cni/flannel/pkg/backend"
	"k8s-lx1036/k8s/network/cni/flannel/pkg/subnet"
)

func init() {
	backend.Register("vxlan", New)
}

type VxlanBackend struct {
	subnetMgr subnet.Manager
	extIface  *backend.ExternalInterface
}

func New(subnetMgr subnet.Manager, extIface *backend.ExternalInterface) (backend.Backend, error) {
	return &VxlanBackend{
		subnetMgr: subnetMgr,
		extIface:  extIface,
	}, nil
}

func (backend *VxlanBackend) RegisterNetwork(ctx context.Context, wg *sync.WaitGroup, config *subnet.Config) (backend.Network, error) {

	return newNetwork(backend.subnetMgr, backend.extIface, dev, ip.IP4Net{}, lease)
}

// So we can make it JSON (un)marshalable
type hardwareAddr net.HardwareAddr
