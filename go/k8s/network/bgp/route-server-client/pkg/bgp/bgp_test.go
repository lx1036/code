package bgp

import (
	"context"
	"net"
	"testing"
	"time"
	
	api "github.com/osrg/gobgp/api"
	gobgp "github.com/osrg/gobgp/pkg/server"
	
	"k8s.io/klog/v2"
)

// https://github.com/metallb/metallb/blob/main/internal/bgp/native/native_test.go

func TestTCP(test *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	
	paths, err := runBGPRouteServer(ctx, 1790)
	if err != nil {
		test.Fatalf("starting GoBGP: %s", err)
	}
	
	sess := NewSession( "127.0.0.1:1790", net.ParseIP("10.20.30.40"), 64543, 64543, net.ParseIP("2.3.4.5"), 10*time.Second)
	defer sess.Close()
	
	adv := &Advertisement{
		Prefix:      ipnet("1.2.3.0/24"),
		NextHop:     net.ParseIP("10.20.30.40"),
		LocalPref:   42,
		Communities: []uint32{1234, 2345},
	}
	if err = sess.Set(adv); err != nil {
		test.Fatalf("setting advertisement: %s", err)
	}
	
	
	for {
		select {
		case <-ctx.Done():
			test.Fatalf("test timed out waiting for route")
		case path := <-paths:
			klog.Infof(path.String())
			/*if err := checkPath(path, adv); err != nil {
				test.Fatalf("path did not match expectations: %s", err)
			}*/
			return
		}
	}
}

func runBGPRouteServer(ctx context.Context, port int32) (chan *api.Path, error) {
	s := gobgp.NewBgpServer()
	go s.Serve()
	
	global := &api.StartBgpRequest{
		Global: &api.Global{
			As:         64543,
			RouterId:   "1.2.3.4",
			ListenPort: port,
			ListenAddresses: []string{"127.0.0.1"},
		},
	}
	if err := s.StartBgp(ctx, global); err != nil {
		return nil, err
	}
	
	peer := &api.AddPeerRequest{
		Peer: &api.Peer{
			Conf: &api.PeerConf{
				NeighborAddress: "127.0.0.1",
				PeerAs:          64543,
			},
			Transport: &api.Transport{
				PassiveMode: true,
			},
		},
	}
	if err := s.AddPeer(context.Background(), peer); err != nil {
		return nil, err
	}
	
	paths := make(chan *api.Path, 1000)
	newPath := func(path *api.Path) {
		paths <- path
	}
	w := &api.MonitorTableRequest{
		TableType:    api.TableType_GLOBAL,
		Name:    "monitor bgp",
		Current: true,
	}
	if err := s.MonitorTable(context.Background(), w, newPath); err != nil {
		return nil, err
	}
	
	return paths, nil
}
