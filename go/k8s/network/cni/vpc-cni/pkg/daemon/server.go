package daemon

import (
	"fmt"
	"io/ioutil"
	"k8s.io/klog/v2"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"k8s-lx1036/k8s/network/cni/vpc-cni/pkg/rpc"

	"google.golang.org/grpc"
)

func Run(pidFilePath, socketFilePath, daemonMode, configFilePath, kubeconfig string) error {
	var err error
	// write pid file
	if len(pidFilePath) != 0 {
		pidFilePath, err = filepath.Abs(pidFilePath)
		if err != nil {
			return err
		}

		if _, err := os.Stat(filepath.Dir(pidFilePath)); err != nil && os.IsNotExist(err) {
			if err = os.MkdirAll(filepath.Dir(pidFilePath), 0666); err != nil {
				return fmt.Errorf("create pid file %s err:%v", pidFilePath, err)
			}
		}

		if err := ioutil.WriteFile(pidFilePath, []byte(fmt.Sprintf("%d", os.Getpid())), 0666); err != nil {
			return fmt.Errorf("write pid file %s err:%v", pidFilePath, err)
		}
	}

	socketFilePath, err = filepath.Abs(socketFilePath)
	if err != nil {
		return err
	}
	listener, err := net.Listen("unix", socketFilePath)
	if err != nil {
		return fmt.Errorf("error listen at %s: %v", socketFilePath, err)
	}

	eniBackendServer, err := newEniBackendServer(daemonMode, configFilePath, kubeconfig)
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
