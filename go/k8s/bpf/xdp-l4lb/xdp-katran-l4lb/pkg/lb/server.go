package lb

import (
	"context"
	"fmt"
	"net"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-katran-l4lb/pkg/rpc"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func Run(ctx context.Context, addr string) error {

	stop := make(chan struct{})

	openLbService, err := newOpenLbService(ctx, configFilePath, daemonMode)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	rpc.RegisterOpenLbServiceServer(grpcServer, openLbService)

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("error listen at %s: %v", addr, err)
	}

	go func() {
		err = grpcServer.Serve(l)
		if err != nil {
			logrus.Errorf("error start grpc server: %v", err)
			close(stop)
		}
	}()

	select {
	case <-ctx.Done():
	case <-stop:
	}
	grpcServer.Stop()
	return nil
}

func newOpenLbService() (rpc.OpenLbServiceServer, error) {
	service := &OpenLb{}
	return service, nil
}
