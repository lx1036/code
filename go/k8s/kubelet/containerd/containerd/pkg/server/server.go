package server

import (
	"context"
	"fmt"
	"net"

	"k8s-lx1036/k8s/kubelet/containerd/containerd/pkg/plugin"
	"k8s-lx1036/k8s/kubelet/containerd/containerd/pkg/server/config"
	"k8s-lx1036/k8s/kubelet/containerd/containerd/pkg/util"

	"google.golang.org/grpc"
	"k8s.io/klog/v2"
)

// Server is the containerd main daemon
type Server struct {
	localAddr  string
	grpcServer *grpc.Server

	shutdownCh chan struct{}
}

// grpcService allows GRPC services to be registered with the underlying server
type grpcService interface {
	Register(*grpc.Server) error
}

func New(ctx context.Context, config *config.Config) (*Server, error) {
	server := &Server{
		localAddr:  config.GRPC.Address,
		grpcServer: grpc.NewServer(),

		shutdownCh: make(chan struct{}, 1),
	}

	var (
		serverOpts = []grpc.ServerOption{
			grpc.MaxRecvMsgSize(util.DefaultMaxRecvMsgSize),
			grpc.MaxSendMsgSize(util.DefaultMaxSendMsgSize),
		}
		grpcServices []grpcService
		grpcServer   = grpc.NewServer(serverOpts...)
	)
	plugins := plugin.Graph()
	initialized := plugin.NewPluginSet()
	for _, p := range plugins {
		initContext := plugin.NewContext(
			ctx,
			p,
			initialized,
			config.Root,
			config.State,
		)
		initContext.Address = config.GRPC.Address
		// load the plugin specific configuration if it is provided
		/*if p.Config != nil {
			pc, err := config.Decode(p)
			if err != nil {
				return nil, err
			}
			initContext.Config = pc
		}*/
		result := p.Init(initContext)
		if err := initialized.Add(result); err != nil {
			return nil, fmt.Errorf("could not add plugin result to plugin set: %w", err)
		}
		instance, err := result.Instance()
		if err != nil {
			return nil, err
		}

		// check for grpc services that should be registered with the server
		if src, ok := instance.(grpcService); ok {
			grpcServices = append(grpcServices, src)
		}
	}

	// register services after all plugins have been initialized
	for _, service := range grpcServices {
		if err := service.Register(grpcServer); err != nil {
			return nil, err
		}
	}

	return server, nil
}

func (server *Server) Serve() error {
	listener, err := net.Listen("unix", server.localAddr)
	if err != nil {
		return err
	}

	go func() {
		defer server.grpcServer.Stop()
		if err = server.grpcServer.Serve(listener); err != nil {
			klog.Errorf(fmt.Sprintf("[Run]serve grpcServer err: %v", err))
			server.shutdownCh <- struct{}{}
		}
	}()

	return nil
}

func (server *Server) Shutdown() <-chan struct{} {
	return server.shutdownCh
}
