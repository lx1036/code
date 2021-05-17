package hostport

import (
	"net"

	v1 "k8s.io/api/core/v1"
)

// PortMapping represents a network port in a container
type PortMapping struct {
	Name          string
	HostPort      int32
	ContainerPort int32
	Protocol      v1.Protocol
	HostIP        string
}

// PodPortMapping represents a pod's network state and associated container port mappings
type PodPortMapping struct {
	Namespace    string
	Name         string
	PortMappings []*PortMapping
	HostNetwork  bool
	IP           net.IP
}

type hostport struct {
	port     int32
	protocol string
}
