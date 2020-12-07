package main

import (
	"k8s-lx1036/k8s/operator/configmap-secret-controller/k8s-trigger-controller/pkg/cmd"
	"os"
)

// go run . --debug=true
// http://localhost:8001/metrics
func main() {
	if err := cmd.NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
