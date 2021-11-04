package bgp

import (
	"context"
	"net"
	"testing"
	"time"

	api "github.com/osrg/gobgp/api"
	gobgp "github.com/osrg/gobgp/pkg/server"

	"github.com/golang/protobuf/ptypes"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/anypb"
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

	sess := NewSession("127.0.0.1:1790", net.ParseIP("10.20.30.40"), 64543, 64543, net.ParseIP("2.3.4.5"), 10*time.Second)
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

func TestBGP(test *testing.T) {
	peer, err := runBGPRouteServer(context.TODO(), 1799)
	if err != nil {
		klog.Fatal(err)
	}

	for p := range peer {
		klog.Info(p.String())
	}
}

// `gobgp neighbor add 127.0.0.3 -p 50063 -d`
func runBGPRouteServer(ctx context.Context, port int32) (chan *api.Peer, error) {
	s := gobgp.NewBgpServer(gobgp.GrpcListenAddress(":50063"))
	go s.Serve()

	global := &api.StartBgpRequest{
		Global: &api.Global{
			As:              64543,
			RouterId:        "1.2.3.4",
			ListenPort:      port,
			ListenAddresses: []string{"127.0.0.1"},
		},
	}
	if err := s.StartBgp(ctx, global); err != nil {
		return nil, err
	}

	/*paths := make(chan *api.Path, 1000)
	newPath := func(path *api.Path) {
		klog.Info(path.String())
		paths <- path
	}
	w := &api.MonitorTableRequest{
		TableType:    api.TableType_GLOBAL,
		Name:    "monitor bgp",
		Current: true,
	}
	if err := s.MonitorTable(context.Background(), w, newPath); err != nil {
		return nil, err
	}*/

	peers := make(chan *api.Peer, 1000)
	newPeer := func(peer *api.Peer) {
		klog.Info(peer.String())
		peers <- peer
	}
	if err := s.MonitorPeer(context.TODO(), &api.MonitorPeerRequest{}, newPeer); err != nil {
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
				RemotePort:  1792,
			},
		},
	}
	if err := s.AddPeer(context.Background(), peer); err != nil {
		return nil, err
	}

	return peers, nil
}

// debug in mac local:
// gobgp neighbor -p 50064 127.0.0.1
// gobgp neighbor -p 50065 127.0.0.1
// gobgp -p 50064 global rib
// gobgp -p 50065 global rib
// gobgp -p 50065 global rib add -a ipv4 100.0.0.0/24 nexthop 20.20.20.20
func TestBGPMonitor(test *testing.T) {
	log.SetLevel(log.DebugLevel)

	// bgp1
	s := gobgp.NewBgpServer(gobgp.GrpcListenAddress(":50064"))
	go s.Serve()
	_ = s.StartBgp(context.Background(), &api.StartBgpRequest{
		Global: &api.Global{
			As:         1,
			RouterId:   "1.1.1.1",
			ListenPort: 10179,
		},
	})
	defer s.StopBgp(context.Background(), &api.StopBgpRequest{})
	p1 := &api.Peer{
		Conf: &api.PeerConf{
			NeighborAddress: "127.0.0.1",
			PeerAs:          2,
		},
		Transport: &api.Transport{
			PassiveMode: true,
			RemotePort:  10180,
		},
	}
	_ = s.AddPeer(context.Background(), &api.AddPeerRequest{Peer: p1})

	nlri, _ := ptypes.MarshalAny(&api.IPAddressPrefix{
		Prefix:    "10.20.30.40",
		PrefixLen: 32,
	})
	a1, _ := ptypes.MarshalAny(&api.OriginAttribute{
		Origin: 0,
	})
	a2, _ := ptypes.MarshalAny(&api.NextHopAttribute{
		NextHop: "1.1.1.1",
	})
	attrs := []*anypb.Any{a1, a2}
	s.AddPath(context.TODO(), &api.AddPathRequest{
		Path: &api.Path{
			Family: &api.Family{Afi: api.Family_AFI_IP, Safi: api.Family_SAFI_UNICAST},
			Nlri:   nlri,
			Pattrs: attrs,
		},
	})

	// bgp2
	t := gobgp.NewBgpServer(gobgp.GrpcListenAddress(":50065"))
	go t.Serve()
	_ = t.StartBgp(context.Background(), &api.StartBgpRequest{
		Global: &api.Global{
			As:         2,
			RouterId:   "2.2.2.2",
			ListenPort: 10180,
		},
	})
	defer t.StopBgp(context.Background(), &api.StopBgpRequest{})

	p2 := &api.Peer{
		Conf: &api.PeerConf{
			NeighborAddress: "127.0.0.1",
			PeerAs:          1,
		},
		Transport: &api.Transport{
			RemotePort: 10179,
		},
		Timers: &api.Timers{
			Config: &api.TimersConfig{
				ConnectRetry:           1,
				IdleHoldTimeAfterReset: 1,
			},
		},
	}
	ch := make(chan struct{})
	go t.MonitorPeer(context.Background(), &api.MonitorPeerRequest{}, func(peer *api.Peer) {
		if peer.State.SessionState == api.PeerState_ESTABLISHED {
			klog.Info(peer.String())
			//close(ch)
		}
	})

	// gobgp -p 50065 global rib add -a ipv4 100.0.0.0/24 nexthop 20.20.20.20
	// gobgp -p 50064 global rib
	// gobgp -p 50065 global rib summary
	go t.MonitorTable(context.TODO(), &api.MonitorTableRequest{}, func(path *api.Path) {
		klog.Info(path.String())
	})

	_ = t.AddPeer(context.Background(), &api.AddPeerRequest{Peer: p2})

	<-ch
}


