package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"

	"k8s-lx1036/k8s/network/cilium/cilium/cmd/daemon/app"
	linuxdatapath "k8s-lx1036/k8s/network/cilium/cilium/pkg/datapath/linux"
	nodeTypes "k8s-lx1036/k8s/network/cilium/cilium/pkg/k8s/node/types"

	"github.com/cilium/cilium/pkg/option"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	genericapiserver "k8s.io/apiserver/pkg/server"
)

var (
	RootCmd = &cobra.Command{
		Use:   "cilium-agent",
		Short: "Run the cilium agent",
		Run: func(cmd *cobra.Command, args []string) {
			cmdRefDir := viper.GetString(option.CMDRef)
			if cmdRefDir != "" {
				os.Exit(0)
			}
			initEnv(cmd)
			runDaemon()
		},
	}
)

func main() {
	stopCh := genericapiserver.SetupSignalHandler()
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	<-stopCh
}

func runDaemon() {

	d, restoredEndpoints, err := app.NewDaemon(linuxdatapath.NewDatapath(datapathConfig, iptablesManager))

	// wait for cache sync
	<-d.k8sCachesSynced

	restoreComplete := d.initRestore(restoredEndpoints)
	go func() {
		if restoreComplete != nil {
			<-restoreComplete
		}

	}()

	srv := server.NewServer(d.instantiateAPI())
	srv.EnabledListeners = []string{"unix"}
	srv.SocketPath = option.Config.SocketPath
	srv.ReadTimeout = apiTimeout
	srv.WriteTimeout = apiTimeout
	srv.ConfigureAPI()
	defer srv.Shutdown()

	errs := make(chan error, 1)

	go func() {
		errs <- srv.Serve()
	}()

	k8s.Client().MarkNodeReady(nodeTypes.GetName())

	metricsErrs := app.InitMetrics()

	select {
	case err := <-metricsErrs:
		if err != nil {
			log.WithError(err).Fatal("Cannot start metrics server")
		}
	case err := <-errs:
		if err != nil {
			log.WithError(err).Fatal("Error returned from non-returning Serve() call")
		}
	}
}
