package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"k8s-lx1036/k8s/storage/sunfs/pkg/server"
	"k8s-lx1036/k8s/storage/sunfs/pkg/server/master"
	"k8s-lx1036/k8s/storage/sunfs/pkg/server/metadata"
	"k8s-lx1036/k8s/storage/sunfs/pkg/util/config"

	"k8s.io/klog/v2"
)

var (
	configFile       = flag.String("c", "", "config file path")
	configVersion    = flag.Bool("v", false, "show version")
	configForeground = flag.Bool("f", false, "run foreground")
)

const (
	ConfigKeyRole       = "role"
	ConfigKeyLogDir     = "logDir"
	ConfigKeyLogLevel   = "logLevel"
	ConfigKeyProfPort   = "prof"
	ConfigKeyWarnLogDir = "warnLogDir"
)

const (
	RoleMaster = "master"
	RoleMeta   = "metanode"
)

const (
	ModuleMaster = "master"
	ModuleMeta   = "metaNode"
)

func main() {
	klog.InitFlags(nil)
	flag.Set("logtostderr", "true")
	flag.Parse()

	/*
	 * LoadConfigFile should be checked before start daemon, since it will
	 * call os.Exit() w/o notifying the parent process.
	 */
	cfg, err := config.LoadConfigFile(*configFile)
	if err != nil {
		klog.Error(err)
		os.Exit(1)
	}

	/*
	 * We are in daemon from here.
	 * Must notify the parent process through SignalOutcome anyway.
	 */
	role := cfg.GetString(ConfigKeyRole)
	//logDir := cfg.GetString(ConfigKeyLogDir)
	//logLevel := cfg.GetString(ConfigKeyLogLevel)
	//profPort := cfg.GetString(ConfigKeyProfPort)
	//umpDatadir := cfg.GetString(ConfigKeyWarnLogDir)

	// Init server instance with specified role configuration.
	var (
		srv server.Server
	)
	switch role {
	case RoleMeta:
		srv = metadata.NewServer()
	case RoleMaster:
		srv = master.NewServer()
	default:
		klog.Errorf("Fatal: role mismatch: %v", role)
		os.Exit(1)
	}

	interceptSignal(srv)
	err = srv.Start(cfg)
	if err != nil {
		klog.Error(err)
		os.Exit(1)
	}

	// Block main goroutine until server shutdown.
	srv.Sync()
	os.Exit(0)
}

func interceptSignal(s server.Server) {
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM)
	klog.Infof("action[interceptSignal] register system signal.")
	go func() {
		sig := <-sigC
		klog.Infof("action[interceptSignal] received signal: %s.", sig.String())
		s.Shutdown()
	}()
}
