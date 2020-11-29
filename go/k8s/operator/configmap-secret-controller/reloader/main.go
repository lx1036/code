package main

import (
	"k8s-lx1036/k8s/operator/configmap-secret-controller/reloader/pkg/cmd"
	"os"
)

// https://github.com/stakater/Reloader
func main() {
	if err := cmd.NewReloaderCommand().Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
