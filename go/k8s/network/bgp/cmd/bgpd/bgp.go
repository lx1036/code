package main

import (
	"context"
	"fmt"
	"k8s.io/klog/v2"

	"k8s-lx1036/k8s/network/bgp/pkg/config"
	"k8s-lx1036/k8s/network/bgp/pkg/server"

	api "github.com/osrg/gobgp/api"
	log "github.com/sirupsen/logrus"
)

func StartBgp(ctx context.Context, bgpServer *server.BgpServer, newConfig *BgpConfigSet, isGracefulRestart bool) (*BgpConfigSet, error) {
	if err := bgpServer.StartBgp(ctx, &api.StartBgpRequest{
		Global: config.NewGlobalFromConfigStruct(&newConfig.Global),
	}); err != nil {
		log.Fatalf("failed to set global config: %s", err)
	}

	for _, c := range newConfig.BmpServers {
		if err := bgpServer.AddBmp(ctx, &api.AddBmpRequest{
			Address:           c.Config.Address,
			Port:              c.Config.Port,
			SysName:           c.Config.SysName,
			SysDescr:          c.Config.SysDescr,
			Policy:            api.AddBmpRequest_MonitoringPolicy(c.Config.RouteMonitoringPolicy.ToInt()),
			StatisticsTimeout: int32(c.Config.StatisticsTimeout),
		}); err != nil {
			log.Fatalf("failed to set bmp config: %s", err)
		}
	}
	for _, c := range newConfig.MrtDump {
		if len(c.Config.FileName) == 0 {
			continue
		}
		if err := bgpServer.EnableMrt(ctx, &api.EnableMrtRequest{
			DumpType:         int32(c.Config.DumpType.ToInt()),
			Filename:         c.Config.FileName,
			DumpInterval:     c.Config.DumpInterval,
			RotationInterval: c.Config.RotationInterval,
		}); err != nil {
			log.Fatalf("failed to set mrt config: %s", err)
		}
	}
	p := config.ConfigSetToRoutingPolicy(newConfig)
	rp, err := table.NewAPIRoutingPolicyFromConfigStruct(p)
	if err != nil {
		log.Warn(err)
	} else {
		bgpServer.SetPolicies(ctx, &api.SetPoliciesRequest{
			DefinedSets: rp.DefinedSets,
			Policies:    rp.Policies,
		})
	}

	assignGlobalpolicy(ctx, bgpServer, &newConfig.Global.ApplyPolicy.Config)

	added := newConfig.Neighbors
	addedPg := newConfig.PeerGroups
	if isGracefulRestart {
		for i, n := range added {
			if n.GracefulRestart.Config.Enabled {
				added[i].GracefulRestart.State.LocalRestarting = true
			}
		}
	}

	//addPeerGroups(ctx, bgpServer, addedPg)
	//addDynamicNeighbors(ctx, bgpServer, newConfig.DynamicNeighbors)
	addNeighbors(ctx, bgpServer, added)
	return newConfig, nil
}

func addNeighbors(ctx context.Context, bgpServer *server.BgpServer, added []config.Neighbor) {
	for _, p := range added {
		klog.Infof(fmt.Sprintf("Peer %v is added", p.State.NeighborAddress))
		if err := bgpServer.AddPeer(ctx, &api.AddPeerRequest{
			Peer: config.NewPeerFromConfigStruct(&p),
		}); err != nil {
			klog.Warning(err)
		}
	}
}
