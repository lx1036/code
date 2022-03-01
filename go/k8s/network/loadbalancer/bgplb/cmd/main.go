package main

import (
	"k8s-lx1036/k8s/network/loadbalancer/bgplb/cmd/app"
	"k8s-lx1036/k8s/network/loadbalancer/bgplb/pkg/apis/bgplb/v1"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"os"
	"runtime"
	
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/component-base/logs"
)

func init() {
	_ = v1.AddToScheme(scheme.Scheme)
}

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()
	
	if len(os.Getenv("GOMAXPROCS")) == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}
	
	cmd := app.NewBGPLBCommand(genericapiserver.SetupSignalHandler())
	cmd.Flags().AddGoFlagSet(flag.CommandLine)
	if err := cmd.Execute(); err != nil {
		panic(err)
	}

}
