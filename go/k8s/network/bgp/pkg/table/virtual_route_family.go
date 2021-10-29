package table

import "github.com/osrg/gobgp/pkg/packet/bgp"

type VirtualRouteFamily struct {
	Name      string
	Id        uint32
	Rd        bgp.RouteDistinguisherInterface
	ImportRt  []bgp.ExtendedCommunityInterface
	ExportRt  []bgp.ExtendedCommunityInterface
	MplsLabel uint32
}
