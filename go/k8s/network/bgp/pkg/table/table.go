package table

import "github.com/osrg/gobgp/pkg/packet/bgp"

type Table struct {
	routeFamily  bgp.RouteFamily
	destinations map[string]*Destination
}
