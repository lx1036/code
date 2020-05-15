package k8s

import (
	"errors"
	"github.com/moby/ipvs"
	"net"
	"strconv"
)

// rs in ipvs
type RealServer struct {
	Address             net.IP
	Port                uint16
	Weight              int
	ActiveConnections   int
	InactiveConnections int
}

func (realServer *RealServer) String() string {
	return net.JoinHostPort(realServer.Address.String(), strconv.Itoa(int(realServer.Port)))
}

func (realServer *RealServer) Equal(other *RealServer) bool {
	return realServer.Address.Equal(other.Address) && realServer.Port == other.Port
}

func toRealServer(destination *ipvs.Destination) (*RealServer, error) {
	if destination == nil {
		return nil, errors.New("ipvs destination should not be empty")
	}
	return &RealServer{
		Address:             destination.Address,
		Port:                destination.Port,
		Weight:              destination.Weight,
		ActiveConnections:   destination.ActiveConnections,
		InactiveConnections: destination.InactiveConnections,
	}, nil
}

func toIpvsDestination(server *RealServer) (*ipvs.Destination, error) {
	if server == nil {
		return nil, errors.New("ipvs real server should not be empty")
	}

	return &ipvs.Destination{
		Address:             server.Address,
		Port:                server.Port,
		Weight:              server.Weight,
		ActiveConnections:   server.ActiveConnections,
		InactiveConnections: server.InactiveConnections,
	}, nil
}
