package loadbalancer

import (
	"crypto/sha512"
	"fmt"
	"net"

	log "github.com/sirupsen/logrus"
)

// SVCType is a type of a service.
type SVCType string

const (
	SVCTypeNone         = SVCType("NONE")
	SVCTypeHostPort     = SVCType("HostPort")
	SVCTypeClusterIP    = SVCType("ClusterIP")
	SVCTypeNodePort     = SVCType("NodePort")
	SVCTypeExternalIPs  = SVCType("ExternalIPs")
	SVCTypeLoadBalancer = SVCType("LoadBalancer")
)

const (
	NONE = L4Type("NONE")
	// TCP type.
	TCP = L4Type("TCP")
	// UDP type.
	UDP = L4Type("UDP")
)

var (
	// AllProtocols is the list of all supported L4 protocols
	AllProtocols = []L4Type{TCP, UDP}
)

// L4Type name.
type L4Type string

// L4Addr is an abstraction for the backend port with a L4Type, usually tcp or udp, and
// the Port number.
type L4Addr struct {
	Protocol L4Type
	Port     uint16
}

// NewL4Addr creates a new L4Addr.
func NewL4Addr(protocol L4Type, number uint16) *L4Addr {
	return &L4Addr{Protocol: protocol, Port: number}
}

// SVCTrafficPolicy defines which backends are chosen
type SVCTrafficPolicy string

const (
	SVCTrafficPolicyNone    = SVCTrafficPolicy("NONE")
	SVCTrafficPolicyCluster = SVCTrafficPolicy("Cluster")
	SVCTrafficPolicyLocal   = SVCTrafficPolicy("Local")
)

// FEPortName is the name of the frontend's port.
type FEPortName string

// L3n4AddrID is used to store, as an unique L3+L4 plus the assigned ID, in the
// KVStore.
type L3n4AddrID struct {
	L3n4Addr
	ID ID
}

// L3n4Addr is used to store, as an unique L3+L4 address in the KVStore. It also
// includes the lookup scope for frontend addresses which is used in service
// handling for externalTrafficPolicy=Local, that is, Scope{External,Internal}.
type L3n4Addr struct {
	IP net.IP
	L4Addr
	Scope uint8
}

// Hash calculates L3n4Addr's internal SHA256Sum.
func (a L3n4Addr) Hash() string {
	// FIXME: Remove Protocol's omission once we care about protocols.
	protoBak := a.Protocol
	a.Protocol = ""
	defer func() {
		a.Protocol = protoBak
	}()

	str := []byte(fmt.Sprintf("%+v", a))
	return fmt.Sprintf("%x", sha512.Sum512_256(str))
}

// ID is the ID of L3n4Addr endpoint (either service or backend).
type ID uint32

// BackendID is the backend's ID.
type BackendID uint16

// Backend represents load balancer backend.
type Backend struct {
	// ID of the backend
	ID BackendID
	// Node hosting this backend. This is used to determine backends local to
	// a node.
	NodeName string
	L3n4Addr
}

// NewBackend creates the Backend struct instance from given params.
func NewBackend(id BackendID, protocol L4Type, ip net.IP, portNumber uint16) *Backend {
	lbport := NewL4Addr(protocol, portNumber)
	b := Backend{
		ID:       id,
		L3n4Addr: L3n4Addr{IP: ip, L4Addr: *lbport},
	}
	log.WithField("backend", b).Debug("created new LBBackend")

	return &b
}

// SVC is a structure for storing service details.
type SVC struct {
	Frontend                  L3n4AddrID       // SVC frontend addr and an allocated ID
	Backends                  []Backend        // List of service backends
	Type                      SVCType          // Service type
	TrafficPolicy             SVCTrafficPolicy // Service traffic policy
	SessionAffinity           bool
	SessionAffinityTimeoutSec uint32
	HealthCheckNodePort       uint16 // Service health check node port
	Name                      string // Service name
	Namespace                 string // Service namespace
}
