package table

import "github.com/osrg/gobgp/pkg/packet/bgp"

type originInfo struct {
	nlri               bgp.AddrPrefixInterface
	source             *PeerInfo
	timestamp          int64
	noImplicitWithdraw bool
	isFromExternal     bool
	eor                bool
	stale              bool
}

type Path struct {
	info      *originInfo
	parent    *Path
	pathAttrs []bgp.PathAttributeInterface
	dels      []bgp.BGPAttrType
	attrsHash uint32
	rejected  bool
	// doesn't exist in the adj
	dropped bool

	// For BGP Nexthop Tracking, this field shows if nexthop is invalidated by IGP.
	IsNexthopInvalid bool
	IsWithdraw       bool
}

func (path *Path) GetRouteFamily() bgp.RouteFamily {
	return bgp.AfiSafiToRouteFamily(path.OriginInfo().nlri.AFI(), path.OriginInfo().nlri.SAFI())
}

func (path *Path) OriginInfo() *originInfo {
	return path.root().info
}

func (path *Path) root() *Path {
	p := path
	for p.parent != nil {
		p = p.parent
	}
	return p
}

func (path *Path) IsEOR() bool {
	if path.info != nil && path.info.eor {
		return true
	}
	return false
}

type Bitmap struct {
	bitmap []uint64
}
