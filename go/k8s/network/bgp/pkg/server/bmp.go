package server

import (
	"context"
	"fmt"

	"k8s-lx1036/k8s/network/bgp/pkg/config"
	"k8s-lx1036/k8s/network/bgp/pkg/table"

	api "github.com/osrg/gobgp/api"
)

type ribout map[string][]*table.Path

func newribout() ribout {
	return make(map[string][]*table.Path)
}

type bmpClient struct {
	s      *BgpServer
	dead   chan struct{}
	host   string
	c      *config.BmpServerConfig
	ribout ribout
}

type bmpClientManager struct {
	s         *BgpServer
	clientMap map[string]*bmpClient
}

func newBmpClientManager(s *BgpServer) *bmpClientManager {
	return &bmpClientManager{
		s:         s,
		clientMap: make(map[string]*bmpClient),
	}
}

func (s *BgpServer) AddBmp(ctx context.Context, r *api.AddBmpRequest) error {
	if r == nil {
		return fmt.Errorf("nil request")
	}
	return s.mgmtOperation(func() error {
		_, ok := api.AddBmpRequest_MonitoringPolicy_name[int32(r.Policy)]
		if !ok {
			return fmt.Errorf("invalid bmp route monitoring policy: %v", r.Policy)
		}
		return s.bmpManager.addServer(&config.BmpServerConfig{
			Address:               r.Address,
			Port:                  r.Port,
			SysName:               r.SysName,
			SysDescr:              r.SysDescr,
			RouteMonitoringPolicy: config.IntToBmpRouteMonitoringPolicyTypeMap[int(r.Policy)],
			StatisticsTimeout:     uint16(r.StatisticsTimeout),
		})
	}, true)
}
