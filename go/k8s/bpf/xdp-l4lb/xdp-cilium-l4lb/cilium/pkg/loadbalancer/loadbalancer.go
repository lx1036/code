package loadbalancer

import (
	"net"
)

type L4Type = string

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

// ID is the ID of L3n4Addr endpoint (either service or backend).
type ID uint32

type BackendID uint16

// +deepequal-gen=true
// +deepequal-gen:private-method=true
type L4Addr struct {
	Protocol L4Type
	Port     uint16
}

func NewL4Addr(protocol L4Type, number uint16) *L4Addr {
	return &L4Addr{Protocol: protocol, Port: number}
}

// L3n4Addr Scope{External,Internal}
// +deepequal-gen=true
// +deepequal-gen:private-method=true
type L3n4Addr struct {
	// +deepequal-gen=false
	IP net.IP
	L4Addr
	Scope uint8
}

func NewL3n4Addr(protocol L4Type, ip net.IP, portNumber uint16, scope uint8) *L3n4Addr {
	lbport := NewL4Addr(protocol, portNumber)

	addr := L3n4Addr{IP: ip, L4Addr: *lbport, Scope: scope}

	return &addr
}

// L3n4AddrID is used to store, as an unique L3+L4 plus the assigned ID, in the KVStore.
//
// +deepequal-gen=true
// +deepequal-gen:private-method=true
type L3n4AddrID struct {
	L3n4Addr
	ID ID
}

func NewL3n4AddrID(protocol L4Type, ip net.IP, portNumber uint16, scope uint8, id ID) *L3n4AddrID {
	l3n4Addr := NewL3n4Addr(protocol, ip, portNumber, scope)
	return &L3n4AddrID{L3n4Addr: *l3n4Addr, ID: id}
}

type SVCType string

const (
	SVCTypeNone          = SVCType("NONE")
	SVCTypeHostPort      = SVCType("HostPort")
	SVCTypeClusterIP     = SVCType("ClusterIP")
	SVCTypeNodePort      = SVCType("NodePort")
	SVCTypeExternalIPs   = SVCType("ExternalIPs")
	SVCTypeLoadBalancer  = SVCType("LoadBalancer")
	SVCTypeLocalRedirect = SVCType("LocalRedirect")
)

type SVCTrafficPolicy string

const (
	SVCTrafficPolicyNone    = SVCTrafficPolicy("NONE")
	SVCTrafficPolicyCluster = SVCTrafficPolicy("Cluster")
	SVCTrafficPolicyLocal   = SVCTrafficPolicy("Local")
)

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
	LoadBalancerSourceRanges  []*cidr.CIDR
}

// Backend represents load balancer backend.
type Backend struct {
	// ID of the backend
	ID BackendID
	// Node hosting this backend. This is used to determine backends local to
	// a node.
	NodeName string
	L3n4Addr
}

func NewBackend(id BackendID, protocol L4Type, ip net.IP, portNumber uint16) *Backend {
	lbport := NewL4Addr(protocol, portNumber)
	b := Backend{
		ID:       BackendID(id),
		L3n4Addr: L3n4Addr{IP: ip, L4Addr: *lbport},
	}

	return &b
}
