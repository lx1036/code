package table

import (
	"k8s-lx1036/k8s/network/bgp/pkg/config"
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

func NewPeerInfo(g *config.Global, p *config.Neighbor) *PeerInfo {
	clusterID := net.ParseIP(string(p.RouteReflector.State.RouteReflectorClusterId)).To4()
	// exclude zone info
	naddr, _ := net.ResolveIPAddr("ip", p.State.NeighborAddress)
	return &PeerInfo{
		AS:                      p.Config.PeerAs,
		LocalAS:                 g.Config.As,
		LocalID:                 net.ParseIP(g.Config.RouterId).To4(),
		RouteReflectorClient:    p.RouteReflector.Config.RouteReflectorClient,
		Address:                 naddr.IP,
		RouteReflectorClusterID: clusterID,
		MultihopTtl:             p.EbgpMultihop.Config.MultihopTtl,
		Confederation:           p.IsConfederationMember(g),
	}
}

type Destination struct {
	routeFamily   bgp.RouteFamily
	nlri          bgp.AddrPrefixInterface
	knownPathList []*Path
	localIdMap    *Bitmap
}
