package table

import "github.com/osrg/gobgp/pkg/packet/bgp"

type AdjRib struct {
	accepted map[bgp.RouteFamily]int
	table    map[bgp.RouteFamily]*Table
}
