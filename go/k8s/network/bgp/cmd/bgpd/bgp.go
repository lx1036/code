package main

import (
	"context"
	"fmt"
	"k8s.io/klog/v2"

	"k8s-lx1036/k8s/network/bgp/pkg/config"
	"k8s-lx1036/k8s/network/bgp/pkg/server"

	api "github.com/osrg/gobgp/api"
)

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
