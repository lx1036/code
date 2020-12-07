package main

import (
	"k8s-lx1036/k8s/storage/log/filebeat/daemonset-operator/pkg/cmd"
	"os"
)

// go run . --debug=true --node=docker4401
// http://localhost:8001/metrics
func main() {
	if err := cmd.NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
