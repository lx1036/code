package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"k8s-lx1036/k8s/network/bgp/pkg/server"

	log "github.com/sirupsen/logrus"
	"k8s.io/klog/v2"
)

func main() {

	bgpServer := server.NewBgpServer(server.GrpcListenAddress(opts.GrpcHosts), server.GrpcOption(grpcOpts)) // localhost:50051
	go bgpServer.Serve()

	initialConfig, err := ReadConfigfile(opts.ConfigFile, opts.ConfigType)
	if err != nil {
		log.WithFields(log.Fields{
			"Topic": "Config",
			"Error": err,
		}).Fatalf("Can't read config file %s", opts.ConfigFile)
	}
	log.WithFields(log.Fields{
		"Topic": "Config",
	}).Info("Finished reading the config file")

	currentConfig, err := InitialConfig(context.Background(), bgpServer, initialConfig, opts.GracefulRestart)
	if err != nil {
		log.WithFields(log.Fields{
			"Topic": "Config",
			"Error": err,
		}).Fatalf("Failed to apply initial configuration %s", opts.ConfigFile)
	}

	stop := make(chan struct{})

	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigs
		klog.Infof(fmt.Sprintf("[Run]got system signal: %s, exiting", sig.String()))
		stop <- struct{}{}
	}()

	<-stop
}
