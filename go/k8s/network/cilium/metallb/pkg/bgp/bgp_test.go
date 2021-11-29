package bgp

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	gobgpapi "github.com/osrg/gobgp/api"
	gobgp "github.com/osrg/gobgp/pkg/server"

	"github.com/golang/protobuf/ptypes"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"google.golang.org/protobuf/types/known/anypb"
	"k8s.io/klog/v2"
)

// https://github.com/metallb/metallb/blob/main/internal/bgp/native/native_test.go

// INFO: 查看接收的路由：gobgp -p 50063 -d neighbor 127.0.0.1 adj-in
func TestBGPClient(test *testing.T) {
	log.SetLevel(log.DebugLevel)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()

	paths, err := runBGPRouteServer(ctx, 1790)
	if err != nil {
		test.Fatalf("starting BGP Server: %s", err)
	}

	sess, err := New("127.0.0.1:1790", net.ParseIP("127.0.0.1"), 64543, 64544, time.Second*10, "")
	//defer sess.Close()
	if err != nil {
		test.Fatalf("starting BGP Client: %s", err)
	}

	adv := &Advertisement{
		Prefix:      ipnet("1.2.3.0/24"),
		NextHop:     net.ParseIP("10.20.30.40"),
		LocalPref:   42,
		Communities: []uint32{1234, 2345},
	}
	if err = sess.AddPath(adv); err != nil {
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
			//return
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
func runBGPRouteServer(ctx context.Context, port int32) (chan *gobgpapi.Peer, error) {
	s := gobgp.NewBgpServer(gobgp.GrpcListenAddress(":50063"))
	go s.Serve()

	global := &gobgpapi.StartBgpRequest{
		Global: &gobgpapi.Global{
			As:              64543,
			RouterId:        "1.2.3.4",
			ListenPort:      port,
			ListenAddresses: []string{"127.0.0.1"},
		},
	}
	if err := s.StartBgp(ctx, global); err != nil {
		return nil, err
	}

	/*paths := make(chan *gobgpapi.Path, 1000)
	newPath := func(path *gobgpapi.Path) {
		klog.Info(path.String())
		paths <- path
	}
	w := &gobgpapi.MonitorTableRequest{
		TableType:    gobgpapi.TableType_GLOBAL,
		Name:    "monitor bgp",
		Current: true,
	}
	if err := s.MonitorTable(context.Background(), w, newPath); err != nil {
		return nil, err
	}*/

	peers := make(chan *gobgpapi.Peer, 1000)
	newPeer := func(peer *gobgpapi.Peer) {
		klog.Info(peer.String())
		peers <- peer
	}
	if err := s.MonitorPeer(context.TODO(), &gobgpapi.MonitorPeerRequest{}, newPeer); err != nil {
		return nil, err
	}

	peer := &gobgpapi.AddPeerRequest{
		Peer: &gobgpapi.Peer{
			Conf: &gobgpapi.PeerConf{
				NeighborAddress: "127.0.0.1",
				PeerAs:          64544,
			},
			Transport: &gobgpapi.Transport{
				PassiveMode: true,
				RemotePort:  1791,
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
	_ = s.StartBgp(context.Background(), &gobgpapi.StartBgpRequest{
		Global: &gobgpapi.Global{
			As:         1,
			RouterId:   "1.1.1.1",
			ListenPort: 10179,
		},
	})
	defer s.StopBgp(context.Background(), &gobgpapi.StopBgpRequest{})
	p1 := &gobgpapi.Peer{
		Conf: &gobgpapi.PeerConf{
			NeighborAddress: "127.0.0.1",
			PeerAs:          2,
		},
		Transport: &gobgpapi.Transport{
			PassiveMode: true,
			RemotePort:  10180,
		},
	}
	_ = s.AddPeer(context.Background(), &gobgpapi.AddPeerRequest{Peer: p1})

	nlri, _ := ptypes.MarshalAny(&gobgpapi.IPAddressPrefix{
		Prefix:    "10.20.30.40",
		PrefixLen: 32,
	})
	a1, _ := ptypes.MarshalAny(&gobgpapi.OriginAttribute{
		Origin: 0,
	})
	a2, _ := ptypes.MarshalAny(&gobgpapi.NextHopAttribute{
		NextHop: "1.1.1.1",
	})
	attrs := []*anypb.Any{a1, a2}
	s.AddPath(context.TODO(), &gobgpapi.AddPathRequest{
		Path: &gobgpapi.Path{
			Family: &gobgpapi.Family{Afi: gobgpapi.Family_AFI_IP, Safi: gobgpapi.Family_SAFI_UNICAST},
			Nlri:   nlri,
			Pattrs: attrs,
		},
	})

	// bgp2
	t := gobgp.NewBgpServer(gobgp.GrpcListenAddress(":50065"))
	go t.Serve()
	_ = t.StartBgp(context.Background(), &gobgpapi.StartBgpRequest{
		Global: &gobgpapi.Global{
			As:         2,
			RouterId:   "2.2.2.2",
			ListenPort: 10180,
		},
	})
	defer t.StopBgp(context.Background(), &gobgpapi.StopBgpRequest{})

	p2 := &gobgpapi.Peer{
		Conf: &gobgpapi.PeerConf{
			NeighborAddress: "127.0.0.1",
			PeerAs:          1,
		},
		Transport: &gobgpapi.Transport{
			RemotePort: 10179,
		},
		Timers: &gobgpapi.Timers{
			Config: &gobgpapi.TimersConfig{
				ConnectRetry:           1,
				IdleHoldTimeAfterReset: 1,
			},
		},
	}
	ch := make(chan struct{})
	go t.MonitorPeer(context.Background(), &gobgpapi.MonitorPeerRequest{}, func(peer *gobgpapi.Peer) {
		if peer.State.SessionState == gobgpapi.PeerState_ESTABLISHED {
			klog.Info(peer.String())
			//close(ch)
		}
	})

	injectRoute := func(path *gobgpapi.Path) {
		dst, nextHop, err := parseBGPPath(path)
		if err != nil {
			klog.Error(err)
			return
		}

		klog.Infof(fmt.Sprintf("dst:%s, nextHop:%s", dst.String(), nextHop.String()))
	}
	// gobgp -p 50065 global rib add -a ipv4 100.0.0.0/24 nexthop 20.20.20.20
	// gobgp -p 50064 global rib
	// gobgp -p 50065 global rib summary
	go t.MonitorTable(context.TODO(), &gobgpapi.MonitorTableRequest{
		TableType: gobgpapi.TableType_GLOBAL,
		Family: &gobgpapi.Family{
			Afi:  gobgpapi.Family_AFI_IP,
			Safi: gobgpapi.Family_SAFI_UNICAST,
		},
	}, injectRoute)

	_ = t.AddPeer(context.Background(), &gobgpapi.AddPeerRequest{Peer: p2})

	<-ch
}

// parseBGPPath takes in a GoBGP Path and parses out the destination subnet and the next hop from its attributes.
// If successful, it will return the destination of the BGP path as a subnet form and the next hop. If it
// can't parse the destination or the next hop IP, it returns an error.
func parseBGPPath(path *gobgpapi.Path) (*net.IPNet, net.IP, error) {
	nextHop, err := parseBGPNextHop(path)
	if err != nil {
		return nil, nil, err
	}

	nlri := path.GetNlri()
	var prefix gobgpapi.IPAddressPrefix
	err = ptypes.UnmarshalAny(nlri, &prefix)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid nlri in advertised path")
	}
	dstSubnet, err := netlink.ParseIPNet(prefix.Prefix + "/" + fmt.Sprint(prefix.PrefixLen))
	if err != nil {
		return nil, nil, fmt.Errorf("couldn't parse IP subnet from nlri advertised path")
	}
	return dstSubnet, nextHop, nil
}

// parseBGPNextHop takes in a GoBGP Path and parses out the destination's next hop from its attributes. If it
// can't parse a next hop IP from the GoBGP Path, it returns an error.
func parseBGPNextHop(path *gobgpapi.Path) (net.IP, error) {
	for _, pAttr := range path.GetPattrs() {
		var value ptypes.DynamicAny
		if err := ptypes.UnmarshalAny(pAttr, &value); err != nil {
			return nil, fmt.Errorf("failed to unmarshal path attribute: %s", err)
		}
		// nolint:gocritic // We can't change this to an if condition because it is a .(type) expression
		switch a := value.Message.(type) {
		case *gobgpapi.NextHopAttribute:
			nextHop := net.ParseIP(a.NextHop).To4()
			if nextHop == nil {
				if nextHop = net.ParseIP(a.NextHop).To16(); nextHop == nil {
					return nil, fmt.Errorf("invalid nextHop address: %s", a.NextHop)
				}
			}
			return nextHop, nil
		}
	}
	return nil, fmt.Errorf("could not parse next hop received from GoBGP for path: %s", path)
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
	_ = s.StartBgp(context.Background(), &gobgpapi.StartBgpRequest{
		Global: &gobgpapi.Global{
			As:         65001,     // AS Number, 公司内需要调用 NetOPS API 会给本机和交换机 AS Number
			RouterId:   "2.2.2.2", // 一般选择当前机器 IP
			ListenPort: 1791,
		},
	})
	defer s.StopBgp(context.Background(), &gobgpapi.StopBgpRequest{})

	// 把 route-server(即交换机那一端) 加入到本地 bgp-server 中来
	p1 := &gobgpapi.Peer{
		Conf: &gobgpapi.PeerConf{
			NeighborAddress: "127.0.0.1",
			PeerAs:          64512,
		},
		Transport: &gobgpapi.Transport{
			RemotePort: 1790,
		},
	}
	_ = s.AddPeer(context.Background(), &gobgpapi.AddPeerRequest{Peer: p1})

	nlri, _ := ptypes.MarshalAny(&gobgpapi.IPAddressPrefix{
		Prefix:    "10.20.30.0",
		PrefixLen: 24,
	})
	a1, _ := ptypes.MarshalAny(&gobgpapi.OriginAttribute{
		Origin: 0,
	})
	a2, _ := ptypes.MarshalAny(&gobgpapi.NextHopAttribute{
		NextHop: "30.30.30.30",
	})
	attrs := []*anypb.Any{a1, a2}
	s.AddPath(context.TODO(), &gobgpapi.AddPathRequest{
		Path: &gobgpapi.Path{
			Family: &gobgpapi.Family{Afi: gobgpapi.Family_AFI_IP, Safi: gobgpapi.Family_SAFI_UNICAST},
			Nlri:   nlri,
			Pattrs: attrs,
		},
	})

	<-ch
}
