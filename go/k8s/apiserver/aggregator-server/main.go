package main

import (
	"flag"
	"os"

	"k8s-lx1036/k8s/apiserver/aggregator-server/pkg/cmd/server"

	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/component-base/logs"
	"k8s.io/klog/v2"
)

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	stopCh := genericapiserver.SetupSignalHandler()
	options := server.NewDefaultOptions(os.Stdout, os.Stderr)
	cmd := server.NewCommandStartAggregator(options, stopCh)
	cmd.Flags().AddGoFlagSet(flag.CommandLine)
	if err := cmd.Execute(); err != nil {
		klog.Fatal(err)
	}
}
