package server

import (
	"context"
	"fmt"
	"k8s-lx1036/k8s/network/bgp/pkg/config"
	"k8s-lx1036/k8s/network/bgp/pkg/table"
	"net"

	"github.com/google/uuid"
	api "github.com/osrg/gobgp/api"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type options struct {
	grpcAddress string
	grpcOption  []grpc.ServerOption
}

type ServerOption func(*options)

func GrpcListenAddress(addr string) ServerOption {
	return func(o *options) {
		o.grpcAddress = addr
	}
}

func GrpcOption(opt []grpc.ServerOption) ServerOption {
	return func(o *options) {
		o.grpcOption = opt
	}
}

type mgmtOp struct {
	f           func() error
	errCh       chan error
	checkActive bool // check BGP global setting is configured before calling f()
}

type BgpServer struct {
	bgpConfig config.Bgp

	listeners []*tcpListener
	acceptCh  chan *net.TCPConn

	neighborMap  map[string]*peer
	peerGroupMap map[string]*peerGroup

	globalRib *table.TableManager // rib: route information based
	rsRib     *table.TableManager

	// bgp monitor protocol
	bmpManager *bmpClientManager

	policy *table.RoutingPolicy

	mgmtCh chan *mgmtOp
}

func NewBgpServer(opt ...ServerOption) *BgpServer {
	opts := options{}
	for _, o := range opt {
		o(&opts)
	}

	server := &BgpServer{
		listeners: make([]*tcpListener, 0, 2),
		acceptCh:  make(chan *net.TCPConn, 4096),

		neighborMap: make(map[string]*peer),
		//peerGroupMap: make(map[string]*peerGroup),
		policy:     table.NewRoutingPolicy(),
		mgmtCh:     make(chan *mgmtOp, 1),
		watcherMap: make(map[watchEventType][]*watcher),
		uuidMap:    make(map[string]uuid.UUID),
		roaManager: newROAManager(roaTable),
		roaTable:   roaTable,
	}
	//server.bmpManager = newBmpClientManager(server)
	//server.mrtManager = newMrtManager(server)
	if len(opts.grpcAddress) != 0 {
		grpc.EnableTracing = false
		grpcServer := newAPIserver(server, grpc.NewServer(opts.grpcOption...), opts.grpcAddress)
		go func() {
			if err := grpcServer.serve(); err != nil {
				log.Fatalf("failed to listen grpc port: %s", err)
			}
		}()
	}

	return server
}

func (server *BgpServer) Serve() {

	for {
		select {
		case op := <-server.mgmtCh:
			server.handleMGMTOp(op)

		}
	}

}

// StartBgp INFO: 读取 conf 配置
func (server *BgpServer) StartBgp(ctx context.Context, r *api.StartBgpRequest) error {
	if r == nil || r.Global == nil {
		return fmt.Errorf("nil request")
	}

	return server.mgmtOperation(func() error {
		g := r.Global
		if net.ParseIP(g.RouterId) == nil {
			return fmt.Errorf("invalid router-id format: %s", g.RouterId)
		}

		c := newGlobalFromAPIStruct(g)
		if err := config.SetDefaultGlobalConfigValues(c); err != nil {
			return err
		}

		if c.Config.Port > 0 {
			acceptCh := make(chan *net.TCPConn, 4096)
			for _, addr := range c.Config.LocalAddressList {
				l, err := newTCPListener(addr, uint32(c.Config.Port), g.BindToDevice, acceptCh)
				if err != nil {
					return err
				}
				server.listeners = append(server.listeners, l)
			}
			server.acceptCh = acceptCh
		}

		rfs, _ := config.AfiSafis(c.AfiSafis).ToRfList()
		server.globalRib = table.NewTableManager(rfs)
		server.rsRib = table.NewTableManager(rfs)

		if err := server.policy.Initialize(); err != nil {
			return err
		}

		server.bgpConfig.Global = *c

		return nil
	}, false)

}

func (server *BgpServer) mgmtOperation(f func() error, checkActive bool) (err error) {
	ch := make(chan error)
	defer func() { err = <-ch }()
	server.mgmtCh <- &mgmtOp{
		f:           f,
		errCh:       ch,
		checkActive: checkActive,
	}

	return
}

func (server *BgpServer) handleMGMTOp(op *mgmtOp) {
	if op.checkActive {
		if err := server.active(); err != nil {
			op.errCh <- err
			return
		}
	}

	op.errCh <- op.f()
}

func (server *BgpServer) active() error {
	if server.bgpConfig.Global.Config.As == 0 {
		return fmt.Errorf("bgp server hasn't started yet")
	}

	return nil
}
