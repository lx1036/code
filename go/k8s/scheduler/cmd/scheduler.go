package main

import (
	"flag"
	"os"

	"k8s-lx1036/k8s/scheduler/cmd/app"

	"k8s.io/klog/v2"
)

// debug in local: `make dev`
func main() {
	klog.InitFlags(nil)
	flag.Set("logtostderr", "true")

	command := app.NewSchedulerCommand()
	flag.Parse()

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}

}
