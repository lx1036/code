package backend

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/test/tunnel/vxlan/flannel/pkg/subnet"
)

type Manager interface {
	GetBackend(backendType string) (Backend, error)
}

type Backend interface {
	RegisterNetwork(ctx context.Context, config *subnet.Config) (Network, error)
}

type Network interface {
	Lease() *subnet.Lease
	MTU() int
	Run(ctx context.Context)
}

type ExternalInterface struct {
	Iface     *net.Interface
	IfaceAddr net.IP
	ExtAddr   net.IP
}

type manager struct {
	wg  sync.WaitGroup
	mux sync.Mutex
	ctx context.Context

	subnetMgr subnet.Manager
	extIface  *ExternalInterface
	active    map[string]Backend
}

func NewManager(ctx context.Context, subnetMgr subnet.Manager, extIface *ExternalInterface) Manager {
	return &manager{
		ctx:       ctx,
		subnetMgr: subnetMgr,
		extIface:  extIface,
		active:    make(map[string]Backend),
	}
}

func (m *manager) GetBackend(backendType string) (Backend, error) {
	m.mux.Lock()
	defer m.mux.Unlock()

	// see if one is already running
	backendType = strings.ToLower(backendType)
	if be, ok := m.active[backendType]; ok {
		return be, nil
	}

	// first request, need to create and run it
	constructor, ok := constructors[backendType]
	if !ok {
		return nil, fmt.Errorf("unknown backend type: %v", backendType)
	}

	be, err := constructor(m.subnetMgr, m.extIface)
	if err != nil {
		return nil, err
	}
	m.active[backendType] = be

	return be, nil
}

var constructors = make(map[string]BackendConstructor)

type BackendConstructor func(sm subnet.Manager, ei *ExternalInterface) (Backend, error)

func Register(name string, ctor BackendConstructor) {
	constructors[name] = ctor
}
