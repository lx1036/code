package main

import (
	"k8s-lx1036/k8s/operator/configmap-secret-controller/configmap-controller/pkg/cmd"
	"os"
)

// https://github.com/fabric8io/configmapcontroller
func main() {
	if err := cmd.NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
