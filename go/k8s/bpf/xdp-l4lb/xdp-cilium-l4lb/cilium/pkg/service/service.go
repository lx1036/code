package service

import (
	"github.com/cilium/cilium/pkg/counter"
	"github.com/cilium/cilium/pkg/lock"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/loadbalancer"
	"sync/atomic"
)

// ServiceID is the service's ID.
type ServiceID uint16

// BackendID is the backend's ID.
type BackendID uint16

// ID is the ID of L3n4Addr endpoint (either service or backend).
type ID uint32

// Service is a service handler. Its main responsibility is to reflect
// service-related changes into BPF maps used by datapath BPF programs.
// The changes can be triggered either by k8s_watcher or directly by
// API calls to the /services endpoint.
// 注意：可以被 k8s_watcher 调用，或者直接被 cli /services 调用
type Service struct {
	lock.RWMutex

	svcByHash map[string]*svcInfo
	svcByID   map[loadbalancer.ID]*svcInfo

	backendRefCount counter.StringCounter
	backendByHash   map[string]*loadbalancer.Backend

	healthServer  healthServer
	monitorNotify monitorNotify

	lbmap         LBMap
	lastUpdatedTs atomic.Value
}

// UpsertService inserts or updates the given service.
//
// The first return value is true if the service hasn't existed before.
func (s *Service) UpsertService(params *loadbalancer.SVC) (bool, loadbalancer.ID, error) {

}
