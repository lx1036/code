package server

import (
	"fmt"
	"context"
	"os"
	
	api "github.com/osrg/gobgp/api"
	"k8s-lx1036/k8s/network/bgp/pkg/config"
)

type mrtWriter struct {
	dead             chan struct{}
	s                *BgpServer
	c                *config.MrtConfig
	file             *os.File
	rotationInterval uint64
	dumpInterval     uint64
}

type mrtManager struct {
	bgpServer *BgpServer
	writer    map[string]*mrtWriter
}

func (m *mrtManager) enable(c *config.MrtConfig) error {
	
	return nil
}

func (server *BgpServer) EnableMrt(ctx context.Context, r *api.EnableMrtRequest) error {
	if r == nil {
		return fmt.Errorf("nil request")
	}
	return server.mgmtOperation(func() error {
		return server.mrtManager.enable(&config.MrtConfig{
			DumpInterval:     r.DumpInterval,
			RotationInterval: r.RotationInterval,
			DumpType:         config.IntToMrtTypeMap[int(r.DumpType)],
			FileName:         r.Filename,
		})
	}, false)
}
