package main

import (
	"flag"
	"k8s-lx1036/k8s/monitor/hpa/cmd/app"
	"os"

	"k8s.io/klog/v2"
)

func main() {
	klog.InitFlags(nil)
	flag.Set("logtostderr", "true")

	command := app.NewHPACommand()
	flag.Parse()

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}

}