// gobgpd -f ./route-server-conf.conf -l debug --api-hosts ":50052" --pprof-disable
// gobgpd -f ./route-client-conf.conf -l debug --api-hosts ":50053" --pprof-disable
// node这边添加路由：gobgp -p 50053 -d global rib add -a ipv4 100.0.0.0/24 nexthop 20.20.20.20
// 验证交换机那边是否收到路由：gobgp -p 50052 -d neighbor 127.0.0.1 adj-in
// 交换机这边添加路由：gobgp -p 50052 -d global rib add -a ipv4 200.0.0.0/24 nexthop 20.20.20.20
// 验证node这边是否收到路由，不应该收到路由：gobgp -p 50053 -d neighbor 127.0.0.1 adj-in
func TestRouteServer(test *testing.T) {
	log.SetLevel(log.DebugLevel)
	ch := make(chan struct{})
	
	// bgp1
	s := gobgp.NewBgpServer(gobgp.GrpcListenAddress(":50053"))
	go s.Serve()
	_ = s.StartBgp(context.Background(), &api.StartBgpRequest{
		Global: &api.Global{
			As:         65001, // AS Number, 公司内需要调用 NetOPS API 会给本机和交换机 AS Number
			RouterId:   "2.2.2.2", // 一般选择当前机器 IP
			ListenPort: 1791,
		},
	})
	defer s.StopBgp(context.Background(), &api.StopBgpRequest{})
	
	// 把 route-server(即交换机那一端) 加入到本地 bgp-server 中来
	p1 := &api.Peer{
		Conf: &api.PeerConf{
			NeighborAddress: "127.0.0.1",
			PeerAs:          64512,
		},
		Transport: &api.Transport{
			RemotePort:  1790,
		},
	}
	_ = s.AddPeer(context.Background(), &api.AddPeerRequest{Peer: p1})
	
	nlri, _ := ptypes.MarshalAny(&api.IPAddressPrefix{
		Prefix:    "10.20.30.0",
		PrefixLen: 24,
	})
	a1, _ := ptypes.MarshalAny(&api.OriginAttribute{
		Origin: 0,
	})
	a2, _ := ptypes.MarshalAny(&api.NextHopAttribute{
		NextHop: "30.30.30.30",
	})
	attrs := []*anypb.Any{a1, a2}
	s.AddPath(context.TODO(), &api.AddPathRequest{
		Path: &api.Path{
			Family: &api.Family{Afi: api.Family_AFI_IP, Safi: api.Family_SAFI_UNICAST},
			Nlri:   nlri,
			Pattrs: attrs,
		},
	})
	
	<-ch
}
