package main

import (
	"k8s-lx1036/k8s/concepts/components/controller-manager/calico-node-controller/pkg/cmd"
	"os"
)

// go run . --debug=true
// http://localhost:8001/metrics
func main() {
	if err := cmd.NewNodeControllerCommand().Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
