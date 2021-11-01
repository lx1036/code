package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"k8s-lx1036/k8s/network/bgp/pkg/server"

	"github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"k8s.io/klog/v2"
)

func main() {
	var opts struct {
		ConfigFile      string `short:"f" long:"config-file" description:"specifying a config file"`
		ConfigType      string `short:"t" long:"config-type" description:"specifying config type (toml, yaml, json)" default:"toml"`
		LogLevel        string `short:"l" long:"log-level" description:"specifying log level"`
		LogPlain        bool   `short:"p" long:"log-plain" description:"use plain format for logging (json by default)"`
		UseSyslog       string `short:"s" long:"syslog" description:"use syslogd"`
		Facility        string `long:"syslog-facility" description:"specify syslog facility"`
		DisableStdlog   bool   `long:"disable-stdlog" description:"disable standard logging"`
		CPUs            int    `long:"cpus" description:"specify the number of CPUs to be used"`
		GrpcHosts       string `long:"api-hosts" description:"specify the hosts that gobgpd listens on" default:":50051"`
		GracefulRestart bool   `short:"r" long:"graceful-restart" description:"flag restart-state in graceful-restart capability"`
		Dry             bool   `short:"d" long:"dry-run" description:"check configuration"`
		PProfHost       string `long:"pprof-host" description:"specify the host that gobgpd listens on for pprof" default:"localhost:6060"`
		PProfDisable    bool   `long:"pprof-disable" description:"disable pprof profiling"`
		UseSdNotify     bool   `long:"sdnotify" description:"use sd_notify protocol"`
		TLS             bool   `long:"tls" description:"enable TLS authentication for gRPC API"`
		TLSCertFile     string `long:"tls-cert-file" description:"The TLS cert file"`
		TLSKeyFile      string `long:"tls-key-file" description:"The TLS key file"`
		Version         bool   `long:"version" description:"show version number"`
	}
	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}

	if opts.CPUs == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	} else {
		if runtime.NumCPU() < opts.CPUs {
			log.Errorf("Only %d CPUs are available but %d is specified", runtime.NumCPU(), opts.CPUs)
			os.Exit(1)
		}
		runtime.GOMAXPROCS(opts.CPUs)
	}

	switch opts.LogLevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	maxSize := 256 << 20 // 256MB
	grpcOpts := []grpc.ServerOption{grpc.MaxRecvMsgSize(maxSize), grpc.MaxSendMsgSize(maxSize)}

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

	_, err = StartBgp(context.Background(), bgpServer, initialConfig, opts.GracefulRestart)
	if err != nil {
		log.WithFields(log.Fields{
			"Topic": "Config",
			"Error": err,
		}).Fatalf("Failed to apply initial configuration %s", opts.ConfigFile)
	}

	// TODO: 动态修改 conf 文件，后续可以加上这个 feature

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
