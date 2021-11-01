package table

import (
	"net"

	"github.com/osrg/gobgp/pkg/packet/bgp"
)

type PeerInfo struct {
	AS                      uint32
	ID                      net.IP
	LocalAS                 uint32
	LocalID                 net.IP
	Address                 net.IP
	LocalAddress            net.IP
	RouteReflectorClient    bool
	RouteReflectorClusterID net.IP
	MultihopTtl             uint8
	Confederation           bool
}

type Destination struct {
	routeFamily   bgp.RouteFamily
	nlri          bgp.AddrPrefixInterface
	knownPathList []*Path
	localIdMap    *Bitmap
}
