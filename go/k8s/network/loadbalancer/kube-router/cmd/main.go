package main

import (
	"flag"
	"os"
	"runtime"

	"k8s-lx1036/k8s/network/loadbalancer/kube-router/cmd/app"

	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/component-base/logs"
)

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	if len(os.Getenv("GOMAXPROCS")) == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	cmd := app.NewKubeRouterCommand(genericapiserver.SetupSignalHandler())
	cmd.Flags().AddGoFlagSet(flag.CommandLine)
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
