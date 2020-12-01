package main

import (
	"k8s-lx1036/k8s/storage/kafka/kubetool/pkg/cmd"
	"os"
)

// go run . --configfile=./config.toml
func main() {
	if err := cmd.NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
