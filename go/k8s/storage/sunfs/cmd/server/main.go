package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"k8s-lx1036/k8s/storage/sunfs/cmd/server/master"
	"k8s-lx1036/k8s/storage/sunfs/cmd/server/meta"
	"k8s-lx1036/k8s/storage/sunfs/pkg/config"

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

type Server interface {
	Start(cfg *config.Config) error
	Shutdown()
	// Sync will block invoker goroutine until this MetaNode shutdown.
	Sync()
}

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

	role := cfg.GetString(ConfigKeyRole)
	var (
		server Server
	)
	switch role {
	case RoleMeta:
		server = meta.NewServer()
	case RoleMaster:
		server = master.NewServer()
	default:
		klog.Errorf("Fatal: role mismatch: %v", role)
		os.Exit(1)
	}

	interceptSignal(server)
	err = server.Start(cfg)
	if err != nil {
		klog.Error(err)
		os.Exit(1)
	}

	// Block main goroutine until server shutdown.
	server.Sync()
	os.Exit(0)
}

func interceptSignal(server Server) {
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM)
	klog.Infof("action[interceptSignal] register system signal.")
	go func() {
		sig := <-sigC
		klog.Infof("action[interceptSignal] received signal: %s.", sig.String())
		server.Shutdown()
	}()
}
