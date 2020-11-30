package main

import (
	"k8s-lx1036/k8s/operator/configmap-secret-controller/k8s-trigger-controller/pkg/cmd"
	"os"
)

// https://github.com/mfojtik/k8s-trigger-controller
func main() {
	if err := cmd.NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
