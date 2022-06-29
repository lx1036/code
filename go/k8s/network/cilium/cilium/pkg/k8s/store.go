package k8s

import (
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/k8s/loadbalancer"
)

// PortConfiguration is the L4 port configuration of a frontend or backend. The
// map is indexed by the name of the port and the value constains the L4 port
// and protocol.
type PortConfiguration map[string]*loadbalancer.L4Addr
