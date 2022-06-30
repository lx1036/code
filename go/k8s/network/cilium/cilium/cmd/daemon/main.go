package main

import (
	"fmt"
	"os"

	"k8s-lx1036/k8s/network/cilium/cilium/cmd/daemon/app"

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

	d, restoredEndpoints, err := app.NewDaemon()

	// wait for cache sync
	<-d.k8sCachesSynced

	restoreComplete := d.initRestore(restoredEndpoints)

}
