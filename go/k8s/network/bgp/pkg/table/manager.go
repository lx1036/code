package table

import "github.com/osrg/gobgp/pkg/packet/bgp"

type TableManager struct {
	Tables map[bgp.RouteFamily]*Table
	Vrfs   map[string]*VirtualRouteFamily
	rfList []bgp.RouteFamily
}

func NewTableManager(rfList []bgp.RouteFamily) *TableManager {
	t := &TableManager{
		Tables: make(map[bgp.RouteFamily]*Table),
		Vrfs:   make(map[string]*VirtualRouteFamily),
		rfList: rfList,
	}
	for _, rf := range rfList {
		t.Tables[rf] = NewTable(rf)
	}
	return t
}
