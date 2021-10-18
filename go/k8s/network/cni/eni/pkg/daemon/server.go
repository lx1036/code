package daemon

import (
	"fmt"
	"google.golang.org/grpc"
	"k8s-lx1036/k8s/network/cni/eni/rpc"
	"k8s.io/klog/v2"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func Run() error {

	listener, err := net.Listen("unix", socketFilePath)
	if err != nil {
		return fmt.Errorf("error listen at %s: %v", socketFilePath, err)
	}

	eniBackendServer, err := newEniBackendServer()
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()
	rpc.RegisterEniBackendServer(grpcServer, eniBackendServer)

	stop := make(chan struct{})

	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigs
		klog.Infof(fmt.Sprintf("[Run]got system signal: %s, exiting", sig.String()))
		stop <- struct{}{}
	}()

	go func() {
		err := grpcServer.Serve(listener)
		if err != nil {
			klog.Errorf(fmt.Sprintf("[Run]serve grpcServer err: %v", err))
			stop <- struct{}{}
		}
	}()

	<-stop
	return nil
}
